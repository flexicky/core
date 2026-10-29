// Package validator is a thin wrapper around go-playground/validator so the
// transport layer depends on a small, mockable type.
package validator

import "github.com/go-playground/validator"

// Validator validates structs annotated with `validate` tags.
type Validator struct {
	validate *validator.Validate
}

func New() *Validator {
	return &Validator{validate: validator.New()}
}

// Validate reports the first violated constraint of value.
func (v *Validator) Validate(value any) error {
	return v.validate.Struct(value)
}
