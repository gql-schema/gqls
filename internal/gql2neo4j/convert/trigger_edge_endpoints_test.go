package convert

import (
	"testing"
)

func Test_newEdgeEndpointTrigger(t *testing.T) {
	trigger, err := newEdgeEndpointTrigger("LIKES", "Person", "Post")
	if err != nil {
		t.Fatalf("newEdgeEndpointTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("edge_endpoint_likes_person_post", "UNWIND $createdRelationships AS rel WITH rel WHERE type(rel) = \"LIKES\" AND NOT (startNode(rel):Person AND endNode(rel):Post) CALL apoc.util.validate( true, \"LIKES relationship must be between Person and Post\", [] ) RETURN NULL", {phase: "before"})`
	if got != want {
		t.Errorf("newEdgeEndpointTrigger() got = %s, want = %s", got, want)
	}
}
