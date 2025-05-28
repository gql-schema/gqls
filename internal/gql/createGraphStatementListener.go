// Package gql provides functionality for parsing and handling GQL (Graph Query Language) statements.
package gql

import (
	"errors"
	"log"

	"github.com/gql-schema/gqls/internal/gql/parser"
)

// ErrNotImplemented indicates that a feature is not yet implemented.
var ErrNotImplemented = errors.New("feature not implemented")

// CreateGraphStatementListener is a custom ANTLR listener for parsing GQL create graph statements.
type CreateGraphStatementListener struct {
	CreateGraphTypeStatements []*GraphType
	*parser.BaseGQLListener
}

var _ parser.GQLListener = (*CreateGraphStatementListener)(nil)

// ExitCreateGraphStatement is called when exiting a CREATE GRAPH statement.
func (l *CreateGraphStatementListener) ExitCreateGraphStatement(ctx *parser.CreateGraphStatementContext) {
	graphName := ctx.CatalogGraphParentAndName().GetText()

	ofGraphType := ctx.OfGraphType()
	if ofGraphType == nil {
		log.Printf("Warning: No graph type found for graph name %s\n", graphName)
		return
	}

	nestedGraphTypeSpec := ofGraphType.NestedGraphTypeSpecification()
	if nestedGraphTypeSpec == nil {
		log.Printf("Warning: No nested graph type specification found for graph type %s\n", graphName)
		return
	}

	elementTypeSpecs := nestedGraphTypeSpec.GraphTypeSpecificationBody().ElementTypeList().AllElementTypeSpecification()
	if elementTypeSpecs == nil {
		log.Printf("Warning: No element type specifications found for graph type %s\n", graphName)
		return
	}

	graphType := &GraphType{CatalogGraphParentAndName: graphName}

	for _, elementTypeSpec := range elementTypeSpecs {
		if err := parseElementTypeSpecification(elementTypeSpec, graphType); err != nil {
			log.Printf("Error parsing element type spec for %s: %v", graphName, err)
			return
		}
	}

	l.CreateGraphTypeStatements = append(l.CreateGraphTypeStatements, graphType)
}
