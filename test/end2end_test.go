//go:build integration

package test_test

import (
	"context"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/goccy/go-yaml"
	"github.com/neo4j/neo4j-go-driver/v5/neo4j"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/gql-schema/gqls/internal/cli"
)

const (
	neo4jVersion            = "2025.04-enterprise"
	neo4jBoltPort           = "7687/tcp"
	containerStartupTimeout = 2 * time.Minute
	testSpecPattern         = "testdata/*.spec.yml"
)

var (
	sharedContainer testcontainers.Container
	sharedDriver    neo4j.DriverWithContext
)

// testCase represents a single test case within a test specification.
type testCase struct {
	Name    string `yaml:"name"`
	Setup   string `yaml:"setup"`
	Query   string `yaml:"query"`
	WantErr bool   `yaml:"want_err"`
}

// testSpec represents a complete test specification with schema and test cases.
type testSpec struct {
	Schema string     `yaml:"schema"`
	Cases  []testCase `yaml:"cases"`
}

// TestMain sets up and tears down the shared Neo4j container for all integration tests.
func TestMain(m *testing.M) {
	ctx, cancel := context.WithTimeout(context.Background(), containerStartupTimeout)
	defer cancel()

	if err := setupInfrastructure(ctx); err != nil {
		fmt.Printf("failed to setup test infrastructure: %v\n", err)
		os.Exit(1)
	}

	exitCode := m.Run()

	teardownTestInfrastructure()

	os.Exit(exitCode)
}

func TestIntegrationEnd2End(t *testing.T) {
	testSpecPaths, err := filepath.Glob(testSpecPattern)
	if err != nil {
		t.Fatalf("failed to find test cases: %v", err)
	}

	for _, testSpecPath := range testSpecPaths {
		name := filepath.Base(testSpecPath)
		t.Run(name, func(t *testing.T) {
			t.Helper()

			for i, tSpec := range loadTestSpecs(t, testSpecPath) {
				t.Run(fmt.Sprintf("[%d]", i), func(t *testing.T) {
					t.Helper()

					withNeo4jSession(t, func(session neo4j.SessionWithContext) {
						t.Helper()

						for _, tCase := range tSpec.Cases {
							t.Run(tCase.Name, func(t *testing.T) {
								cleanDatabase(t, session)
								createSchema(t, session, name, tSpec.Schema)
								executeTestCase(t, session, tCase)
							})
						}
					})
				})
			}
		})
	}
}

// createSchema applies the GQL schema to the Neo4j database.
func createSchema(t *testing.T, session neo4j.SessionWithContext, specName, schema string) {
	t.Helper()

	statements := convertGQLToNeo4j(t, schema)
	for _, statement := range splitStatements(statements) {
		t.Logf("Creating schema for spec %q: %s", specName, statement)
		if _, err := session.Run(t.Context(), statement, nil); err != nil {
			t.Fatalf("failed to create statement for spec %q: %v", specName, err)
		}
	}
}

// executeTestCase runs a single test case and validates the result.
func executeTestCase(t *testing.T, session neo4j.SessionWithContext, tCase testCase) {
	t.Helper()

	if tCase.Setup != "" {
		if err := executeStatement(t, session, tCase.Setup); err != nil {
			assert.Failf(t, "Setup failed", "failed to execute setup for test case %q: %v", tCase.Name, err)
		}
	}

	err := executeStatement(t, session, tCase.Query)
	if tCase.WantErr {
		assert.Error(t, err, "expected an error for test case %q, but got none", tCase.Name)
	} else {
		assert.NoError(t, err, "expected no error for test case %q", tCase.Name)
	}
}

func executeStatement(t *testing.T, session neo4j.SessionWithContext, statement string) error {
	t.Helper()

	result, err := session.Run(t.Context(), statement, nil)
	if err != nil {
		return fmt.Errorf("failed to execute statement %q: %w", statement, err)
	}

	// Consume the result to check for APOC trigger errors.
	if _, err := result.Consume(t.Context()); err != nil {
		return fmt.Errorf("failed to consume result for statement %q: %w", statement, err)
	}

	return nil
}

// loadTestSpecs loads test specifications from a YAML file.
func loadTestSpecs(t *testing.T, path string) []testSpec {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read test file %q: %v", path, err)
	}

	var ts []testSpec
	if err := yaml.Unmarshal(data, &ts); err != nil {
		t.Fatalf("failed to unmarshal test file %q: %v", path, err)
	}
	return ts
}

// splitStatements splits a string into individual statements based on semicolons.
func splitStatements(s string) []string {
	var statements []string
	for _, statement := range strings.Split(s, ";") {
		statement = strings.TrimSpace(statement)
		if statement != "" {
			statements = append(statements, statement)
		}
	}
	return statements
}

// convertGQLToNeo4j uses the gqls CLI to convert a GQL schema to Neo4j Cypher statements.
func convertGQLToNeo4j(t *testing.T, schema string) string {
	t.Helper()

	inputPath := path.Join(t.TempDir(), "schema.gql")
	if err := os.WriteFile(inputPath, []byte(schema), 0o644); err != nil {
		t.Fatalf("failed to write schema to %q: %v", inputPath, err)
	}

	outputPath := path.Join(t.TempDir(), "converted.cypher")

	cli.Run(t.Context(), "<version>", "<commit>", []string{
		"gqls",
		"--input", inputPath,
		"--output", outputPath,
		"--neo4j-version", "latest",
		"--neo4j-edition", "enterprise",
		"--apoc",
	})

	converted, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("failed to read converted file %q: %v", outputPath, err)
	}
	return string(converted)
}

// cleanDatabase removes all data and schema objects from the Neo4j database.
func cleanDatabase(t *testing.T, session neo4j.SessionWithContext) {
	t.Helper()

	if _, err := session.Run(t.Context(), "MATCH (n) DETACH DELETE n", nil); err != nil {
		t.Logf("failed to delete nodes and relationships: %v", err)
	}

	if err := dropAllConstraints(t, session); err != nil {
		t.Logf("failed to drop constraints: %v", err)
	}

	if err := dropAllTriggers(t, session); err != nil {
		t.Logf("failed to drop APOC triggers: %v", err)
	}
}

// dropAllConstraints removes all constraints from the database.
func dropAllConstraints(t *testing.T, session neo4j.SessionWithContext) error {
	t.Helper()

	result, err := session.Run(t.Context(), "SHOW CONSTRAINTS", nil)
	if err != nil {
		return fmt.Errorf("failed to list constraints: %w", err)
	}

	constraints, err := result.Collect(t.Context())
	if err != nil {
		return fmt.Errorf("failed to collect constraints: %w", err)
	}

	for _, constraint := range constraints {
		if name, found := constraint.AsMap()["name"]; found {
			if nameStr, ok := name.(string); ok {
				dropQuery := fmt.Sprintf("DROP CONSTRAINT %s", nameStr)
				if _, err := session.Run(t.Context(), dropQuery, nil); err != nil {
					t.Logf("failed to drop constraint %s: %v", nameStr, err)
				}
			}
		}
	}
	return nil
}

// dropAllTriggers removes all APOC triggers from the database.
func dropAllTriggers(t *testing.T, session neo4j.SessionWithContext) error {
	t.Helper()

	_, err := session.Run(t.Context(), "CALL apoc.trigger.removeAll()", nil)
	if err != nil {
		return fmt.Errorf("failed to remove APOC triggers: %w", err)
	}
	return nil
}

func withNeo4jSession(t *testing.T, testFunc func(session neo4j.SessionWithContext)) {
	t.Helper()

	if sharedDriver == nil {
		t.Fatal("shared Neo4j driver not initialized")
	}

	session := sharedDriver.NewSession(t.Context(), neo4j.SessionConfig{})
	defer func() {
		if closeErr := session.Close(t.Context()); closeErr != nil {
			t.Errorf("failed to close Neo4j session: %v", closeErr)
		}
	}()

	testFunc(session)
}

// setupInfrastructure initializes the shared Neo4j container and driver.
func setupInfrastructure(ctx context.Context) error {
	container, err := startNeo4jContainer(ctx)
	if err != nil {
		return fmt.Errorf("failed to start Neo4j container: %w", err)
	}
	sharedContainer = container

	host, err := container.Host(ctx)
	if err != nil {
		return fmt.Errorf("failed to get container host: %w", err)
	}

	port, err := container.MappedPort(ctx, neo4jBoltPort)
	if err != nil {
		return fmt.Errorf("failed to get mapped port: %w", err)
	}

	driver, err := neo4j.NewDriverWithContext(
		fmt.Sprintf("bolt://%s:%d", host, port.Int()),
		neo4j.NoAuth(),
	)
	if err != nil {
		return fmt.Errorf("failed to create Neo4j driver: %w", err)
	}
	sharedDriver = driver

	if err := driver.VerifyConnectivity(ctx); err != nil {
		return fmt.Errorf("failed to verify connectivity: %w", err)
	}

	return nil
}

// teardownTestInfrastructure cleans up the shared Neo4j container and driver.
func teardownTestInfrastructure() {
	ctx := context.Background()

	if sharedDriver != nil {
		if err := sharedDriver.Close(ctx); err != nil {
			fmt.Printf("failed to close Neo4j driver: %v\n", err)
		}
	}

	if sharedContainer != nil {
		if err := sharedContainer.Terminate(ctx); err != nil {
			fmt.Printf("failed to terminate container: %v\n", err)
		}
	}
}

// startNeo4jContainer starts a Neo4j container with the specified configuration.
func startNeo4jContainer(ctx context.Context) (testcontainers.Container, error) {
	req := testcontainers.GenericContainerRequest{
		ProviderType: getProviderType(),
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        "docker.io/neo4j:" + neo4jVersion,
			ExposedPorts: []string{"7687/tcp"},
			WaitingFor:   wait.ForLog("Bolt enabled on"),
			Env: map[string]string{
				"NEO4J_AUTH":                       "none",
				"NEO4J_ACCEPT_LICENSE_AGREEMENT":   "yes",
				"NEO4J_dbms_usage__report_enabled": "false",
				"NEO4J_PLUGINS":                    "[\"apoc\"]",
				"NEO4J_apoc_trigger_enabled":       "true",
			},
			HostConfigModifier: func(hc *container.HostConfig) {
				hc.AutoRemove = true
			},
		},
		Started: true,
	}

	container, err := testcontainers.GenericContainer(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to start container: %w", err)
	}

	return container, nil
}

// getProviderType returns the provider type based on the TESTCONTAINERS_DRIVER environment variable.
func getProviderType() testcontainers.ProviderType {
	if os.Getenv("TESTCONTAINERS_DRIVER") == "podman" {
		return testcontainers.ProviderPodman
	}
	return testcontainers.ProviderDocker
}
