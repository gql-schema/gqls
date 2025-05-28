// Package gql2neo4j provides functionality to convert GQL schema definitions into Neo4j constraints.
package gql2neo4j

import (
	"fmt"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/gql/parser"
	"github.com/gql-schema/gqls/internal/gql2neo4j/convert"
	"github.com/gql-schema/gqls/internal/neo4j"
)

// ConvertBytes parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertBytes(data []byte, meta convert.DatabaseMetadata) ([]byte, error) {
	result, err := ConvertString(string(data), meta)
	if err != nil {
		return nil, fmt.Errorf("error converting GQL to Neo4j: %w", err)
	}
	return []byte(result), nil
}

// ConvertString parses a GQL schema and generates Neo4j constraints for the specified version.
func ConvertString(data string, meta convert.DatabaseMetadata) (string, error) {
	parsed, err := parseGQL(data)
	if err != nil {
		return "", fmt.Errorf("error parsing GQL: %w", err)
	}

	var result strings.Builder
	for _, graphType := range parsed.CreateGraphTypeStatements {
		schemaObjects, err := Convert(graphType, meta)
		if err != nil {
			return "", fmt.Errorf("error converting graph type: %w", err)
		}
		for _, c := range schemaObjects {
			_, _ = fmt.Fprintf(&result, "%s;\n", c.CreateStatement())
		}
	}

	return result.String(), nil
}

// Convert builds Neo4j constraints from a GQL GraphType definition.
func Convert(graphType *gql.GraphType, meta convert.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	for _, nodeType := range graphType.NodeTypeSet {
		if len(nodeType.NodeTypeLabelSet) != 1 {
			return nil, fmt.Errorf("error Neo4j only supports one label per node type")
		}
		label := nodeType.NodeTypeLabelSet[0]

		nodeSchemaObjects, err := convertNode(nodeType, meta)
		if err != nil {
			return nil, fmt.Errorf("error converting node type: %w", err)
		}
		schemaObjects = append(schemaObjects, nodeSchemaObjects...)

		objects, err := processProperties(convert.NodeEntityType, label, nodeType.PropertyTypeSet, &meta)
		if err != nil {
			return nil, fmt.Errorf("error processing properties: %w", err)
		}
		schemaObjects = append(schemaObjects, objects...)
	}

	for _, edgeType := range graphType.EdgeTypeSet {
		if len(edgeType.EdgeTypeLabelSet) != 1 {
			return nil, fmt.Errorf("error Neo4j only supports one label per edge type")
		}
		label := edgeType.EdgeTypeLabelSet[0]

		edgeSchemaObjects, err := convertEdge(edgeType, meta)
		if err != nil {
			return nil, fmt.Errorf("error converting edge type: %w", err)
		}
		schemaObjects = append(schemaObjects, edgeSchemaObjects...)

		propertyObjects, err := processProperties(convert.EdgeEntityType, label, edgeType.EdgeTypePropertySet, &meta)
		if err != nil {
			return nil, fmt.Errorf("error processing properties: %w", err)
		}
		schemaObjects = append(schemaObjects, propertyObjects...)
	}

	return schemaObjects, nil
}

// processProperties processes properties for a given label and version.
func processProperties(entityType convert.EntityType, entityLabel string, properties []*gql.PropertyType, meta *convert.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject

	for _, property := range properties {
		for _, propertyConverter := range convert.Converters[convert.PropertyConverter]() {
			isCompatible, err := propertyConverter.IsCompatible(meta)
			if err != nil {
				return nil, fmt.Errorf("error checking compatibility: %w", err)
			}
			if !isCompatible {
				continue
			}

			convertedObjects, err := propertyConverter.ConvertProperty(entityType, entityLabel, property)
			if err != nil {
				return nil, fmt.Errorf("error converting property: %w", err)
			}
			schemaObjects = append(schemaObjects, convertedObjects...)
		}
	}

	return schemaObjects, nil
}

func convertNode(nodeType *gql.NodeType, meta convert.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject
	for _, nodeConverter := range convert.Converters[convert.NodeConverter]() {
		isCompatible, err := nodeConverter.IsCompatible(&meta)
		if err != nil {
			return nil, fmt.Errorf("error checking compatibility: %w", err)
		}

		if !isCompatible {
			continue
		}

		convertedObjects, err := nodeConverter.ConvertNode(nodeType)
		if err != nil {
			return nil, fmt.Errorf("error converting node type: %w", err)
		}
		schemaObjects = append(schemaObjects, convertedObjects...)
	}
	return schemaObjects, nil
}

func convertEdge(edgeType *gql.EdgeType, meta convert.DatabaseMetadata) ([]neo4j.SchemaObject, error) {
	var schemaObjects []neo4j.SchemaObject
	for _, edgeConverter := range convert.Converters[convert.EdgeConverter]() {
		isCompatible, err := edgeConverter.IsCompatible(&meta)
		if err != nil {
			return nil, fmt.Errorf("error checking compatibility: %w", err)
		}

		if !isCompatible {
			continue
		}

		convertedObjects, err := edgeConverter.ConvertEdge(edgeType)
		if err != nil {
			return nil, fmt.Errorf("error converting edge type: %w", err)
		}
		schemaObjects = append(schemaObjects, convertedObjects...)
	}
	return schemaObjects, nil
}

// parseGQL parses a GQL input string into a listener structure.
func parseGQL(input string) (*gql.CreateGraphStatementListener, error) {
	lexer := parser.NewGQLLexer(antlr.NewInputStream(input))
	stream := antlr.NewCommonTokenStream(lexer, antlr.TokenDefaultChannel)
	p := parser.NewGQLParser(stream)

	errorListener := &gql.ErrorListener{}
	p.AddErrorListener(errorListener)

	createGraphTypeListener := &gql.CreateGraphStatementListener{}
	antlr.ParseTreeWalkerDefault.Walk(createGraphTypeListener, p.GqlProgram())

	if len(errorListener.SyntaxErrors) > 0 {
		return nil, fmt.Errorf("syntax errors: %v", errorListener.SyntaxErrors)
	}

	return createGraphTypeListener, nil
}
