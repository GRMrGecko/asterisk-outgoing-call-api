package main

import (
	"math/rand"
	"time"
)

// App is the standard structure that allows different parts of the application to access common parameters/configuration.
type App struct {
	flags      *Flags
	httpServer *HTTPServer
	config     Config
}

var app *App

func main() {
	// We use rand for file naming, best set seed at start.
	rand.Seed(time.Now().UnixNano())

	app = new(App)
	app.flags = new(Flags)
	app.flags.Init()
	app.ReadConfig()

	HTTPServe()
}
