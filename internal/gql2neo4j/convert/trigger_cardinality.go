package convert

import (
	"fmt"
	"math"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/triggers"
)

var _ EdgeConverter = (*cardinalityTriggerConverter)(nil)

type cardinalityTriggerConverter struct{}

func (p *cardinalityTriggerConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.APOCEnabled, nil
}

func (p *cardinalityTriggerConverter) ConvertEdge(edgeType *gql.EdgeType) ([]neo4j.SchemaObject, error) {
	if edgeType.SourceCardinality == nil && edgeType.DestinationCardinality == nil {
		return nil, nil
	}
	if (edgeType.SourceCardinality == nil || (edgeType.SourceCardinality.LowerBound == 0 && edgeType.SourceCardinality.UpperBound == math.MaxInt)) &&
		(edgeType.DestinationCardinality == nil || (edgeType.DestinationCardinality.LowerBound == 0 && edgeType.DestinationCardinality.UpperBound == math.MaxInt)) {
		return nil, nil
	}
	if edgeType.SourceNodeType == nil || edgeType.DestinationNodeType == nil {
		return nil, fmt.Errorf("error Neo4j requires both source and destination node types for cardinality trigger")
	}

	edgeTypeLabel, err := getPrimaryLabel(edgeType.EdgeTypeLabelSet)
	if err != nil {
		return nil, fmt.Errorf("error getting primary label for edge type: %w", err)
	}

	cardinalityTrigger, err := newCardinalityTrigger(edgeType, edgeTypeLabel)
	if err != nil {
		return nil, fmt.Errorf("error creating cardinality trigger: %w", err)
	}
	return []neo4j.SchemaObject{cardinalityTrigger}, nil
}

// CardinalityTriggerData holds the data for a cardinality trigger.
type CardinalityTriggerData struct {
	RelationshipType      string
	SourceLowerBound      int
	SourceUpperBound      int
	DestinationLowerBound int
	DestinationUpperBound int
	SourceNodeType        string
	DestinationNodeType   string
}

// newCardinalityTrigger creates a new cardinality trigger.
func newCardinalityTrigger(edgeType *gql.EdgeType, relationshipType string) (*triggers.Trigger, error) {
	data := CardinalityTriggerData{
		RelationshipType: relationshipType,
	}

	data.SourceLowerBound = 0
	data.SourceUpperBound = math.MaxInt
	if edgeType.SourceCardinality != nil {
		data.SourceLowerBound = edgeType.SourceCardinality.LowerBound
		data.SourceUpperBound = edgeType.SourceCardinality.UpperBound
	}

	data.DestinationLowerBound = 0
	data.DestinationUpperBound = math.MaxInt
	if edgeType.DestinationCardinality != nil {
		data.DestinationLowerBound = edgeType.DestinationCardinality.LowerBound
		data.DestinationUpperBound = edgeType.DestinationCardinality.UpperBound
	}

	data.SourceNodeType, _ = getPrimaryLabel(edgeType.SourceNodeType.NodeTypeLabelSet)
	data.DestinationNodeType, _ = getPrimaryLabel(edgeType.DestinationNodeType.NodeTypeLabelSet)

	kernelTransaction, err := renderTemplate("cardinality_trigger.cypher.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := mustFormatName("cardinality", []string{relationshipType})
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}
