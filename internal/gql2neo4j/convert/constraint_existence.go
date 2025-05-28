package convert

import (
	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/constraints"
)

var _ PropertyConverter = (*existenceConstraintConverter)(nil)

type existenceConstraintConverter struct{}

func (t *existenceConstraintConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.Edition == Neo4jEnterpriseEdition, nil
}

func (t *existenceConstraintConverter) ConvertProperty(entityType EntityType, entityLabel string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.PropertyValueType.IsNullable() {
		return nil, nil
	}

	return []neo4j.SchemaObject{
		&constraints.PropertyExistenceConstraint{
			Name:         mustFormatName("existence", []string{entityLabel, propertyType.PropertyName}),
			Pattern:      getPattern(entityType, "n", entityLabel),
			Alias:        "n",
			PropertyName: propertyType.PropertyName,
		},
	}, nil
}
