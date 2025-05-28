package triggers

import "testing"

func TestTrigger_CreateStatement(t *testing.T) {
	trigger := &Trigger{
		Name:              "some_name",
		kernelTransaction: "FOO",
		Selector:          Selector{Phase: After},
	}

	want := "CALL apoc.trigger.add(\"some_name\", \"FOO\", {phase: \"after\"})"
	if got := trigger.CreateStatement(); got != want {
		t.Errorf("CreateStatement() = %v, want %v", got, want)
	}
}
