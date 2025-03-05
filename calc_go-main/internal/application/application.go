package application

import (
	"calc-go/pkg/calculation"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/handlers"
)

type Expression struct {
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	Result         *float64 `json:"result"`
	Tasks          []Task   `json:"-"`
	ProcessingTask *Task    `json:"-"` // Новое поле для текущей выполняемой задачи
}

type Task struct {
	NumTask       int    `json:"task"`
	Arg1          string `json:"arg1"`
	Arg2          string `json:"arg2"`
	Op            string `json:"operation"`
	OperationTime int    `json:"operation_time"`
	Status        string `json:"status"` // Новое поле!
}

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
		log.Println("Error in calculation:", err)
		if errors.Is(err, calculation.ErrDivisionByZero) {
			return 0, calculation.ErrDivisionByZero
		}
		return 0, calculation.ErrExpressionIsNotValid
	}
	return result, nil
}

// CalcHandler processes HTTP requests for calculating mathematical expressions.
var (
	expressions = make(map[string]*Expression) // Теперь ключи — UUID
	mu          sync.Mutex
)

func CalcHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	var req Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println("Error decoding JSON:", err)
		writeJSONResponse(w, "Expression is not valid", http.StatusUnprocessableEntity)
		return
	}

	log.Println("Received expression:", req.Expression)

	// Генерируем UUID
	exprID := uuid.New().String()

	// Парсим выражение
	tasks, err := parseExpression(req.Expression)
	if err != nil {
		http.Error(w, "Failed to parse expression", http.StatusBadRequest)
		return
	}

	// Создаём объект выражения со статусом "pending"
	expr := &Expression{
		ID:     exprID,
		Status: "pending",
		Tasks:  tasks,
	}

	// Сохраняем выражение в список
	mu.Lock()
	expressions[exprID] = expr
	mu.Unlock()

	// Отправляем ответ сразу, не дожидаясь вычисления
	writeJSONResponse(w, map[string]interface{}{
		"id": exprID,
	}, http.StatusOK)

	// Фоновый расчёт с задержкой
	go func() {
		time.Sleep(10 * time.Second) // Имитация обработки (10 секунд)

		result, err := eval(req.Expression)

		mu.Lock()
		defer mu.Unlock()

		if err != nil {
			log.Println("Error evaluating expression:", err)
			expr.Status = "error"
			return
		}

		expr.Result = &result
		expr.Status = "completed"
		log.Printf("Expression %s completed: %f", expr.ID, result)
	}()
}

// writeJSONResponse sends a JSON response.
func writeJSONResponse(w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// parseExpression разбивает выражение на составляющие, возвращая задачи для дальнейшего выполнения
func parseExpression(expression string) ([]Task, error) {
	expression = strings.TrimSpace(expression)

	var operator string
	var operatorIndex int
	for i, ch := range expression {
		if ch == '+' || ch == '-' || ch == '*' || ch == '/' {
			operator = string(ch)
			operatorIndex = i
			break
		}
	}

	if operator == "" {
		return nil, errors.New("invalid expression format: operator not found")
	}

	arg1 := expression[:operatorIndex]
	arg2 := expression[operatorIndex+1:]

	if arg1 == "" || arg2 == "" {
		return nil, errors.New("invalid expression format: missing operand(s)")
	}

	mu.Lock()
	taskCounter++ // Увеличиваем глобальный счётчик
	taskNum := taskCounter
	mu.Unlock()

	task := Task{
		NumTask:       taskNum, // Теперь нумерация задач корректная
		Arg1:          arg1,
		Op:            operator,
		Arg2:          arg2,
		OperationTime: 10,
	}

	return []Task{task}, nil
}

// all expressions
func GetAllExpressionsHandler(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	log.Println("Fetching all expressions. Total count:", len(expressions))

	var list []Expression
	for _, expr := range expressions {
		log.Printf("Expression ID: %s, Status: %s, Result: %v", expr.ID, expr.Status, expr.Result)
		list = append(list, Expression{
			ID:     expr.ID,
			Status: expr.Status,
			Result: expr.Result,
		})
	}

	// Формируем JSON-ответ
	response := map[string]interface{}{
		"expressions": list,
	}

	// Отправляем JSON-ответ
	writeJSONResponse(w, response, http.StatusOK)
}

// get expression from id
func GetExpressionByIDHandler(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/expressions/")

	mu.Lock()
	expr, exists := expressions[id]
	mu.Unlock()

	if !exists {
		http.Error(w, "Not found", http.StatusNotFound)
		return
	}

	response := map[string]interface{}{
		"expression": expr,
	}

	writeJSONResponse(w, response, http.StatusOK)
}

// print tasks
var taskCounter int // Глобальный счётчик задач

func AddNewTask(expression *Expression, task Task) {
	mu.Lock()
	defer mu.Unlock()

	taskCounter++              // Глобальный счётчик задач
	task.NumTask = taskCounter // Каждой новой задаче даём уникальный номер
	expression.Tasks = append(expression.Tasks, task)
}

func TaskHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		mu.Lock()
		defer mu.Unlock()

		var availableTasks []Task

		// Собираем все задачи со статусом "pending"
		for _, expr := range expressions {
			for i := range expr.Tasks {
				if expr.Tasks[i].Status == "" { // Если статус не установлен, считаем "pending"
					expr.Tasks[i].Status = "processing" // Фиксируем выполнение
					availableTasks = append(availableTasks, expr.Tasks[i])
				}
			}
		}

		if len(availableTasks) == 0 {
			http.Error(w, "No tasks", http.StatusNotFound)
			return
		}

		// Логируем задачи
		for _, task := range availableTasks {
			log.Printf("Processing task: %d", task.NumTask)
		}

		// Формируем JSON-ответ
		writeJSONResponse(w, map[string]interface{}{
			"tasks": availableTasks,
		}, http.StatusOK)

	case http.MethodPost:
		var res struct {
			ID     int     `json:"id"`
			Result float64 `json:"result"`
		}
		if err := json.NewDecoder(r.Body).Decode(&res); err != nil {
			http.Error(w, "Invalid request", http.StatusUnprocessableEntity)
			return
		}

		mu.Lock()
		defer mu.Unlock()

		var expr *Expression
		var task *Task

		// Ищем задачу по ID
		for _, e := range expressions {
			for i := range e.Tasks {
				if e.Tasks[i].NumTask == res.ID {
					expr = e
					task = &e.Tasks[i]
					break
				}
			}
			if task != nil {
				break
			}
		}

		if task == nil || task.Status != "processing" {
			http.Error(w, "Task not found or not in processing", http.StatusNotFound)
			return
		}

		// Имитация обработки задачи
		log.Printf("Task %d is being processed...", res.ID)
		time.Sleep(time.Duration(task.OperationTime) * time.Second)

		// Обновляем статус задачи
		task.Status = "completed"

		// Проверяем, все ли задачи выполнены
		allCompleted := true
		for _, t := range expr.Tasks {
			if t.Status != "completed" {
				allCompleted = false
				break
			}
		}
		if allCompleted {
			expr.Status = "completed"
		}

		// Логируем успешное завершение
		log.Printf("Task %d completed with result: %f", res.ID, res.Result)

		w.WriteHeader(http.StatusOK)

	default:
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
	}
}

// RunServer starts the HTTP server for the application.
func (a *Application) RunServer() error {
	http.HandleFunc("/api/v1/calculate", CalcHandler)
	http.HandleFunc("/api/v1/expressions", GetAllExpressionsHandler)
	http.HandleFunc("/api/v1/expressions/{id}", GetExpressionByIDHandler)
	http.HandleFunc("/api/v1/internal/task", TaskHandler)

	// Настроим CORS, чтобы разрешить запросы с фронтенда
	return http.ListenAndServe(":"+a.config.Addr, handlers.CORS(
		handlers.AllowedOrigins([]string{"http://localhost:3000"}),                   // Разрешаем только запросы с фронтенда
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}), // Разрешаем нужные методы
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),           // Разрешаем эти заголовки
	)(http.DefaultServeMux))
}
