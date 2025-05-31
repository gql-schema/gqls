package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gql-schema/gqls/internal/gql"
	"github.com/gql-schema/gqls/internal/gql/datatype"
	"github.com/gql-schema/gqls/internal/neo4j"
	"github.com/gql-schema/gqls/internal/neo4j/constraints"
)

func Test_existenceConstraintConverter_ConvertProperty(t1 *testing.T) {
	type args struct {
		entityType   EntityType
		entityLabel  string
		propertyType *gql.PropertyType
	}
	tests := []struct {
		name    string
		args    args
		want    []neo4j.SchemaObject
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "non-nullable property type",
			args: args{
				entityType:  NodeEntityType,
				entityLabel: "Person",
				propertyType: &gql.PropertyType{
					PropertyName: "name",
					PropertyValueType: &datatype.NumericDataTypeDescriptor{
						Kind:      datatype.NumericKindExact,
						Radix:     datatype.RadixDecimal,
						Precision: 7,
						Nullable:  false,
					},
				},
			},
			want: []neo4j.SchemaObject{
				&constraints.PropertyExistenceConstraint{
					Name:         mustFormatName("existence", []string{"Person", "name"}),
					Pattern:      getPattern(NodeEntityType, "n", "Person"),
					Alias:        "n",
					PropertyName: "name",
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "nullable property type",
			args: args{
				entityType:  NodeEntityType,
				entityLabel: "Person",
				propertyType: &gql.PropertyType{
					PropertyName: "age",
					PropertyValueType: &datatype.NumericDataTypeDescriptor{
						Kind:      datatype.NumericKindExact,
						Radix:     datatype.RadixDecimal,
						Precision: 7,
						Nullable:  true,
					},
				},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &existenceConstraintConverter{}
			got, err := t.ConvertProperty(tt.args.entityType, tt.args.entityLabel, tt.args.propertyType)
			if !tt.wantErr(t1, err, fmt.Sprintf("ConvertProperty(%v, %v, %v)", tt.args.entityType, tt.args.entityLabel, tt.args.propertyType)) {
				return
			}
			assert.Equalf(t1, tt.want, got, "ConvertProperty(%v, %v, %v)", tt.args.entityType, tt.args.entityLabel, tt.args.propertyType)
		})
	}
}
