package convert

import (
	"testing"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/gql/datatype"
)

func Test_newCheckConstraintTrigger(t *testing.T) {
	trigger, err := newCheckConstraintTrigger(NodeEntityType, "Person", "age", "age > 18")
	if err != nil {
		t.Fatalf("newCheckConstraintTrigger() error = %v", err)
	}

	got := trigger.CreateStatement()
	want := `CALL apoc.trigger.add("check_person_age", "UNWIND $createdNodes AS e WITH e WHERE e:Person AND e.age IS NOT NULL AND NOT (e.age > 18) CALL apoc.util.validate(true, 'age violates check constraint: e.age > 18', []) RETURN NULL UNION UNWIND keys($assignedNodeProperties) AS prop UNWIND $assignedNodeProperties[prop] AS change WITH change WHERE change.node:Person AND change.key = \"age\" AND change.new IS NOT NULL AND NOT (change.new > 18) CALL apoc.util.validate(true, 'age violates check constraint: change.new > 18', []) RETURN NULL", {phase: "before"})`
	if got != want {
		t.Errorf("newCheckConstraintTrigger() got = %s, want = %s", got, want)
	}
}

func Test_checkConstraintTriggerConverter_ConvertProperty(t *testing.T) {
	converter := &checkConstraintTriggerConverter{}

	tests := []struct {
		name         string
		propertyType *gql.PropertyType
		wantNil      bool
		wantErr      bool
	}{
		{
			name: "no extended attributes",
			propertyType: &gql.PropertyType{
				PropertyName:      "age",
				PropertyValueType: &datatype.NumericDataTypeDescriptor{},
			},
			wantNil: true,
		},
		{
			name: "no check constraint",
			propertyType: &gql.PropertyType{
				PropertyName:       "age",
				PropertyValueType:  &datatype.NumericDataTypeDescriptor{},
				ExtendedAttributes: &gql.ExtendedPropertyAttributes{},
			},
			wantNil: true,
		},
		{
			name: "empty check constraint",
			propertyType: &gql.PropertyType{
				PropertyName:      "age",
				PropertyValueType: &datatype.NumericDataTypeDescriptor{},
				ExtendedAttributes: &gql.ExtendedPropertyAttributes{
					CheckConstraint: stringPtr("   "),
				},
			},
			wantNil: true,
		},
		{
			name: "valid check constraint",
			propertyType: &gql.PropertyType{
				PropertyName:      "age",
				PropertyValueType: &datatype.NumericDataTypeDescriptor{},
				ExtendedAttributes: &gql.ExtendedPropertyAttributes{
					CheckConstraint: stringPtr("age > 18"),
				},
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := converter.ConvertProperty(NodeEntityType, "Person", tt.propertyType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ConvertProperty() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantNil && got != nil {
				t.Errorf("ConvertProperty() expected nil but got %v", got)
			}
			if !tt.wantNil && got == nil {
				t.Errorf("ConvertProperty() expected non-nil but got nil")
			}
		})
	}
}

func Test_processCheckExpression(t *testing.T) {
	tests := []struct {
		name         string
		expression   string
		propertyName string
		want         string
	}{
		{
			name:         "simple property reference",
			expression:   "age > 18",
			propertyName: "age",
			want:         "e.age > 18",
		},
		{
			name:         "property in complex expression",
			expression:   "age >= 0 AND age <= 120",
			propertyName: "age",
			want:         "e.age >= 0 AND e.age <= 120",
		},
		{
			name:         "no property reference",
			expression:   "true",
			propertyName: "age",
			want:         "true",
		},
		{
			name:         "property with function",
			expression:   "length(name) > 0",
			propertyName: "name",
			want:         "length(e.name) > 0",
		},
		{
			name:         "already processed expression",
			expression:   "e.age > 18",
			propertyName: "age",
			want:         "e.age > 18",
		},
		{
			name:         "expression with change context",
			expression:   "change.age > 18",
			propertyName: "age",
			want:         "change.age > 18",
		},
		{
			name:         "complex expression with multiple property references",
			expression:   "age >= 18 AND age <= 120",
			propertyName: "age",
			want:         "e.age >= 18 AND e.age <= 120",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := processCheckExpressionForContext(tt.expression, tt.propertyName, "e")
			if got != tt.want {
				t.Errorf("processCheckExpression() = %v, want %v", got, tt.want)
			}
		})
	}
}

func stringPtr(s string) *string {
	return &s
}
