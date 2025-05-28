package gql2neo4j_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/stretchr/testify/assert"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/gql/datatype"
	"github.com/gql-schema/gqls/internal/gql2neo4j"
	"github.com/gql-schema/gqls/internal/gql2neo4j/convert"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/constraints"
	"github.com/gql-schema/gqls/internal/neo4j/triggers"
)

type testCase struct {
	Input string `yaml:"input"`
	Want  string `yaml:"want"`
}

func TestConvertString(t *testing.T) {
	testCasePaths, err := filepath.Glob("testdata/*.yml")
	if err != nil {
		t.Fatalf("failed to find test cases: %v", err)
	}

	for _, path := range testCasePaths {
		name := filepath.Base(path)
		t.Run(name, func(t *testing.T) {
			tc := loadTestCase(t, path)

			got, err := gql2neo4j.ConvertString(tc.Input, convert.DatabaseMetadata{APOCEnabled: true})
			if err != nil {
				t.Fatalf("ConvertString() returned error: %v", err)
			}

			want := strings.TrimSpace(tc.Want)
			got = strings.TrimSpace(got)
			assert.Equal(t, want, got)
		})
	}
}

func TestConvert(t *testing.T) {
	type args struct {
		graphType *gql.GraphType
		opts      convert.DatabaseMetadata
	}
	tests := []struct {
		name    string
		args    args
		want    []neo4j.SchemaObject
		wantErr bool
	}{
		{
			name: "empty graph type",
			args: args{
				graphType: &gql.GraphType{},
			},
			want: nil,
		},
		{
			name: "node type constraints",
			args: args{
				graphType: &gql.GraphType{
					NodeTypeSet: []*gql.NodeType{
						{
							NodeTypeLabelSet: []string{"Person"},
							PropertyTypeSet: []*gql.PropertyType{
								{
									PropertyName: "name",
									PropertyValueType: &datatype.CharacterStringDataTypeDescriptor{
										Nullable: true,
									},
								},
								{
									PropertyName: "age",
									PropertyValueType: &datatype.NumericDataTypeDescriptor{
										Kind:      datatype.NumericKindExact,
										Radix:     datatype.RadixDecimal,
										Precision: 7,
										Nullable:  false,
									},
								},
							},
						},
					},
				},
				opts: convert.DatabaseMetadata{
					APOCEnabled: true,
				},
			},
			want: []neo4j.SchemaObject{
				&constraints.PropertyTypeConstraint{
					Name:         "type_person_name",
					Pattern:      "(n:Person)",
					Alias:        "n",
					PropertyName: "name",
					PropertyType: "STRING",
				},
				&constraints.PropertyTypeConstraint{
					Name:         "type_person_age",
					Pattern:      "(n:Person)",
					Alias:        "n",
					PropertyName: "age",
					PropertyType: "INTEGER",
				},
				&constraints.PropertyExistenceConstraint{
					Name:         "existence_person_age",
					Pattern:      "(n:Person)",
					Alias:        "n",
					PropertyName: "age",
				},
				triggers.NewTrigger(
					"numeric_range_person_age",
					"UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.age IS NOT NULL AND (e.age < -128 OR e.age > 127) CALL apoc.util.validate(true, 'age must be between -128 and 127', []) RETURN NULL UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change.node:Person AND change.key = \"age\" AND change.new IS NOT NULL AND (change.new < -128 OR change.new > 127) CALL apoc.util.validate(true, 'age must be between -128 and 127', []) RETURN NULL",
					triggers.Selector{Phase: triggers.Before},
				),
			},
		},
		{
			name: "node type constraints unsupported",
			args: args{
				graphType: &gql.GraphType{
					NodeTypeSet: []*gql.NodeType{
						{
							NodeTypeLabelSet: []string{"Person"},
							PropertyTypeSet: []*gql.PropertyType{
								{
									PropertyName: "foo",
									PropertyValueType: &datatype.MiscellaneousDataTypeDescriptor{
										Kind:     "BAR",
										Nullable: true,
									},
								},
							},
						},
					},
				},
				opts: convert.DatabaseMetadata{
					Version:     "5.8",
					APOCEnabled: true,
				},
			},
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := gql2neo4j.Convert(tt.args.graphType, tt.args.opts)
			if (err != nil) != tt.wantErr {
				t.Errorf("Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

// loadTestCase loads a test case from the specified path.
func loadTestCase(t *testing.T, path string) testCase {
	t.Helper()

	data, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("failed to read test case file %q: %v", path, err)
	}

	var tc testCase
	if err := yaml.Unmarshal(data, &tc); err != nil {
		t.Fatalf("failed to unmarshal test case %q: %v", path, err)
	}

	if strings.TrimSpace(tc.Input) == "" || strings.TrimSpace(tc.Want) == "" {
		t.Skipf("test case %q missing input or want", path)
	}

	return tc
}
