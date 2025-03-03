package calculation

import (
	"strconv"
	"strings"
	"sync"
	"unicode"
)

func Calc(expression string) (float64, error) {
	expression = strings.ReplaceAll(expression, " ", "")
	if len(expression) == 0 {
		return 0, ErrExpressionIsNotValid
	}

	// Обрабатываем скобки
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

	// Запускаем параллельное вычисление
	result, err := parallelCompute(tokens)
	if err != nil {
		return 0, err
	}

	return result, nil
}

func parallelCompute(tokens []string) (float64, error) {
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Канал для передачи промежуточных результатов
	resultCh := make(chan []string, 1)
	resultCh <- tokens // Начальные токены

	// Обработаем умножение и деление
	wg.Add(1)
	go func() {
		defer wg.Done()

		mu.Lock()
		defer mu.Unlock()

		tokens := <-resultCh
		newTokens, err := computeMultiplicationAndDivision(tokens)
		if err != nil {
			resultCh <- nil
			return
		}
		resultCh <- newTokens
	}()

	wg.Wait() // Ждём выполнения первой фазы

	tokens = <-resultCh
	if tokens == nil {
		return 0, ErrExpressionIsNotValid
	}

	// Обработаем сложение и вычитание
	wg.Add(1)
	var finalResult float64
	var err error

	go func() {
		defer wg.Done()

		mu.Lock()
		defer mu.Unlock()

		finalResult, err = computeAdditionAndSubtraction(tokens)
	}()

	wg.Wait()

	return finalResult, err
}

func computeMultiplicationAndDivision(tokens []string) ([]string, error) {
	for i := 0; i < len(tokens); i++ {
		if tokens[i] == "*" || tokens[i] == "/" {
			if i == 0 || i+1 >= len(tokens) {
				return nil, ErrExpressionIsNotValid
			}

			left, err := strconv.ParseFloat(tokens[i-1], 64)
			if err != nil {
				return nil, ErrNotCorrectInput
			}
			right, err := strconv.ParseFloat(tokens[i+1], 64)
			if err != nil {
				return nil, ErrNotCorrectInput
			}
			if tokens[i] == "/" && right == 0 {
				return nil, ErrDivisionByZero
			}
			result := left * right
			if tokens[i] == "/" {
				result = left / right
			}

			tokens = append(tokens[:i-1], append([]string{strconv.FormatFloat(result, 'f', -1, 64)}, tokens[i+2:]...)...)
			i--
		}
	}
	return tokens, nil
}

func computeAdditionAndSubtraction(tokens []string) (float64, error) {
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
