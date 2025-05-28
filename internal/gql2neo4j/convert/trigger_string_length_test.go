package convert

import (
	"testing"
)

func Test_newStringLengthTrigger(t *testing.T) {
	trigger, err := newStringLengthTrigger(EdgeEntityType, "Person", "name", 1, 255)
	if err != nil {
		t.Fatalf("newStringLengthTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("string_length_person_name", "UNWIND $createdRelationships AS e WITH e WHERE e:Person AND e.name IS NOT NULL AND (size(e.name) < 1 OR size(e.name) > 255) CALL apoc.util.validate(true, 'name must be between 1 and 255 characters', []) RETURN NULL UNION UNWIND keys($assignedRelationshipProperties) AS prop UNWIND $assignedRelationshipProperties[prop] AS change WITH change WHERE change.relationship:Person AND change.key = \"name\" AND change.new IS NOT NULL AND (size(change.new) < 1 OR size(change.new) > 255) CALL apoc.util.validate(true, 'name must be between 1 and 255 characters', []) RETURN NULL", {phase: "before"})`
	if got != want {
		t.Errorf("newStringLengthTrigger() got = %s, want = %s", got, want)
	}
}
