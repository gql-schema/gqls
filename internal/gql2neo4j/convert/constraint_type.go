package convert

import (
	"fmt"
	"strings"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/gql/datatype"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/constraints"
)

var _ PropertyConverter = (*typeConstraintConverter)(nil)

type typeConstraintConverter struct{}

func (t *typeConstraintConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	editionSupported := meta.Edition == Neo4jEnterpriseEdition

	versionSupported, err := isVersionSupported(meta.Version, "5.9")
	if err != nil {
		return false, fmt.Errorf("error checking Neo4j version %q: %w", meta.Version, err)
	}

	return editionSupported && versionSupported, nil
}

func (t *typeConstraintConverter) ConvertProperty(entityType EntityType, entityLabel string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	var neo4jType neo4j.PropertyType
	switch valueType := propertyType.PropertyValueType.(type) {
	case *datatype.CharacterStringDataTypeDescriptor:
		neo4jType = neo4j.String
	case *datatype.NumericDataTypeDescriptor:
		switch valueType.Kind {
		case datatype.NumericKindExact:
			neo4jType = neo4j.Integer
		case datatype.NumericKindFloat:
			neo4jType = neo4j.Float
		default:
			return nil, fmt.Errorf("error unsupported numeric GQL type: %s", valueType.Kind)
		}
	case *datatype.ByteStringDataTypeDescriptor:
		return nil, fmt.Errorf("error unsupported GQL type: %v", valueType)
	case *datatype.MiscellaneousDataTypeDescriptor:
		x, err := mapMiscGQLTypeToNeo4j(valueType.Kind)
		if err != nil {
			return nil, fmt.Errorf("error unsupported GQL type: %v", valueType)
		}
		neo4jType = x
	default:
		return nil, fmt.Errorf("error unsupported GQL type: %s", valueType)
	}

	return []neo4j.SchemaObject{
		&constraints.PropertyTypeConstraint{
			Name:         mustFormatName("type", []string{entityLabel, propertyType.PropertyName}),
			Pattern:      getPattern(entityType, "n", entityLabel),
			Alias:        "n",
			PropertyName: propertyType.PropertyName,
			PropertyType: string(neo4jType),
		},
	}, nil
}

// mapMiscGQLTypeToNeo4j handles miscellaneous GQL types.
func mapMiscGQLTypeToNeo4j(kind string) (neo4j.PropertyType, error) {
	switch strings.ToUpper(kind) {
	case "BOOL", "BOOLEAN":
		return neo4j.Boolean, nil
	case "DATE":
		return neo4j.Date, nil
	case "ZONED DATETIME", "TIMESTAMP WITH TIME ZONE":
		return neo4j.ZonedDateTime, nil
	case "LOCAL DATETIME", "TIMESTAMP", "TIMESTAMP WITHOUT TIME ZONE":
		return neo4j.LocalDateTime, nil
	case "ZONED TIME", "TIME WITH TIME ZONE":
		return neo4j.ZonedTime, nil
	case "LOCAL TIME", "TIME WITHOUT TIME ZONE":
		return neo4j.LocalTime, nil
	case "DURATION":
		return neo4j.Duration, nil
	default:
		return "", fmt.Errorf("error unsupported GQL type: %s", kind)
	}
}
