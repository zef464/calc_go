package main

import (
	"calc-go/internal/application"
)

func main() {
	app := application.New()
	// app.Run()
	app.RunServer()
}
