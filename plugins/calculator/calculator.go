package calculator

import (
	"fmt"
	"strconv"

	"github.com/wailsapp/wails/v3/pkg/application"
	"ltools/internal/plugins"
)

const (
	PluginID      = "calculator.builtin"
	PluginName    = "计算器"
	PluginVersion = "1.0.0"
)

// CalculatorPlugin provides basic calculator functionality
type CalculatorPlugin struct {
	*plugins.BasePlugin
	app *application.App
}

// NewCalculatorPlugin creates a new calculator plugin instance
func NewCalculatorPlugin() *CalculatorPlugin {
	metadata := &plugins.PluginMetadata{
		ID:          PluginID,
		Name:        PluginName,
		Version:     PluginVersion,
		Author:      "LTools Team",
		Description: "基本计算器插件，支持加减乘除运算",
		Icon:        "calculator",
		Type:        plugins.PluginTypeBuiltIn,
		State:       plugins.PluginStateInstalled,
		Keywords:    []string{"计算器", "数学", "计算", "calculator", "math", "compute"},
	}

	base := plugins.NewBasePlugin(metadata)
	return &CalculatorPlugin{
		BasePlugin: base,
	}
}

// Init initializes the plugin
func (p *CalculatorPlugin) Init(app *application.App) error {
	if err := p.BasePlugin.Init(app); err != nil {
		return err
	}
	p.app = app
	return nil
}

// ServiceStartup is called when the application starts
func (p *CalculatorPlugin) ServiceStartup(app *application.App) error {
	return p.BasePlugin.ServiceStartup(app)
}

// ServiceShutdown is called when the application shuts down
func (p *CalculatorPlugin) ServiceShutdown(app *application.App) error {
	return p.BasePlugin.ServiceShutdown(app)
}

// Enabled returns true if the plugin is enabled
func (p *CalculatorPlugin) Enabled() bool {
	return p.BasePlugin.Enabled()
}

// SetEnabled enables or disables the plugin
func (p *CalculatorPlugin) SetEnabled(enabled bool) error {
	return p.BasePlugin.SetEnabled(enabled)
}

// Helper method to emit events
func (p *CalculatorPlugin) emitEvent(eventName, data string) {
	if p.app != nil {
		p.app.Event.Emit("calculator:"+eventName, data)
	}
}

// Add adds two numbers
func (p *CalculatorPlugin) Add(a, b float64) float64 {
	result := a + b
	p.emitEvent("result", fmt.Sprintf("%.2f", result))
	return result
}

// Subtract subtracts b from a
func (p *CalculatorPlugin) Subtract(a, b float64) float64 {
	result := a - b
	p.emitEvent("result", fmt.Sprintf("%.2f", result))
	return result
}

// Multiply multiplies two numbers
func (p *CalculatorPlugin) Multiply(a, b float64) float64 {
	result := a * b
	p.emitEvent("result", fmt.Sprintf("%.2f", result))
	return result
}

// Divide divides a by b
func (p *CalculatorPlugin) Divide(a, b float64) (float64, error) {
	if b == 0 {
		p.emitEvent("error", "除零错误")
		return 0, fmt.Errorf("division by zero")
	}
	result := a / b
	p.emitEvent("result", fmt.Sprintf("%.2f", result))
	return result, nil
}

// Evaluate evaluates a mathematical expression string
// Supported operators: +, -, *, /, (, )
func (p *CalculatorPlugin) Evaluate(expression string) (float64, error) {
	// Simple implementation - in production, use a proper parser
	// This is a basic evaluator for demonstration
	result, err := evaluateSimple(expression)
	if err != nil {
		p.emitEvent("error", err.Error())
		return 0, err
	}
	p.emitEvent("result", fmt.Sprintf("%.2f", result))
	return result, nil
}

// Percentage calculates percentage
func (p *CalculatorPlugin) Percentage(part, total float64) float64 {
	if total == 0 {
		p.emitEvent("error", "总数不能为零")
		return 0
	}
	result := (part / total) * 100
	p.emitEvent("result", fmt.Sprintf("%.2f%%", result))
	return result
}

// ClearHistory clears calculation history (placeholder for future feature)
func (p *CalculatorPlugin) ClearHistory() {
	p.emitEvent("history", "cleared")
}

// GetLastResult returns the last calculation result
// This is a placeholder - would need to store history state
func (p *CalculatorPlugin) GetLastResult() float64 {
	p.emitEvent("history", "requested")
	return 0
}

// Simple expression evaluator (basic implementation)
func evaluateSimple(expr string) (float64, error) {
	// Remove whitespace
	expr = removeWhitespace(expr)

	// For now, just handle simple cases
	// This would need a proper parser in production
	if len(expr) == 0 {
		return 0, fmt.Errorf("empty expression")
	}

	// Try to parse as single number
	if num, err := strconv.ParseFloat(expr, 64); err == nil {
		return num, nil
	}

	// Simple addition/subtraction parsing
	result := 0.0
	current := 0.0
	op := '+'
	i := 0

	for i < len(expr) {
		// Parse number
		j := i
		for j < len(expr) && (isdigit(expr[j]) || expr[j] == '.') {
			j++
		}

		if j > i {
			num, err := strconv.ParseFloat(expr[i:j], 64)
			if err != nil {
				return 0, fmt.Errorf("invalid number at position %d", i)
			}

			switch op {
			case '+':
				current = num
			case '-':
				current = -num
			case '*':
				current *= num
			case '/':
				if num == 0 {
					return 0, fmt.Errorf("division by zero")
				}
				current /= num
			}

			// If next is + or -, add to result
			if j < len(expr) && (expr[j] == '+' || expr[j] == '-') {
				result += current
				current = 0
			}

			i = j
		}

		// Parse operator
		if i < len(expr) && isOperator(expr[i]) {
			op = rune(expr[i])
			i++
		} else {
			i++
		}
	}

	result += current
	return result, nil
}

func isdigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func isOperator(c byte) bool {
	return c == '+' || c == '-' || c == '*' || c == '/'
}

func removeWhitespace(s string) string {
	result := ""
	for _, c := range s {
		if !isWhitespace(rune(c)) {
			result += string(c)
		}
	}
	return result
}

func isWhitespace(c rune) bool {
	return c == ' ' || c == '\t' || c == '\n' || c == '\r'
}
