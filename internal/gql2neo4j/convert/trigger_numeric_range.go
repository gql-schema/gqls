package convert

import (
	"fmt"
	"math"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/gql/datatype"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/triggers"
)

var _ PropertyConverter = (*numericRangeTriggerConverter)(nil)

type numericRangeTriggerConverter struct{}

func (p *numericRangeTriggerConverter) IsCompatible(meta *DatabaseMetadata) (bool, error) {
	return meta.APOCEnabled, nil
}

func (p *numericRangeTriggerConverter) ConvertProperty(entityType EntityType, entityLabel string, propertyType *gql.PropertyType) ([]neo4j.SchemaObject, error) {
	typeDesc, ok := propertyType.PropertyValueType.(*datatype.NumericDataTypeDescriptor)
	if !ok {
		return nil, nil
	}
	if typeDesc.Precision >= 64 {
		return nil, fmt.Errorf("error Neo4j does not support precision >= 64 for numeric types")
	}
	if typeDesc.Scale != 0 {
		return nil, fmt.Errorf("error Neo4j does not support scale for numeric types")
	}

	var schemaObjects []neo4j.SchemaObject
	if typeDesc.Precision != 0 && typeDesc.Precision != 63 {
		trigger, err := newNumericRangeTrigger(entityType, entityLabel, propertyType.PropertyName, typeDesc.Precision)
		if err != nil {
			return nil, fmt.Errorf("error creating numeric range trigger: %w", err)
		}
		schemaObjects = append(schemaObjects, trigger)
	}
	return schemaObjects, nil
}

// NumericRangeTriggerData holds the data for a numeric range trigger.
type NumericRangeTriggerData struct {
	CreatedVariable            string
	AssignedPropertiesVariable string
	Label                      string
	PropertyName               string
	MinValue                   int
	MaxValue                   int
}

// newNumericRangeTrigger creates a new numeric range trigger.
func newNumericRangeTrigger(entityType EntityType, label, propertyName string, precision int) (*triggers.Trigger, error) {
	minValue, maxValue, err := computeBoundaries(precision)
	if err != nil {
		return nil, fmt.Errorf("error computing boundaries: %w", err)
	}

	createdVariable, assignedPropertiesVariable := entityTypeToVariables(entityType)
	kernelTransaction, err := renderTemplate("numeric_range_trigger.cypher.tmpl", NumericRangeTriggerData{
		CreatedVariable:            createdVariable,
		AssignedPropertiesVariable: assignedPropertiesVariable,
		Label:                      label,
		PropertyName:               propertyName,
		MinValue:                   minValue,
		MaxValue:                   maxValue,
	})
	if err != nil {
		return nil, fmt.Errorf("error rendering template: %w", err)
	}

	triggerName := mustFormatName("numeric_range", []string{label, propertyName})
	return triggers.NewTrigger(triggerName, kernelTransaction, triggers.Selector{Phase: triggers.Before}), nil
}

// computeBoundaries computes the minimum and maximum values for a numeric range based on the precision.
func computeBoundaries(precision int) (int, int, error) {
	if precision < 1 || precision > 63 {
		return 0, 0, fmt.Errorf("precision must be between 1 and 63")
	}
	minValue := -int(math.Pow(2, float64(precision)))
	maxValue := int(math.Pow(2, float64(precision))) - 1
	return minValue, maxValue, nil
}
