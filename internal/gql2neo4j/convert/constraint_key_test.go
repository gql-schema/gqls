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

func Test_keyConstraintConverter_ConvertNode(t1 *testing.T) {
	type args struct {
		nodeType *gql.NodeType
	}
	tests := []struct {
		name    string
		args    args
		want    []neo4j.SchemaObject
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no composite key",
			args: args{
				nodeType: &gql.NodeType{
					NodeTypeLabelSet:   []string{"Person"},
					ExtendedAttributes: &gql.ExtendedElementAttributes{},
				},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "with composite key",
			args: args{
				nodeType: &gql.NodeType{
					NodeTypeLabelSet: []string{"Person"},
					ExtendedAttributes: &gql.ExtendedElementAttributes{
						CompositePrimaryKeyFields: []string{"id", "email"},
					},
				},
			},
			want: []neo4j.SchemaObject{
				&constraints.KeyConstraint{
					Name:          "key_person_id_email",
					Pattern:       getPattern(NodeEntityType, "n", "Person"),
					Alias:         "n",
					PropertyNames: []string{"id", "email"},
					SubjectType:   constraints.NodeType,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &keyConstraintConverter{}
			got, err := t.ConvertNode(tt.args.nodeType)
			if !tt.wantErr(t1, err, fmt.Sprintf("ConvertNode(%v)", tt.args.nodeType)) {
				return
			}
			assert.Equalf(t1, tt.want, got, "ConvertNode(%v)", tt.args.nodeType)
		})
	}
}

func Test_keyConstraintConverter_ConvertEdge(t1 *testing.T) {
	type args struct {
		edgeType *gql.EdgeType
	}
	tests := []struct {
		name    string
		args    args
		want    []neo4j.SchemaObject
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no composite key",
			args: args{
				edgeType: &gql.EdgeType{
					EdgeTypeLabelSet:   []string{"IS_FRIEND"},
					ExtendedAttributes: &gql.ExtendedElementAttributes{},
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "with composite key",
			args: args{
				edgeType: &gql.EdgeType{
					EdgeTypeLabelSet: []string{"IS_FRIEND"},
					ExtendedAttributes: &gql.ExtendedElementAttributes{
						CompositePrimaryKeyFields: []string{"from", "to"},
					},
				},
			},
			want: []neo4j.SchemaObject{
				&constraints.KeyConstraint{
					Name:          "key_is_friend_from_to",
					Pattern:       getPattern(EdgeEntityType, "n", "IS_FRIEND"),
					Alias:         "n",
					PropertyNames: []string{"from", "to"},
					SubjectType:   constraints.RelationshipType,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &keyConstraintConverter{}
			got, err := t.ConvertEdge(tt.args.edgeType)
			if !tt.wantErr(t1, err, fmt.Sprintf("ConvertEdge(%v)", tt.args.edgeType)) {
				return
			}
			assert.Equalf(t1, tt.want, got, "ConvertEdge(%v)", tt.args.edgeType)
		})
	}
}

func Test_keyConstraintConverter_ConvertProperty(t1 *testing.T) {
	type args struct {
		entityType   EntityType
		label        string
		propertyType *gql.PropertyType
	}
	tests := []struct {
		name    string
		args    args
		want    []neo4j.SchemaObject
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "no primary key",
			args: args{
				entityType: NodeEntityType,
				label:      "Person",
				propertyType: &gql.PropertyType{
					PropertyName:       "name",
					PropertyValueType:  &datatype.MiscellaneousDataTypeDescriptor{Kind: "DATETIME"},
					ExtendedAttributes: &gql.ExtendedPropertyAttributes{IsPrimaryKey: false},
				},
			},
			want:    nil,
			wantErr: assert.NoError,
		},
		{
			name: "node with primary key",
			args: args{
				entityType: NodeEntityType,
				label:      "Person",
				propertyType: &gql.PropertyType{
					PropertyName:       "id",
					PropertyValueType:  &datatype.MiscellaneousDataTypeDescriptor{Kind: "UUID"},
					ExtendedAttributes: &gql.ExtendedPropertyAttributes{IsPrimaryKey: true},
				},
			},
			want: []neo4j.SchemaObject{
				&constraints.KeyConstraint{
					Name:          "key_person_id",
					Pattern:       getPattern(NodeEntityType, "n", "Person"),
					Alias:         "n",
					PropertyNames: []string{"id"},
					SubjectType:   constraints.NodeType,
				},
			},
			wantErr: assert.NoError,
		},
		{
			name: "edge with primary key",
			args: args{
				entityType: EdgeEntityType,
				label:      "IS_FRIEND",
				propertyType: &gql.PropertyType{
					PropertyName:       "id",
					PropertyValueType:  &datatype.MiscellaneousDataTypeDescriptor{Kind: "UUID"},
					ExtendedAttributes: &gql.ExtendedPropertyAttributes{IsPrimaryKey: true},
				},
			},
			want: []neo4j.SchemaObject{
				&constraints.KeyConstraint{
					Name:          "key_is_friend_id",
					Pattern:       getPattern(EdgeEntityType, "n", "IS_FRIEND"),
					Alias:         "n",
					PropertyNames: []string{"id"},
					SubjectType:   constraints.RelationshipType,
				},
			},
			wantErr: assert.NoError,
		},
	}
	for _, tt := range tests {
		t1.Run(tt.name, func(t1 *testing.T) {
			t := &keyConstraintConverter{}
			got, err := t.ConvertProperty(tt.args.entityType, tt.args.label, tt.args.propertyType)
			if !tt.wantErr(t1, err, fmt.Sprintf("ConvertProperty(%v, %v, %v)", tt.args.entityType, tt.args.label, tt.args.propertyType)) {
				return
			}
			assert.Equalf(t1, tt.want, got, "ConvertProperty(%v, %v, %v)", tt.args.entityType, tt.args.label, tt.args.propertyType)
		})
	}
}
