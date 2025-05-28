// Package constraints provides a set of constraints for Neo4j.
package constraints

import (
	"fmt"
	"strings"

	"github.com/gql-schema/gqls/internal/neo4j"
)

// SubjectType defines the subject of a constraint.
type SubjectType uint8

const (
	// NodeType represents a node constraint.
	NodeType SubjectType = iota
	// RelationshipType represents a relationship constraint.
	RelationshipType
)

// Constraint is a common interface for all constraints.
//
//sumtype:decl
type Constraint interface {
	// sealed is a marker to ensure that all types are sealed.
	sealed()

	neo4j.SchemaObject
}

// PropertyUniquenessConstraint ensures uniqueness.
type PropertyUniquenessConstraint struct {
	Name          string
	Pattern       string
	Alias         string
	PropertyNames []string
}

var _ Constraint = (*PropertyUniquenessConstraint)(nil)

// CreateStatement generates the Cypher statement to create the property uniqueness constraint.
func (c *PropertyUniquenessConstraint) CreateStatement() string {
	if len(c.PropertyNames) == 0 {
		panic("property uniqueness constraint must have at least one property name")
	}

	aliasedPropertyNames := make([]string, len(c.PropertyNames))
	for i, name := range c.PropertyNames {
		aliasedPropertyNames[i] = fmt.Sprintf("%s.%s", c.Alias, name)
	}

	if len(aliasedPropertyNames) == 1 {
		return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS UNIQUE", c.Name, c.Pattern, aliasedPropertyNames[0])
	}
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE (%s) IS UNIQUE", c.Name, c.Pattern, strings.Join(aliasedPropertyNames, ", "))
}

func (*PropertyUniquenessConstraint) sealed() {}

// PropertyExistenceConstraint ensures a property is not null.
type PropertyExistenceConstraint struct {
	Name         string
	Pattern      string
	Alias        string
	PropertyName string
}

var _ Constraint = (*PropertyExistenceConstraint)(nil)

// CreateStatement generates the Cypher statement to create the property existence constraint.
func (c *PropertyExistenceConstraint) CreateStatement() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s.%s IS NOT NULL", c.Name, c.Pattern, c.Alias, c.PropertyName)
}

func (*PropertyExistenceConstraint) sealed() {}

// PropertyTypeConstraint enforces a specific type.
type PropertyTypeConstraint struct {
	Name         string
	Pattern      string
	Alias        string
	PropertyName string
	PropertyType string
}

var _ Constraint = (*PropertyTypeConstraint)(nil)

// CreateStatement generates the Cypher statement to create the property type constraint.
func (c *PropertyTypeConstraint) CreateStatement() string {
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s.%s IS :: %s", c.Name, c.Pattern, c.Alias, c.PropertyName, c.PropertyType)
}

func (*PropertyTypeConstraint) sealed() {}

// KeyConstraint ensures a property is a key.
type KeyConstraint struct {
	Name          string
	Pattern       string
	Alias         string
	PropertyNames []string
	SubjectType   SubjectType
}

var _ Constraint = (*KeyConstraint)(nil)

// CreateStatement generates the Cypher statement to create the key constraint.
func (c *KeyConstraint) CreateStatement() string {
	if len(c.PropertyNames) == 0 {
		panic("key constraint must have at least one property name")
	}

	aliasedPropertyNames := make([]string, len(c.PropertyNames))
	for i, name := range c.PropertyNames {
		aliasedPropertyNames[i] = fmt.Sprintf("%s.%s", c.Alias, name)
	}

	var constraintType string
	switch c.SubjectType {
	case NodeType:
		constraintType = "NODE KEY"
	case RelationshipType:
		constraintType = "RELATIONSHIP KEY"
	default:
		panic(fmt.Sprintf("invalid subject type: %d", c.SubjectType))
	}

	if len(aliasedPropertyNames) == 1 {
		return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE %s IS %s", c.Name, c.Pattern, aliasedPropertyNames[0], constraintType)
	}
	return fmt.Sprintf("CREATE CONSTRAINT %s FOR %s REQUIRE (%s) IS %s", c.Name, c.Pattern, strings.Join(aliasedPropertyNames, ", "), constraintType)
}

func (*KeyConstraint) sealed() {}
