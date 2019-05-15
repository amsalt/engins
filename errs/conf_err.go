package errs

import "strconv"

// UnsupportedType represents an unsupported type error.
type UnsupportedType struct {
	Type string
}

func NewUnsupportedType(t string) *UnsupportedType {
	return &UnsupportedType{Type: t}
}

func (e *UnsupportedType) Error() string {
	return "Unsupported type: " + e.Type
}

// FieldMismatch represents a csv field mismatch error.
type FieldMismatch struct {
	expected, found int
}

func NewFieldMismatch(expected, found int) *FieldMismatch {
	return &FieldMismatch{expected: expected, found: found}
}

func (e *FieldMismatch) Error() string {
	return "CSV fields mismatch. expect " + strconv.Itoa(e.expected) + " ,found " + strconv.Itoa(e.found)
}
