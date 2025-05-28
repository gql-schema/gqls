// Package convert provides functionality to convert GQL schema objects into Neo4j schema objects.
package convert

import (
	"embed"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/neo4j"
)

//go:embed templates/*
var triggerTemplates embed.FS

// converters holds all converters.
var converters = []Converter{
	&typeConstraintConverter{},
	&uniqueConstraintConverter{},
	&existenceConstraintConverter{},
	&keyConstraintConverter{},
	&stringLengthTriggerConverter{},
	&numericRangeTriggerConverter{},
	&checkConstraintTriggerConverter{},
	&edgeEndpointsTriggerConverter{},
	&cardinalityTriggerConverter{},
}

// Converters returns a list of all converters of a specific type.
func Converters[T Converter]() []T {
	var filtered []T
	for _, t := range converters {
		if specific, ok := t.(T); ok {
			filtered = append(filtered, specific)
		}
	}
	return filtered
}

// Converter is an interface that defines a converter.
type Converter interface {
	// IsCompatible checks if the Converter is compatible with the database.
	IsCompatible(meta *DatabaseMetadata) (bool, error)
}

// NodeConverter is a Converter that converts gql.NodeType instances.
type NodeConverter interface {
	Converter

	// ConvertNode convers a node type into a list of Neo4j schema objects.
	ConvertNode(nodeType *gql.NodeType) ([]neo4j.SchemaObject, error)
}

// EdgeConverter is a Converter that converts gql.EdgeType instances
type EdgeConverter interface {
	Converter

	// ConvertEdge convers an edge type into a list of Neo4j schema objects.
	ConvertEdge(edgeType *gql.EdgeType) ([]neo4j.SchemaObject, error)
}

// PropertyConverter is a Converter that converts gql.PropertyType instances
type PropertyConverter interface {
	Converter

	// ConvertProperty convers a property type into a list of Neo4j schema objects.
	ConvertProperty(entityType EntityType, entityLabel string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error)
}
