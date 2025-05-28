package convert

import (
	"testing"

	"github.com/gql-schema/gqls/internal/gql"
)

func Test_newCardinalityTrigger(t *testing.T) {
	edgeType := &gql.EdgeType{
		SourceCardinality: &gql.Cardinality{
			LowerBound: 1,
			UpperBound: 5,
		},
		DestinationCardinality: &gql.Cardinality{
			LowerBound: 0,
			UpperBound: 10,
		},
	}

	trigger, err := newCardinalityTrigger(edgeType, "LIKES")
	if err != nil {
		t.Fatalf("newCardinalityTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("cardinality_likes", "UNWIND $createdRelationships AS rel WITH rel, startNode(rel) AS startNode, endNode(rel) AS endNode WHERE type(rel) = \"LIKES\" AND (size([(startNode)-[:LIKES]-() | 1]) < 1 OR size([(startNode)-[:LIKES]-() | 1]) > 5 OR size([()-[:LIKES]->(endNode) | 1]) > 10 ) CALL apoc.util.validate( true, \"LIKES relationship cardinality constraint violated\", [] ) RETURN null", {phase: "before"})`
	if got != want {
		t.Errorf("newCardinalityTrigger() got = %s, want = %s", got, want)
	}
}

func Test_cardinalityTriggerConverter_ConvertEdge(t *testing.T) {
	converter := &cardinalityTriggerConverter{}

	tests := []struct {
		name     string
		edgeType *gql.EdgeType
		wantNil  bool
	}{
		{
			name: "with cardinality constraints",
			edgeType: &gql.EdgeType{
				EdgeTypeLabelSet: []string{"LIKES"},
				SourceCardinality: &gql.Cardinality{
					LowerBound: 1,
					UpperBound: 5,
				},
				DestinationCardinality: &gql.Cardinality{
					LowerBound: 0,
					UpperBound: 10,
				},
			},
			wantNil: false,
		},
		{
			name: "without cardinality constraints",
			edgeType: &gql.EdgeType{
				EdgeTypeLabelSet: []string{"LIKES"},
			},
			wantNil: true,
		},
		{
			name: "with default cardinality",
			edgeType: &gql.EdgeType{
				EdgeTypeLabelSet: []string{"LIKES"},
				SourceCardinality: &gql.Cardinality{
					LowerBound: 0,
					UpperBound: -1,
				},
				DestinationCardinality: &gql.Cardinality{
					LowerBound: 0,
					UpperBound: -1,
				},
			},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.ConvertEdge(tt.edgeType)
			if err != nil {
				t.Errorf("ConvertEdge() error = %v", err)
				return
			}
			if tt.wantNil && got != nil {
				t.Errorf("ConvertEdge() expected nil, got %v", got)
			}
			if !tt.wantNil && got == nil {
				t.Errorf("ConvertEdge() expected non-nil result")
			}
		})
	}
}
