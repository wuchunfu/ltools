package calculator

import (
	"github.com/wailsapp/wails/v3/pkg/application"
)

// CalculatorService exposes Calculator functionality to the frontend
type CalculatorService struct {
	app    *application.App
	plugin *CalculatorPlugin
}

// NewCalculatorService creates a new calculator service
func NewCalculatorService(plugin *CalculatorPlugin, app *application.App) *CalculatorService {
	return &CalculatorService{
		app:    app,
		plugin: plugin,
	}
}

// ServiceStartup is called when the application starts
func (s *CalculatorService) ServiceStartup(app *application.App) error {
	// Initialize calculator plugin
	return nil
}

// ServiceShutdown is called when the application shuts down
func (s *CalculatorService) ServiceShutdown(app *application.App) error {
	// Cleanup calculator plugin
	return nil
}

// Add adds two numbers
func (s *CalculatorService) Add(a, b float64) float64 {
	return s.plugin.Add(a, b)
}

// Subtract subtracts b from a
func (s *CalculatorService) Subtract(a, b float64) float64 {
	return s.plugin.Subtract(a, b)
}

// Multiply multiplies two numbers
func (s *CalculatorService) Multiply(a, b float64) float64 {
	return s.plugin.Multiply(a, b)
}

// Divide divides a by b
func (s *CalculatorService) Divide(a, b float64) (float64, error) {
	return s.plugin.Divide(a, b)
}

// Evaluate evaluates a mathematical expression
func (s *CalculatorService) Evaluate(expression string) (float64, error) {
	return s.plugin.Evaluate(expression)
}

// Percentage calculates percentage
func (s *CalculatorService) Percentage(part, total float64) float64 {
	return s.plugin.Percentage(part, total)
}

// ClearHistory clears calculation history
func (s *CalculatorService) ClearHistory() {
	s.plugin.ClearHistory()
}

// GetLastResult returns the last calculation result
func (s *CalculatorService) GetLastResult() float64 {
	return s.plugin.GetLastResult()
}
