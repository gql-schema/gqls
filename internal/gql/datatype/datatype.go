// Package datatype provides data type descriptors for GQL.
package datatype

// Descriptor describes a GQL data type.
// Based on chapter 4.16 of the GQL standard.
//
//sumtype:decl
type Descriptor interface {
	// IsNullable returns true if the data type is nullable, false otherwise.
	IsNullable() bool

	// sealed is a marker to ensure that all types are sealed.
	sealed()
}

// --- CHARACTER STRING ---

// CharacterStringDataTypeDescriptor describes a CHAR or VARCHAR data type.
type CharacterStringDataTypeDescriptor struct {
	MinLength int
	MaxLength int
	Nullable  bool
}

var _ Descriptor = (*CharacterStringDataTypeDescriptor)(nil)

// IsNullable returns true if the data type is nullable, false otherwise.
func (c *CharacterStringDataTypeDescriptor) IsNullable() bool { return c.Nullable }

func (c *CharacterStringDataTypeDescriptor) sealed() {}

// --- BYTE STRING ---

// ByteStringDataTypeDescriptor describes a BINARY or VARBINARY data type.
type ByteStringDataTypeDescriptor struct {
	MinLength int
	MaxLength int
	Nullable  bool
}

var _ Descriptor = (*ByteStringDataTypeDescriptor)(nil)

// IsNullable returns true if the data type is nullable, false otherwise.
func (b *ByteStringDataTypeDescriptor) IsNullable() bool { return b.Nullable }

func (b *ByteStringDataTypeDescriptor) sealed() {}

// --- NUMERIC ---

// NumericKind defines the kind of numeric data.
type NumericKind string

const (
	// NumericKindExact is for exact numeric data types.
	NumericKindExact NumericKind = "EXACT NUMERIC DATA"
	// NumericKindFloat is for approximate numeric data types.
	NumericKindFloat NumericKind = "FLOAT NUMERIC DATA"
)

// Radix defines the numeric base used.
type Radix string

const (
	// RadixBinary is for binary numeric data.
	RadixBinary Radix = "BINARY"
	// RadixDecimal is for decimal numeric data.
	RadixDecimal Radix = "DECIMAL"
)

// NumericDataTypeDescriptor describes a numeric data type.
type NumericDataTypeDescriptor struct {
	// Kind is the kind of numeric data.
	Kind NumericKind
	// Radix is the numeric base used.
	Radix Radix
	// Precision is a positive integer that determines the number of significant digits.
	Precision int // > 0
	// Scale is a non-negative integer such that S <= P. Values are scaled to N*10^(-S).
	Scale int // >= 0
	// Nullable is true if the data type is nullable, false otherwise.
	Nullable bool
}

var _ Descriptor = (*NumericDataTypeDescriptor)(nil)

// IsNullable returns true if the data type is nullable, false otherwise.
func (n *NumericDataTypeDescriptor) IsNullable() bool { return n.Nullable }

func (n *NumericDataTypeDescriptor) sealed() {}

// --- MISCELLANEOUS ---

// MiscellaneousDataTypeDescriptor describes a miscellaneous data type.
type MiscellaneousDataTypeDescriptor struct {
	Kind     string
	Nullable bool
}

var _ Descriptor = (*MiscellaneousDataTypeDescriptor)(nil)

// IsNullable returns true if the data type is nullable, false otherwise.
func (m *MiscellaneousDataTypeDescriptor) IsNullable() bool { return m.Nullable }

func (m *MiscellaneousDataTypeDescriptor) sealed() {}
