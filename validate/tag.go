package validate

import (
	validator "gopkg.in/go-playground/validator.v9"
)

func init() {
	V = validator.New()
}
