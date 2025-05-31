package gql

import "github.com/gql-schema/gqls/internal/gql/datatype"

// ExtendedElementAttributes represents extended element attributes.
type ExtendedElementAttributes struct {
	CompositePrimaryKeyFields       []string
	CompositeUniqueConstraintFields []string
}

// ExtendedPropertyAttributes represents extended property attributes.
type ExtendedPropertyAttributes struct {
	IsPrimaryKey    bool
	IsUnique        bool
	CheckConstraint *string
}

// PropertyType represents a property type.
type PropertyType struct {
	PropertyName       string
	PropertyValueType  datatype.Descriptor
	ExtendedAttributes *ExtendedPropertyAttributes
}

// NodeType represents a node type.
type NodeType struct {
	Alias              string
	NodeTypeLabelSet   []string
	PropertyTypeSet    []*PropertyType
	ExtendedAttributes *ExtendedElementAttributes
}

// Cardinality represents the cardinality of a node type.
type Cardinality struct {
	// LowerBound is the lower bound of the cardinality. 0 implies no lower bound.
	LowerBound int
	// UpperBound is the upper bound of the cardinality. math.MaxInt implies no upper bound.
	UpperBound int
}

// EdgeKind represents the stringKind of edge type.
type EdgeKind int

const (
	// EdgeKindDirected represents a directed edge type.
	EdgeKindDirected EdgeKind = iota
	// EdgeKindUndirected represents an undirected edge type.
	EdgeKindUndirected
)

// EdgeType represents an edge type.
//
// Each edge type specification has a set of edge type labels, a set of edge
// type properties, a source node type, a destination node type, and a
// directed flag.
type EdgeType struct {
	// EdgeTypeLabelSet comprises a set of zero or more labels.
	EdgeTypeLabelSet []string
	// EdgeTypePropertySet comprises a set of zero or more properties.
	EdgeTypePropertySet []*PropertyType
	// SourceNodeType one of the endpoint node types.
	SourceNodeType *NodeType
	// DestinationNodeType the other endpoint node type.
	DestinationNodeType *NodeType
	// SourceCardinality is the cardinality of the source node type.
	SourceCardinality *Cardinality
	// DestinationCardinality is the cardinality of the destination node type.
	DestinationCardinality *Cardinality
	// Kind is the edge type stringKind (directed or undirected).
	Kind EdgeKind
	// ExtendedAttributes represents extended attributes of the edge type.
	ExtendedAttributes *ExtendedElementAttributes
}

// GraphType represents a graph type as a collection of node and edge type specifications.
type GraphType struct {
	CatalogGraphParentAndName string
	// NodeTypeSet is the set of node type descriptors.
	NodeTypeSet []*NodeType
	// EdgeTypeSet is the set of edge type descriptors.
	EdgeTypeSet []*EdgeType
}
