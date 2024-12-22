package application

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"calc-go/pkg/calculation"
)

type Config struct {
	Addr string
}

// ConfigFromEnv initializes configuration from environment variables with a default value for PORT.
func ConfigFromEnv() *Config {
	addr := os.Getenv("PORT")
	if addr == "" {
		addr = "8080"
	}
	return &Config{Addr: addr}
}

type Application struct {
	config *Config
}

// New creates a new Application instance.
func New() *Application {
	return &Application{
		config: ConfigFromEnv(),
	}
}

type Request struct {
	Expression string `json:"expression"`
}

type Response struct {
	Result *float64 `json:"result,omitempty"`
	Error  *string  `json:"error,omitempty"`
}

// eval evaluates the mathematical expression using the calculation package.
func eval(expression string) (float64, error) {
	result, err := calculation.Calc(expression)
	if err != nil {
		if errors.Is(err, calculation.ErrDivisionByZero) {
			return 0, calculation.ErrDivisionByZero
		}
		return 0, calculation.ErrExpressionIsNotValid
	}
	return result, nil
}

// CalcHandler processes HTTP requests for calculating mathematical expressions.
func CalcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, "Expression is not valid", http.StatusUnprocessableEntity)
		return
	}

	result, err := eval(req.Expression)
	response := Response{}

	if err != nil {
		errMessage := err.Error()
		response.Error = &errMessage
		w.WriteHeader(http.StatusUnprocessableEntity)
	} else {
		response.Result = &result
		w.WriteHeader(http.StatusOK)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// writeErrorResponse sends a JSON error response.
func writeErrorResponse(w http.ResponseWriter, errorMessage string, statusCode int) {
	response := Response{
		Error: &errorMessage,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// RunServer starts the HTTP server for the application.
func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	return http.ListenAndServe(":"+a.config.Addr, nil)
}
