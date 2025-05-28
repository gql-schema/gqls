package convert

import (
	"fmt"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/constraints"
)

var (
	_ NodeConverter     = (*keyConstraintConverter)(nil)
	_ EdgeConverter     = (*keyConstraintConverter)(nil)
	_ PropertyConverter = (*keyConstraintConverter)(nil)
)

type keyConstraintConverter struct{}

func (t *keyConstraintConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.Edition == Neo4jEnterpriseEdition, nil
}

func (t *keyConstraintConverter) ConvertNode(nodeType *gql.NodeType) ([]neo4j.SchemaObject, error) {
	if nodeType.ExtendedAttributes == nil || len(nodeType.ExtendedAttributes.CompositePrimaryKeyFields) == 0 {
		return nil, nil
	}

	label, err := getPrimaryLabel(nodeType.NodeTypeLabelSet)
	if err != nil {
		return nil, fmt.Errorf("error getting label for node type %s: %w", nodeType.NodeTypeLabelSet, err)
	}

	return newKeyConstraint(NodeEntityType, label, nodeType.ExtendedAttributes.CompositePrimaryKeyFields)
}

func (t *keyConstraintConverter) ConvertEdge(edgeType *gql.EdgeType) ([]neo4j.SchemaObject, error) {
	if edgeType.ExtendedAttributes == nil || len(edgeType.ExtendedAttributes.CompositePrimaryKeyFields) == 0 {
		return nil, nil
	}

	label, err := getPrimaryLabel(edgeType.EdgeTypeLabelSet)
	if err != nil {
		return nil, fmt.Errorf("error getting label for edge type %s: %w", edgeType.EdgeTypeLabelSet, err)
	}

	return newKeyConstraint(EdgeEntityType, label, edgeType.ExtendedAttributes.CompositePrimaryKeyFields)
}

func (t *keyConstraintConverter) ConvertProperty(entityType EntityType, entityLabel string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.ExtendedAttributes == nil || !propertyType.ExtendedAttributes.IsPrimaryKey {
		return nil, nil
	}

	return newKeyConstraint(entityType, entityLabel, []string{propertyType.PropertyName})
}

func newKeyConstraint(entityType EntityType, label string, fields []string) ([]neo4j.SchemaObject, error) {
	subjectType, ok := mapSubjectType(entityType)
	if !ok {
		return nil, fmt.Errorf("unsupported entity type %v for key constraint", entityType)
	}

	return []neo4j.SchemaObject{&constraints.KeyConstraint{
		Name:          mustFormatName("key", append([]string{label}, fields...)),
		Pattern:       getPattern(entityType, "n", label),
		Alias:         "n",
		PropertyNames: fields,
		SubjectType:   subjectType,
	}}, nil
}

func mapSubjectType(entityType EntityType) (constraints.SubjectType, bool) {
	switch entityType {
	case NodeEntityType:
		return constraints.NodeType, true
	case EdgeEntityType:
		return constraints.RelationshipType, true
	default:
		return constraints.SubjectType(0), false
	}
}
