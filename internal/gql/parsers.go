package gql

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/antlr4-go/antlr/v4"

	"github.com/gql-schema/gqls/internal/gql/datatype"
	"github.com/gql-schema/gqls/internal/gql/parser"
)

// parseElementTypeSpecification parses the element type specification from the context.
func parseElementTypeSpecification(elementTypeSpec parser.IElementTypeSpecificationContext, graphType *GraphType) error {
	var compositePrimaryKeyFields []string
	if compositePrimaryKey := elementTypeSpec.CompositePrimaryKey(); compositePrimaryKey != nil {
		for _, propertyName := range compositePrimaryKey.FieldNameList().AllFieldName() {
			compositePrimaryKeyFields = append(compositePrimaryKeyFields, propertyName.Identifier().GetText())
		}
	}

	var compositeUniqueConstraintFields []string
	if compositeUniqueConstraint := elementTypeSpec.CompositeUniqueConstraint(); compositeUniqueConstraint != nil {
		for _, propertyName := range compositeUniqueConstraint.FieldNameList().AllFieldName() {
			compositeUniqueConstraintFields = append(compositeUniqueConstraintFields, propertyName.Identifier().GetText())
		}
	}

	if nodeTypeSpec := elementTypeSpec.NodeTypeSpecification(); nodeTypeSpec != nil {
		nodeType, err := parseNodeTypeSpecification(nodeTypeSpec)
		if err != nil {
			return fmt.Errorf("error parsing node type nodeType: %w", err)
		}
		nodeType.ExtendedAttributes = &ExtendedElementAttributes{
			CompositePrimaryKeyFields:       compositePrimaryKeyFields,
			CompositeUniqueConstraintFields: compositeUniqueConstraintFields,
		}
		graphType.NodeTypeSet = append(graphType.NodeTypeSet, nodeType)
	}

	if edgeTypeSpec := elementTypeSpec.EdgeTypeSpecification(); edgeTypeSpec != nil {
		edgeType, err := parseEdgeTypeSpecification(edgeTypeSpec)
		if err != nil {
			return fmt.Errorf("error parsing edge type edgeType: %w", err)
		}
		edgeType.ExtendedAttributes = &ExtendedElementAttributes{
			CompositePrimaryKeyFields:       compositePrimaryKeyFields,
			CompositeUniqueConstraintFields: compositeUniqueConstraintFields,
		}
		graphType.EdgeTypeSet = append(graphType.EdgeTypeSet, edgeType)
	}

	return nil
}

// parseNodeTypeSpecification parses the node type specification from the context.
func parseNodeTypeSpecification(ctx parser.INodeTypeSpecificationContext) (*NodeType, error) {
	var nodeTypeFiller parser.INodeTypeFillerContext
	var nodeTypeAlias parser.ILocalNodeTypeAliasContext

	if nodeTypePattern := ctx.NodeTypePattern(); nodeTypePattern != nil {
		nodeTypeFiller = nodeTypePattern.NodeTypeFiller()
		nodeTypeAlias = nodeTypePattern.LocalNodeTypeAlias()
	} else if nodeTypePhrase := ctx.NodeTypePhrase(); nodeTypePhrase != nil {
		nodeTypeFiller = nodeTypePhrase.NodeTypePhraseFiller().NodeTypeFiller()
		nodeTypeAlias = nodeTypePhrase.LocalNodeTypeAlias()
	} else {
		return nil, fmt.Errorf("error invalid node type specification: %s", ctx.GetText())
	}

	nodeType, err := parseNodeTypeFiller(nodeTypeFiller)
	if err != nil {
		return nil, fmt.Errorf("error parsing node type filler: %w", err)
	}

	if nodeTypeAlias != nil {
		nodeType.Alias = nodeTypeAlias.GetText()
	}

	return nodeType, nil
}

// parseNodeTypeFiller parses the node type filler from the context.
func parseNodeTypeFiller(fillerContext parser.INodeTypeFillerContext) (*NodeType, error) {
	impliedContent := fillerContext.NodeTypeImpliedContent()
	if impliedContent == nil {
		return nil, fmt.Errorf("error missing implied content in node type filler")
	}

	nodeTypePropertyTypes, err := parseNodeTypePropertyTypes(impliedContent.NodeTypePropertyTypes())
	if err != nil {
		return nil, fmt.Errorf("error parsing node type property types: %w", err)
	}

	return &NodeType{
		NodeTypeLabelSet: parseLabelSetPhrase(impliedContent.NodeTypeLabelSet().LabelSetPhrase()),
		PropertyTypeSet:  nodeTypePropertyTypes,
	}, nil
}

// parseLabelSetPhrase parses the label set phrase from the context.
func parseLabelSetPhrase(ctx parser.ILabelSetPhraseContext) []string {
	labelNames := ctx.LabelSetSpecification().AllLabelName()

	keyLabelSet := make([]string, 0, len(labelNames))
	for _, label := range labelNames {
		keyLabelSet = append(keyLabelSet, label.GetText())
	}
	return keyLabelSet
}

// parseNodeTypePropertyTypes parses the property types for a node type.
func parseNodeTypePropertyTypes(ctx parser.INodeTypePropertyTypesContext) ([]*PropertyType, error) {
	if ctx == nil {
		return []*PropertyType{}, nil
	}

	spec, err := parsePropertyTypesSpecification(ctx.PropertyTypesSpecification())
	if err != nil {
		return nil, fmt.Errorf("error parsing node type property types: %w", err)
	}
	return spec, nil
}

// parsePropertyTypesSpecification parses the property types specification.
func parsePropertyTypesSpecification(ctx parser.IPropertyTypesSpecificationContext) ([]*PropertyType, error) {
	propertyTypes := ctx.PropertyTypeList().AllPropertyType()

	propertyTypeSet := make([]*PropertyType, 0, len(propertyTypes))
	for _, propertyType := range propertyTypes {
		propertyValueType, err := parseValueType(propertyType)
		if err != nil {
			return nil, fmt.Errorf("error parsing property value type: %w", err)
		}

		extendedAttributes, err := parseExtendedAttributes(propertyType)
		if err != nil {
			return nil, fmt.Errorf("error parsing extended attributes: %w", err)
		}

		propertyTypeSet = append(propertyTypeSet, &PropertyType{
			PropertyName:       propertyType.PropertyName().GetText(),
			PropertyValueType:  propertyValueType,
			ExtendedAttributes: extendedAttributes,
		})
	}
	return propertyTypeSet, nil
}

func parseExtendedAttributes(ctx parser.IPropertyTypeContext) (*ExtendedPropertyAttributes, error) {
	propertyValueType := ctx.PropertyValueType()
	if propertyValueType == nil {
		return nil, fmt.Errorf("missing property value type in extended attributes")
	}

	var checkConstraint *string
	allCheckConstraints := propertyValueType.AllCheckConstraint()
	if len(allCheckConstraints) > 1 {
		return nil, fmt.Errorf("multiple check constraints are not supported")
	} else if len(allCheckConstraints) == 1 {
		checkConstraintText := ""
		valueExpression := allCheckConstraints[0].SearchCondition().BooleanValueExpression().ValueExpression()
		for i, child := range valueExpression.GetChildren() {
			if i > 0 {
				checkConstraintText += " "
			}
			checkConstraintText += child.(antlr.ParseTree).GetText()
		}

		checkConstraintText = strings.TrimSpace(checkConstraintText)
		if checkConstraintText != "" {
			checkConstraint = &checkConstraintText
		}
	}

	return &ExtendedPropertyAttributes{
		IsPrimaryKey:    len(propertyValueType.AllPrimaryKey()) > 0,
		IsUnique:        len(propertyValueType.AllUNIQUE()) > 0,
		CheckConstraint: checkConstraint,
	}, nil
}

// parseValueType parses the value type of a property.
func parseValueType(ctx parser.IPropertyTypeContext) (datatype.Descriptor, error) {
	propertyValueType := ctx.PropertyValueType()

	predefinedTypeLabelCtx, ok := propertyValueType.ValueType().(*parser.PredefinedTypeLabelContext)
	if !ok {
		return nil, fmt.Errorf("error parsing value type: non-predefined types not supported")
	}

	// Handle predefined types with special handling
	if predefinedType := predefinedTypeLabelCtx.PredefinedType(); predefinedType != nil {
		if characterStringType := predefinedType.CharacterStringType(); characterStringType != nil {
			return parseStringType(StringKindCharacterString, characterStringType)
		} else if byteStringType := predefinedType.ByteStringType(); byteStringType != nil {
			return parseStringType(StringKindByteString, byteStringType)
		} else if numericType := predefinedType.NumericType(); numericType != nil {
			return parseNumericType(numericType)
		}
	}

	// Predefined types without special handling
	valueType := propertyValueType.GetText()
	notNull := strings.HasSuffix(valueType, "NOTNULL")
	if notNull {
		valueType = strings.TrimSuffix(valueType, "NOTNULL")
	}
	return &datatype.MiscellaneousDataTypeDescriptor{
		Kind:     valueType,
		Nullable: !notNull,
	}, nil
}

// stringKind represents the kind of string type.
type stringKind int

const (
	// StringKindCharacterString represents a character string type.
	StringKindCharacterString stringKind = iota
	// StringKindByteString represents a byte string type.
	StringKindByteString
)

// stringType is an interface for string types.
type stringType interface {
	MinLength() parser.IMinLengthContext
	MaxLength() parser.IMaxLengthContext
	FixedLength() parser.IFixedLengthContext
	NotNull() parser.INotNullContext
}

// parseStringType parses the string type from the context.
func parseStringType(stringKind stringKind, stringType stringType) (datatype.Descriptor, error) {
	var minLength, maxLength int
	if stringType.MinLength() != nil {
		minLength, _ = strconv.Atoi(stringType.MinLength().GetText())
	}
	if stringType.MaxLength() != nil {
		maxLength, _ = strconv.Atoi(stringType.MaxLength().GetText())
	}
	if stringType.FixedLength() != nil {
		minLength, _ = strconv.Atoi(stringType.FixedLength().GetText())
		maxLength, _ = strconv.Atoi(stringType.FixedLength().GetText())
	}

	switch stringKind {
	case StringKindCharacterString:
		return &datatype.CharacterStringDataTypeDescriptor{
			MinLength: minLength,
			MaxLength: maxLength,
			Nullable:  parseNotNull(stringType.NotNull()),
		}, nil
	case StringKindByteString:
		return &datatype.ByteStringDataTypeDescriptor{
			MinLength: minLength,
			MaxLength: maxLength,
			Nullable:  parseNotNull(stringType.NotNull()),
		}, nil
	}

	panic("unreachable")
}

// parseNumericType parses the numeric type from the context.
func parseNumericType(numericType parser.INumericTypeContext) (datatype.Descriptor, error) {
	if exactNumericType := numericType.ExactNumericType(); exactNumericType != nil {
		return parseExactNumericType(exactNumericType)
	} else if approximateNumericType := numericType.ApproximateNumericType(); approximateNumericType != nil {
		return parseApproximateNumericType(approximateNumericType)
	}
	panic("unreachable")
}

// parseExactNumericType parses an exact numeric type (e.g., INT8, INT16).
func parseExactNumericType(ctx parser.IExactNumericTypeContext) (*datatype.NumericDataTypeDescriptor, error) {
	if decimalExactNumericType := ctx.DecimalExactNumericType(); decimalExactNumericType != nil {
		return nil, fmt.Errorf("decimal exact numeric type is not supported: %w", ErrNotImplemented)
	} else if binaryExactNumericType := ctx.BinaryExactNumericType(); binaryExactNumericType != nil {
		return parseBinaryExactNumericType(binaryExactNumericType)
	}
	panic("unreachable")
}

// parseBinaryExactNumericType parses the binary exact numeric type (e.g., INT8, INT16).
func parseBinaryExactNumericType(ctx parser.IBinaryExactNumericTypeContext) (*datatype.NumericDataTypeDescriptor, error) {
	if signedBinaryExactNumericType := ctx.SignedBinaryExactNumericType(); signedBinaryExactNumericType != nil {
		var precision int
		switch {
		case signedBinaryExactNumericType.INT8() != nil:
			precision = 7
		case signedBinaryExactNumericType.INT16() != nil:
			precision = 15
		case signedBinaryExactNumericType.INT32() != nil:
			precision = 31
		case signedBinaryExactNumericType.INT64() != nil:
			precision = 63
		case signedBinaryExactNumericType.INT128() != nil:
			precision = 127
		case signedBinaryExactNumericType.INT256() != nil:
			precision = 255
		default:
			// TODO: Handle REAL, FLOAT, and DOUBLE types
			parsedPrecision, err := parseOptionalInt(signedBinaryExactNumericType.Precision())
			if err != nil {
				return nil, fmt.Errorf("error parsing precision: %w", err)
			}
			precision = parsedPrecision
		}

		return &datatype.NumericDataTypeDescriptor{
			Kind:      datatype.NumericKindExact,
			Radix:     datatype.RadixBinary,
			Precision: precision,
			Nullable:  parseNotNull(signedBinaryExactNumericType.NotNull()),
		}, nil
	} else if unsignedBinaryExactNumericType := ctx.UnsignedBinaryExactNumericType(); unsignedBinaryExactNumericType != nil {
		return nil, fmt.Errorf("unsigned binary exact numeric type is not supported: %w", ErrNotImplemented)
	}
	panic("unreachable")
}

// parseApproximateNumericType parses the approximate numeric type (e.g., FLOAT16, FLOAT32).
func parseApproximateNumericType(ctx parser.IApproximateNumericTypeContext) (*datatype.NumericDataTypeDescriptor, error) {
	var precision int
	switch {
	case ctx.FLOAT16() != nil:
		precision = 15
	case ctx.FLOAT32() != nil:
		precision = 31
	case ctx.FLOAT64() != nil:
		precision = 63
	case ctx.FLOAT128() != nil:
		precision = 127
	case ctx.FLOAT256() != nil:
		precision = 255
	default:
		// TODO: Handle SMALLINT, INTEGER, and BIGINT types
		parsedPrecision, err := parseOptionalInt(ctx.Precision())
		if err != nil {
			return nil, fmt.Errorf("error parsing precision: %w", err)
		}
		precision = parsedPrecision
	}

	scale, err := parseOptionalInt(ctx.Scale())
	if err != nil {
		return nil, fmt.Errorf("error parsing scale: %w", err)
	}

	return &datatype.NumericDataTypeDescriptor{
		Kind:      datatype.NumericKindFloat,
		Radix:     datatype.RadixBinary,
		Precision: precision,
		Scale:     scale,
		Nullable:  parseNotNull(ctx.NotNull()),
	}, nil
}

// parseOptionalInt parses an optional integer value from a parser rule context.
func parseOptionalInt(node antlr.ParserRuleContext) (int, error) {
	if node == nil {
		return 0, nil
	}

	// somewhat hacky, but avoids some more complex logic
	value, err := strconv.Atoi(node.GetText())
	if err != nil {
		return 0, fmt.Errorf("error parsing integer value: %w", err)
	}
	return value, nil
}

// parseNotNull checks if the NOT NULL constraint is absent.
func parseNotNull(notNullContext parser.INotNullContext) bool {
	return notNullContext == nil || notNullContext.NOT() == nil
}

// parseEdgeTypeSpecification parses the edge type specification.
func parseEdgeTypeSpecification(ctx parser.IEdgeTypeSpecificationContext) (*EdgeType, error) {
	if edgeTypePattern := ctx.EdgeTypePattern(); edgeTypePattern != nil {
		edgeType, err := parseEdgeTypePattern(edgeTypePattern)
		if err != nil {
			return nil, fmt.Errorf("error parsing edge type pattern: %w", err)
		}
		return edgeType, nil
	} else if edgeTypePhrase := ctx.EdgeTypePhrase(); edgeTypePhrase != nil {
		// edge type phrases work with aliases; they are not yet implemented
		return nil, fmt.Errorf("error edge type phrase not yet implemented: %w", ErrNotImplemented)
	}
	panic("unreachable")
}

// parseEdgeTypePattern parses the edge type pattern.
func parseEdgeTypePattern(ctx parser.IEdgeTypePatternContext) (*EdgeType, error) {
	sourceCardinality := Cardinality{0, -1}
	destinationCardinality := Cardinality{0, -1}

	if directed := ctx.EdgeTypePatternDirected(); directed != nil {
		if pointingRight := directed.EdgeTypePatternPointingRight(); pointingRight != nil {

			var edgeTypeFiller parser.IEdgeTypeFillerContext
			switch {
			case pointingRight.ArcTypePointingRight().SimpleArcTypePointingRight() != nil:
				edgeTypeFiller = pointingRight.ArcTypePointingRight().SimpleArcTypePointingRight().EdgeTypeFiller()
			case pointingRight.ArcTypePointingRight().ArcWithCardinalityPointingRight() != nil:
				edgeTypeFiller = pointingRight.ArcTypePointingRight().ArcWithCardinalityPointingRight().EdgeTypeFiller()

				source, err := parseCardinality(pointingRight.ArcTypePointingRight().ArcWithCardinalityPointingRight().Cardinality(0))
				if err != nil {
					return nil, fmt.Errorf("error parsing cardinality: %w", err)
				}
				sourceCardinality = source

				destination, err := parseCardinality(pointingRight.ArcTypePointingRight().ArcWithCardinalityPointingRight().Cardinality(1))
				if err != nil {
					return nil, fmt.Errorf("error parsing cardinality: %w", err)
				}
				destinationCardinality = destination
			}

			e, err := parseEdgeType(EdgeKindDirected, pointingRight.SourceNodeTypeReference(), pointingRight.DestinationNodeTypeReference(), edgeTypeFiller)
			if err != nil {
				return nil, fmt.Errorf("error parsing edge type: %w", err)
			}

			e.SourceCardinality = &sourceCardinality
			e.DestinationCardinality = &destinationCardinality
			return e, nil
		} else if pointingLeft := directed.EdgeTypePatternPointingLeft(); pointingLeft != nil {
			var edgeTypeFiller parser.IEdgeTypeFillerContext
			switch {
			case pointingLeft.ArcTypePointingLeft().SimpleArcTypePointingLeft() != nil:
				edgeTypeFiller = pointingLeft.ArcTypePointingLeft().SimpleArcTypePointingLeft().EdgeTypeFiller()
			case pointingLeft.ArcTypePointingLeft().ArcWithCardinalityPointingLeft() != nil:
				edgeTypeFiller = pointingLeft.ArcTypePointingLeft().ArcWithCardinalityPointingLeft().EdgeTypeFiller()
				source, err := parseCardinality(pointingLeft.ArcTypePointingLeft().ArcWithCardinalityPointingLeft().Cardinality(0))
				if err != nil {
					return nil, fmt.Errorf("error parsing cardinality: %w", err)
				}
				sourceCardinality = source

				destination, err := parseCardinality(pointingLeft.ArcTypePointingLeft().ArcWithCardinalityPointingLeft().Cardinality(1))
				if err != nil {
					return nil, fmt.Errorf("error parsing cardinality: %w", err)
				}
				destinationCardinality = destination
			}

			edgeType, err := parseEdgeType(EdgeKindDirected, pointingLeft.SourceNodeTypeReference(), pointingLeft.DestinationNodeTypeReference(), edgeTypeFiller)
			if err != nil {
				return nil, fmt.Errorf("error parsing edge type: %w", err)
			}

			edgeType.SourceCardinality = &sourceCardinality
			edgeType.DestinationCardinality = &destinationCardinality
			return edgeType, nil
		}
		return nil, fmt.Errorf("error invalid directed edge type pattern: %s", ctx.GetText())
	} else if undirected := ctx.EdgeTypePatternUndirected(); undirected != nil {
		var edgeTypeFiller parser.IEdgeTypeFillerContext
		switch {
		case undirected.ArcTypeUndirected().SimpleArcTypeUndirected() != nil:
			edgeTypeFiller = undirected.ArcTypeUndirected().SimpleArcTypeUndirected().EdgeTypeFiller()
		case undirected.ArcTypeUndirected().ArcWithCardinalityUndirected() != nil:
			edgeTypeFiller = undirected.ArcTypeUndirected().ArcWithCardinalityUndirected().EdgeTypeFiller()
			source, err := parseCardinality(undirected.ArcTypeUndirected().ArcWithCardinalityUndirected().Cardinality(0))
			if err != nil {
				return nil, fmt.Errorf("error parsing cardinality: %w", err)
			}
			sourceCardinality = source
			destination, err := parseCardinality(undirected.ArcTypeUndirected().ArcWithCardinalityUndirected().Cardinality(1))
			if err != nil {
				return nil, fmt.Errorf("error parsing cardinality: %w", err)
			}
			destinationCardinality = destination
		}

		edgeType, err := parseEdgeType(EdgeKindUndirected, undirected.SourceNodeTypeReference(), undirected.DestinationNodeTypeReference(), edgeTypeFiller)
		if err != nil {
			return nil, fmt.Errorf("error parsing edge type: %w", err)
		}

		edgeType.SourceCardinality = &sourceCardinality
		edgeType.DestinationCardinality = &destinationCardinality
		return edgeType, nil
	}
	return nil, fmt.Errorf("error invalid edge type pattern: %s", ctx.GetText())
}

func parseCardinality(ctx parser.ICardinalityContext) (Cardinality, error) {
	cardinality := Cardinality{LowerBound: 0, UpperBound: -1}
	if ctx == nil {
		return cardinality, nil
	}

	if cardinalityFiller := ctx.CardinalityFiller(); cardinalityFiller != nil {
		if cardinalityLowerBound := cardinalityFiller.CardinalityLowerBound(); cardinalityLowerBound != nil {
			lowerBound, err := strconv.Atoi(cardinalityLowerBound.UNSIGNED_DECIMAL_INTEGER().GetText())
			if err != nil {
				return Cardinality{}, fmt.Errorf("error parsing lower bound: %w", err)
			}
			cardinality.LowerBound = lowerBound
		}
		if cardinalityUpperBound := cardinalityFiller.CardinalityUpperBound(); cardinalityUpperBound != nil {
			if cardinalityUpperBound.ASTERISK() != nil {
				cardinality.UpperBound = -1
			} else {
				upperBound, err := strconv.Atoi(cardinalityUpperBound.UNSIGNED_DECIMAL_INTEGER().GetText())
				if err != nil {
					return Cardinality{}, fmt.Errorf("error parsing upper bound: %w", err)
				}
				cardinality.UpperBound = upperBound
			}
		}
	}

	return cardinality, nil
}

// parseEdgeType parses the edge type from the context.
func parseEdgeType(
	edgeTypeKind EdgeKind,
	sourceCtx parser.ISourceNodeTypeReferenceContext,
	destCtx parser.IDestinationNodeTypeReferenceContext,
	edgeTypeFillerCtx parser.IEdgeTypeFillerContext,
) (*EdgeType, error) {
	var err error

	var source *NodeType
	if nodeTypeAlias := sourceCtx.SourceNodeTypeAlias(); nodeTypeAlias != nil {
		source = &NodeType{Alias: nodeTypeAlias.GetText()}
	} else {
		if nodeTypeFiller := sourceCtx.NodeTypeFiller(); nodeTypeFiller != nil {
			source, err = parseNodeTypeFiller(nodeTypeFiller)
			if err != nil {
				return nil, fmt.Errorf("error parsing source node type: %w", err)
			}
		}
	}

	var destination *NodeType
	if nodeTypeAlias := destCtx.DestinationNodeTypeAlias(); nodeTypeAlias != nil {
		destination = &NodeType{Alias: nodeTypeAlias.GetText()}
	} else {
		if nodeTypeFiller := destCtx.NodeTypeFiller(); nodeTypeFiller != nil {
			destination, err = parseNodeTypeFiller(nodeTypeFiller)
			if err != nil {
				return nil, fmt.Errorf("error parsing destination node type: %w", err)
			}
		}
	}

	labelSet, propertySet, err := parseEdgeTypeFiller(edgeTypeFillerCtx)
	if err != nil {
		return nil, fmt.Errorf("error parsing property types specification: %w", err)
	}

	return &EdgeType{
		Kind:                edgeTypeKind,
		SourceNodeType:      source,
		DestinationNodeType: destination,
		EdgeTypeLabelSet:    labelSet,
		EdgeTypePropertySet: propertySet,
	}, nil
}

// parseEdgeTypeFiller parses the edge type filler.
func parseEdgeTypeFiller(ctx parser.IEdgeTypeFillerContext) ([]string, []*PropertyType, error) {
	impliedContent := ctx.EdgeTypeImpliedContent()
	if impliedContent == nil {
		return nil, nil, fmt.Errorf("error: missing implied content in edge type filler")
	}

	var propertyTypesSpec []*PropertyType
	if edgeTypePropertyTypes := impliedContent.EdgeTypePropertyTypes(); edgeTypePropertyTypes != nil {
		var err error
		propertyTypesSpec, err = parsePropertyTypesSpecification(edgeTypePropertyTypes.PropertyTypesSpecification())
		if err != nil {
			return nil, nil, fmt.Errorf("error parsing edge type property types: %w", err)
		}
	}

	return parseLabelSetPhrase(impliedContent.EdgeTypeLabelSet().LabelSetPhrase()), propertyTypesSpec, nil
}
