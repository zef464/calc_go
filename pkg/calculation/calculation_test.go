package calculation_test

import (
	"testing"

	"calc-go/pkg/calculation"
)

func TestCalc(t *testing.T) {
	testCasesSuccess := []struct {
		name           string
		expression     string
		expectedResult float64
	}{
		{
			name:           "simple",
			expression:     "1+1",
			expectedResult: 2,
		},
		{
			name:           "priority",
			expression:     "(2+2)*2",
			expectedResult: 8,
		},
		{
			name:           "priority",
			expression:     "2+2*2",
			expectedResult: 6,
		},
		{
			name:           "/",
			expression:     "1/2",
			expectedResult: 0.5,
		},
		{
			name:           "multiple operators",
			expression:     "2+3*4-5/5",
			expectedResult: 13,
		},
		{
			name:           "power operation simulated",
			expression:     "2*2*2",
			expectedResult: 8,
		},
		{
			name:           "fractional result",
			expression:     "10/4",
			expectedResult: 2.5,
		},
		{
			name:           "repeated addition",
			expression:     "1+1+1+1",
			expectedResult: 4,
		},
		{
			name:           "zero result",
			expression:     "5-5",
			expectedResult: 0,
		},
		{
			name:           "identity property",
			expression:     "0+123",
			expectedResult: 123,
		},
		{
			name:           "division identity",
			expression:     "123/1",
			expectedResult: 123,
		},
		{
			name:           "multiplication by zero",
			expression:     "123*0",
			expectedResult: 0,
		},
		{
			name:           "multiplication by one",
			expression:     "123*1",
			expectedResult: 123,
		},
		{
			name:           "complex negative",
			expression:     "-3*(2+4)/2",
			expectedResult: -9,
		},
		{
			name:           "negative",
			expression:     "-5-5",
			expectedResult: -10,
		},
	}

	for _, testCase := range testCasesSuccess {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := calculation.Calc(testCase.expression)
			if err != nil {
				t.Fatalf("successful case %s returns error", testCase.expression)
			}
			if val != testCase.expectedResult {
				t.Fatalf("%f should be equal %f", val, testCase.expectedResult)
			}
		})
	}

	testCasesFail := []struct {
		name        string
		expression  string
		expectedErr error
	}{
		{
			name:       "simple",
			expression: "1+1*",
		},
		{
			name:       "priority",
			expression: "2+2**2",
		},
		{
			name:       "priority",
			expression: "((2+2-*(2",
		},
		{
			name:       "/",
			expression: "",
		},
	}

	for _, testCase := range testCasesFail {
		t.Run(testCase.name, func(t *testing.T) {
			val, err := calculation.Calc(testCase.expression)
			if err == nil {
				t.Fatalf("expression %s is invalid but result  %f was obtained", testCase.expression, val)
			}
		})
	}
}
