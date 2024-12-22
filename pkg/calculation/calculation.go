package calculation

import (
	"strconv"
	"strings"
	"unicode"
)

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	if len(expression) == 0 {
		return 0, ErrExpressionIsNotValid
	}

	for strings.Contains(expression, "(") {
		start := strings.LastIndex(expression, "(")
		if start == -1 {
			return 0, ErrExpressionIsNotValid
		}
		end := strings.Index(expression[start:], ")")
		if end == -1 {
			return 0, ErrExpressionIsNotValid
		}
		end += start

		if start > len(expression) || end >= len(expression) {
			return 0, ErrExpressionIsNotValid
		}

		innerResult, err := Calc(expression[start+1 : end])
		if err != nil {
			return 0, err
		}

		expression = expression[:start] + strconv.FormatFloat(innerResult, 'f', -1, 64) + expression[end+1:]
	}

	expression = handleUnaryMinus(expression)

	tokens := tokenize(expression)
	if tokens == nil {
		return 0, ErrExpressionIsNotValid
	}

	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "*" || tokens[i] == "/" {
			if i == 0 || i+1 >= len(tokens) {
				return 0, ErrExpressionIsNotValid
			}

			left, err := strconv.ParseFloat(tokens[i-1], 64)
			if err != nil {
				return 0, ErrNotCorrectInput
			}
			right, err := strconv.ParseFloat(tokens[i+1], 64)
			if err != nil {
				return 0, ErrNotCorrectInput
			}
			if tokens[i] == "/" && right == 0 {
				return 0, ErrDivisionByZero
			}
			result := left * right
			if tokens[i] == "/" {
				result = left / right
			}

			tokens = append(tokens[:i-1], append([]string{strconv.FormatFloat(result, 'f', -1, 64)}, tokens[i+2:]...)...)
			i--
		}
	}

	result, err := strconv.ParseFloat(tokens[0], 64)
	if err != nil {
		return 0, ErrNotCorrectInput
	}
	for i := 1; i < len(tokens); i += 2 {
		if i+1 >= len(tokens) {
			return 0, ErrExpressionIsNotValid
		}

		right, err := strconv.ParseFloat(tokens[i+1], 64)
		if err != nil {
			return 0, ErrNotCorrectInput
		}
		if tokens[i] == "+" {
			result += right
		} else if tokens[i] == "-" {
			result -= right
		}
	}

	return result, nil
}

func tokenize(expression string) []string {
	var tokens []string
	var buffer strings.Builder

	for i, char := range expression {
		if unicode.IsDigit(char) || char == '.' {
			buffer.WriteRune(char)
		} else if isSign(char) {
			if buffer.Len() > 0 {
				tokens = append(tokens, buffer.String())
				buffer.Reset()
			}
			if char == '-' && (i == 0 || isSign(rune(expression[i-1])) || expression[i-1] == '(') {
				buffer.WriteRune(char)
			} else {
				tokens = append(tokens, string(char))
			}
		} else {
			return nil
		}
	}

	if buffer.Len() > 0 {
		tokens = append(tokens, buffer.String())
	}
	return tokens
}

func isSign(value rune) bool {
	return value == '+' || value == '-' || value == '*' || value == '/'
}

func handleUnaryMinus(expression string) string {
	var result strings.Builder
	for i, char := range expression {
		if char == '-' && (i == 0 || isSign(rune(expression[i-1])) || expression[i-1] == '(') {
			result.WriteString("0")
		}
		result.WriteRune(char)
	}
	return result.String()
}
