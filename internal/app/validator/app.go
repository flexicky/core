package validator

import "github.com/go-playground/validator"

type Validator struct {
	validate *validator.Validate
}

func NewValidator() *Validator {
	return &Validator{
		validate: validator.New(),
	}
}

func (v *Validator) Validate(value any) error {
	return v.validate.Struct(value)
}
