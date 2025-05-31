// Package cli provides the command-line interface for the gqls tool.
package cli

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/urfave/cli/v3"

	"github.com/gql-schema/gqls/internal/gql2neo4j"
	"github.com/gql-schema/gqls/internal/gql2neo4j/convert"
)

// Run initializes the CLI application and starts processing commands.
func Run(ctx context.Context, version, commit string, args []string) {
	app := &cli.Command{
		Name:      "gqls",
		Usage:     "Convert a GQL schema into Neo4j format",
		ArgsUsage: "[input]",
		Version:   fmt.Sprintf("%s (%s)", version, commit),
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "input",
				Aliases: []string{"i"},
				Sources: cli.EnvVars("GQLS_INPUT"),
				Usage:   "Path to the input GQL schema file (use '-' to read from stdin)",
			},
			&cli.StringFlag{
				Name:    "output",
				Aliases: []string{"o"},
				Sources: cli.EnvVars("GQLS_OUTPUT"),
				Usage:   "Path to the output file (defaults to stdout)",
			},
			&cli.StringFlag{
				Name:        "neo4j-version",
				Sources:     cli.EnvVars("GQLS_NEO4J_VERSION"),
				Usage:       "Specify the Neo4j version to use for conversion",
				DefaultText: "latest",
			},
			&cli.StringFlag{
				Name:        "neo4j-edition",
				Sources:     cli.EnvVars("GQLS_NEO4J_EDITION"),
				Usage:       "Specify the Neo4j edition to use for conversion ('community', 'enterprise')",
				DefaultText: "enterprise",
			},
			&cli.BoolFlag{
				Name:        "apoc",
				Usage:       "Generate APOC triggers",
				DefaultText: "false",
			},
		},
		Action: convertSchema,
	}

	if err := app.Run(ctx, args); err != nil {
		log.Fatalf("application error: %v", err)
	}
}

// convertSchema is the action function for the CLI command that converts a GQL schema to Neo4j format.
func convertSchema(_ context.Context, cmd *cli.Command) error {
	inputPath := cmd.String("input")
	content, err := readInput(inputPath)
	if err != nil {
		return fmt.Errorf("failed to read input: %w", err)
	}

	var edition convert.Neo4jEdition
	switch strings.ToLower(cmd.String("neo4j-edition")) {
	case "community":
		edition = convert.Neo4jCommunityEdition
	case "enterprise":
		edition = convert.Neo4jEnterpriseEdition
	default:
		return fmt.Errorf("invalid Neo4j edition: %s, must be 'community' or 'enterprise'", cmd.String("neo4j-edition"))
	}

	converted, err := gql2neo4j.ConvertBytes(content, convert.DatabaseMetadata{
		Version:     cmd.String("neo4j-version"),
		Edition:     edition,
		APOCEnabled: cmd.Bool("apoc"),
	})
	if err != nil {
		return fmt.Errorf("failed to convert GQL to Neo4j format: %w", err)
	}

	if outputPath := cmd.String("output"); outputPath != "" {
		if err := os.WriteFile(outputPath, converted, 0o644); err != nil {
			return fmt.Errorf("failed to write output: %w", err)
		}
	} else {
		if _, err := os.Stdout.Write(converted); err != nil {
			return fmt.Errorf("failed to write to stdout: %w", err)
		}
	}

	return nil
}

// readInput reads the input from a file or stdin based on the provided path.
func readInput(path string) ([]byte, error) {
	if path == "-" || path == "" {
		return io.ReadAll(os.Stdin)
	}
	return os.ReadFile(filepath.Clean(path))
}
