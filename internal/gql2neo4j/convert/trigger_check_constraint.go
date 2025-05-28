package convert

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/triggers"
)

var _ PropertyConverter = (*checkConstraintTriggerConverter)(nil)

type checkConstraintTriggerConverter struct{}

func (p *checkConstraintTriggerConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.APOCEnabled, nil
}

func (p *checkConstraintTriggerConverter) ConvertProperty(entityType EntityType, entityLabel string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	if propertyType.ExtendedAttributes == nil || propertyType.ExtendedAttributes.CheckConstraint == nil {
		return nil, nil
	}

	checkExpression := *propertyType.ExtendedAttributes.CheckConstraint
	if strings.TrimSpace(checkExpression) == "" {
		return nil, nil
	}

	trigger, err := newCheckConstraintTrigger(entityType, entityLabel, propertyType.PropertyName, checkExpression)
	if err != nil {
		return nil, fmt.Errorf("error creating check constraint trigger: %w", err)
	}
	return []neo4j.SchemaObject{trigger}, nil
}

// CheckConstraintTriggerData holds the data for a check constraint trigger.
type CheckConstraintTriggerData struct {
	CreatedVariable            string
	AssignedPropertiesVariable string
	Label                      string
	PropertyName               string
	CheckExpression            string
	ChangeCheckExpression      string
}

// newCheckConstraintTrigger creates a new check constraint trigger.
func newCheckConstraintTrigger(entityType EntityType, label, propertyName, checkExpression string) (*triggers.Trigger, error) {
	createdVariable, assignedPropertiesVariable := entityTypeToVariables(entityType)

	processedExpression := processCheckExpressionForContext(checkExpression, propertyName, "e")
	changeExpression := processCheckExpressionForContext(checkExpression, propertyName, "change")
	changeExpression = strings.ReplaceAll(changeExpression, "change."+propertyName, "change.new")

	kernelTransaction, err := renderTemplate("check_constraint_trigger.cypher.tmpl", CheckConstraintTriggerData{
		CreatedVariable:            createdVariable,
		AssignedPropertiesVariable: assignedPropertiesVariable,
		Label:                      label,
		PropertyName:               propertyName,
		CheckExpression:            processedExpression,
		ChangeCheckExpression:      changeExpression,
	})
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := mustFormatName("check", []string{label, propertyName})
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}

// processCheckExpressionForContext ensures property references in the check expression are prefixed with the correct context.
func processCheckExpressionForContext(expression, propertyName, context string) string {
	prefixedProperty := context + "." + propertyName
	if strings.Contains(expression, prefixedProperty) {
		return expression
	}

	pattern := `\b` + regexp.QuoteMeta(propertyName) + `\b`
	re := regexp.MustCompile(pattern)
	return re.ReplaceAllString(expression, prefixedProperty)
}
