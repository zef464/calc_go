package calculation

import "errors"

var (
	ErrExpressionIsNotValid = errors.New("Expression is not valid")
	ErrDivisionByZero       = errors.New("Division by zero")
	ErrInternalServerError  = errors.New("Internal server error")
	ErrNotCorrectInput      = errors.New("Not correct input")
)
