package convert

import (
	"math"
	"testing"

	"github.com/gql-schema/gqls/internal/gql"
)

func Test_newCardinalityTrigger(t *testing.T) {
	edgeType := &gql.EdgeType{
		SourceNodeType: &gql.NodeType{
			NodeTypeLabelSet: []string{"User"},
		},
		SourceCardinality: &gql.Cardinality{
			LowerBound: 1,
			UpperBound: 5,
		},
		DestinationNodeType: &gql.NodeType{
			NodeTypeLabelSet: []string{"Post"},
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
	want := `CALL apoc.trigger.add("cardinality_likes", "UNWIND $createdNodes + [r IN $createdRelationships WHERE type(r) = 'LIKES' | startNode(r)] + [r IN $deletedRelationships WHERE type(r) = 'LIKES' | startNode(r)] AS node WITH DISTINCT node WHERE 'User' IN labels(node) AND node IS NOT NULL OPTIONAL MATCH (node)-[r:LIKES]->() WITH node, count(r) AS relCount CALL apoc.util.validate( relCount < 1 OR relCount > 5, 'User LIKES constraint violated', [] ) RETURN NULL UNION UNWIND $createdNodes + [r IN $createdRelationships WHERE type(r) = 'LIKES' | endNode(r)] + [r IN $deletedRelationships WHERE type(r) = 'LIKES' | endNode(r)] AS node WITH DISTINCT node WHERE 'Post' IN labels(node) AND node IS NOT NULL OPTIONAL MATCH ()-[r:LIKES]->(node) WITH node, count(r) AS relCount CALL apoc.util.validate( relCount < 0 OR relCount > 10, 'Post LIKES constraint violated', [] ) RETURN NULL", {phase: "before"})`
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
				EdgeTypeLabelSet:    []string{"LIKES"},
				SourceNodeType:      &gql.NodeType{NodeTypeLabelSet: []string{"User"}},
				DestinationNodeType: &gql.NodeType{NodeTypeLabelSet: []string{"Post"}},
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
				EdgeTypeLabelSet:    []string{"LIKES"},
				SourceNodeType:      &gql.NodeType{NodeTypeLabelSet: []string{"User"}},
				DestinationNodeType: &gql.NodeType{NodeTypeLabelSet: []string{"Post"}},
				SourceCardinality: &gql.Cardinality{
					LowerBound: 0,
					UpperBound: math.MaxInt,
				},
				DestinationCardinality: &gql.Cardinality{
					LowerBound: 0,
					UpperBound: math.MaxInt,
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
