package convert

import (
	"fmt"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/triggers"
)

var _ EdgeConverter = (*edgeEndpointsTriggerConverter)(nil)

type edgeEndpointsTriggerConverter struct{}

func (p *edgeEndpointsTriggerConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.APOCEnabled, nil
}

func (p *edgeEndpointsTriggerConverter) ConvertEdge(edgeType *gql.EdgeType) ([]neo4j.SchemaObject, error) {
	if edgeType.SourceNodeType == nil && edgeType.DestinationNodeType == nil {
		return nil, nil
	}
	if edgeType.SourceNodeType == nil || edgeType.DestinationNodeType == nil {
		return nil, fmt.Errorf("error Neo4j requires both source and destination node types for edge endpoints trigger")
	}

	if len(edgeType.EdgeTypeLabelSet) != 1 {
		return nil, fmt.Errorf("error Neo4j only supports one label per edge type")
	}
	edgeTypeLabel := edgeType.EdgeTypeLabelSet[0]

	if len(edgeType.SourceNodeType.NodeTypeLabelSet) != 1 {
		return nil, fmt.Errorf("error Neo4j only supports one label per source node type")
	}
	startNodeType := edgeType.SourceNodeType.NodeTypeLabelSet[0]

	if len(edgeType.DestinationNodeType.NodeTypeLabelSet) != 1 {
		return nil, fmt.Errorf("error Neo4j only supports one label per destination node type")
	}
	endNodeType := edgeType.DestinationNodeType.NodeTypeLabelSet[0]

	edgeEndpointTrigger, err := newEdgeEndpointTrigger(edgeTypeLabel, startNodeType, endNodeType)
	if err != nil {
		return nil, fmt.Errorf("error creating edge endpoint trigger: %w", err)
	}
	return []neo4j.SchemaObject{edgeEndpointTrigger}, nil
}

// EdgeEndpointTriggerData holds the data for an edge endpoint trigger.
type EdgeEndpointTriggerData struct {
	RelationshipType string
	StartNodeType    string
	EndNodeType      string
}

// newEdgeEndpointTrigger creates a new edge endpoint trigger.
func newEdgeEndpointTrigger(relationshipType, startNodeType, endNodeType string) (*triggers.Trigger, error) {
	kernelTransaction, err := renderTemplate("edge_endpoint_trigger.cypher.tmpl", EdgeEndpointTriggerData{
		RelationshipType: relationshipType,
		StartNodeType:    startNodeType,
		EndNodeType:      endNodeType,
	})
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := mustFormatName("edge_endpoint", []string{relationshipType, startNodeType, endNodeType})
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}
