package convert

import (
	"fmt"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/constraints"
)

var (
	_ NodeConverter     = (*uniqueConstraintConverter)(nil)
	_ EdgeConverter     = (*uniqueConstraintConverter)(nil)
	_ PropertyConverter = (*uniqueConstraintConverter)(nil)
)

type uniqueConstraintConverter struct{}

func (t *uniqueConstraintConverter) IsCompatible(_ *DatabaseMetadata) (bool, error) {
	return true, nil
}

func (t *uniqueConstraintConverter) ConvertNode(nodeType *gql.NodeType) ([]neo4j.SchemaObject, error) {
	if nodeType.ExtendedAttributes == nil || len(nodeType.ExtendedAttributes.CompositeUniqueConstraintFields) == 0 {
		return nil, nil
	}

	label, err := getPrimaryLabel(nodeType.NodeTypeLabelSet)
	if err != nil {
		return nil, fmt.Errorf("error getting label for node type %s: %w", nodeType.NodeTypeLabelSet, err)
	}

	return newUniqueConstraint(NodeEntityType, label, nodeType.ExtendedAttributes.CompositeUniqueConstraintFields)
}

func (t *uniqueConstraintConverter) ConvertEdge(edgeType *gql.EdgeType) ([]neo4j.SchemaObject, error) {
	if edgeType.ExtendedAttributes == nil || len(edgeType.ExtendedAttributes.CompositeUniqueConstraintFields) == 0 {
		return nil, nil
	}

	label, err := getPrimaryLabel(edgeType.EdgeTypeLabelSet)
	if err != nil {
		return nil, fmt.Errorf("error getting label for edge type %s: %w", edgeType.EdgeTypeLabelSet, err)
	}

	return newUniqueConstraint(EdgeEntityType, label, edgeType.ExtendedAttributes.CompositeUniqueConstraintFields)
}

func (t *uniqueConstraintConverter) ConvertProperty(entityType EntityType, entityLabel string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.ExtendedAttributes == nil || !propertyType.ExtendedAttributes.IsUnique {
		return nil, nil
	}

	return newUniqueConstraint(entityType, entityLabel, []string{propertyType.PropertyName})
}

func newUniqueConstraint(entityType EntityType, label string, fields []string) ([]neo4j.SchemaObject, error) {
	return []neo4j.SchemaObject{
		&constraints.PropertyUniquenessConstraint{
			Name:          mustFormatName("unique", append([]string{label}, fields...)),
			Pattern:       getPattern(entityType, "n", label),
			Alias:         "n",
			PropertyNames: fields,
		},
	}, nil
}
