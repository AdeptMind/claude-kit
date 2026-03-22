package main

import "context"

// App is the main application struct for Wails bindings.
type App struct {
	ctx context.Context
}

// NewApp creates a new App instance.
func NewApp() *App {
	return &App{}
}

// startup is called when the Wails app starts.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// GetVersion returns the current application version.
func (a *App) GetVersion() string {
	return "0.1.0"
}
