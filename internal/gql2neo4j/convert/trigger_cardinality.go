package convert

import (
	"fmt"

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
	// Check if cardinality constraints are defined
	if edgeType.SourceCardinality == nil && edgeType.DestinationCardinality == nil {
		return nil, nil
	}

	// Check if default cardinality (no constraints)
	if (edgeType.SourceCardinality == nil || (edgeType.SourceCardinality.LowerBound == 0 && edgeType.SourceCardinality.UpperBound == -1)) &&
		(edgeType.DestinationCardinality == nil || (edgeType.DestinationCardinality.LowerBound == 0 && edgeType.DestinationCardinality.UpperBound == -1)) {
		return nil, nil
	}

	if len(edgeType.EdgeTypeLabelSet) != 1 {
		return nil, fmt.Errorf("error Neo4j only supports one label per edge type")
	}
	edgeTypeLabel := edgeType.EdgeTypeLabelSet[0]

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
}

// newCardinalityTrigger creates a new cardinality trigger.
func newCardinalityTrigger(edgeType *gql.EdgeType, relationshipType string) (*triggers.Trigger, error) {
	data := CardinalityTriggerData{
		RelationshipType: relationshipType,
	}

	if edgeType.SourceCardinality != nil {
		data.SourceLowerBound = edgeType.SourceCardinality.LowerBound
		data.SourceUpperBound = edgeType.SourceCardinality.UpperBound
	} else {
		data.SourceLowerBound = 0
		data.SourceUpperBound = -1
	}

	if edgeType.DestinationCardinality != nil {
		data.DestinationLowerBound = edgeType.DestinationCardinality.LowerBound
		data.DestinationUpperBound = edgeType.DestinationCardinality.UpperBound
	} else {
		data.DestinationLowerBound = 0
		data.DestinationUpperBound = -1
	}

	kernelTransaction, err := renderTemplate("cardinality_trigger.cypher.tmpl", data)
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := mustFormatName("cardinality", []string{relationshipType})
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}
