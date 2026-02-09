package plugins

import (
	"fmt"
	"strings"
)

// KeyStability represents the stability level of a key combination
type KeyStability int

const (
	// KeyStabilityStable - Key combination is completely reliable
	KeyStabilityStable KeyStability = iota
	// KeyStabilityMostlyStable - Key combination is generally reliable but may have edge cases
	KeyStabilityMostlyStable
	// KeyStabilityUnstable - Key combination is unreliable and should be avoided
	KeyStabilityUnstable
	// KeyStabilitySystemReserved - Key combination is reserved by the system
	KeyStabilitySystemReserved
)

// ValidationResult represents the result of validating a key combination
type ValidationResult struct {
	Valid      bool         // Whether the key combination is valid
	Stability  KeyStability // Stability level
	Warnings   []string     // Warning messages
	Errors     []string     // Error messages
	Suggestion string       // Suggested alternative combination
}

// KeyValidationConfig holds configuration for key validation
type KeyValidationConfig struct {
	Platform              string // Current platform (darwin, windows, linux)
	AllowSystemReserved   bool   // Whether to allow system-reserved shortcuts
	AllowUnstableKeys     bool   // Whether to allow unstable keys
	StrictMode            bool   // Whether to use strict validation
}

// DefaultKeyValidationConfig returns the default validation configuration
func DefaultKeyValidationConfig(platform string) *KeyValidationConfig {
	return &KeyValidationConfig{
		Platform:            platform,
		AllowSystemReserved: false,
		AllowUnstableKeys:   false,
		StrictMode:          true,
	}
}

// Stable keys that have been tested and verified as reliable
var stableKeys = map[string]bool{
	// All letter keys (A-Z) are stable
	"a": true, "b": true, "c": true, "d": true, "e": true,
	"f": true, "g": true, "h": true, "i": true, "j": true,
	"k": true, "l": true, "m": true, "n": true, "o": true,
	"p": true, "q": true, "r": true, "s": true, "t": true,
	"u": true, "v": true, "w": true, "x": true, "y": true, "z": true,
	// Number keys 0-5 are stable
	"0": true, "1": true, "2": true, "3": true, "4": true, "5": true,
	// Function keys are stable
	"f1": true, "f2": true, "f3": true, "f4": true, "f5": true,
	"f6": true, "f7": true, "f8": true, "f9": true, "f10": true,
	"f11": true, "f12": true, "f13": true, "f14": true, "f15": true,
	"f16": true, "f17": true, "f18": true, "f19": true,
	// Space, Enter, Tab, Escape are stable
	"space": true, "enter": true, "return": true, "tab": true, "escape": true, "esc": true,
}

// Unstable keys that have known issues
var unstableKeys = map[string]bool{
	"6": true, "7": true, "8": true, "9": true,
}

// System-reserved shortcuts on macOS
var systemReservedDarwin = map[string]string{
	"cmd+c":  "Copy",
	"cmd+v":  "Paste",
	"cmd+x":  "Cut",
	"cmd+z":  "Undo",
	"cmd+shift+z": "Redo",
	"cmd+a":  "Select All",
	"cmd+s":  "Save",
	"cmd+f":  "Find",
	"cmd+q":  "Quit Application",
	"cmd+w":  "Close Window",
	"cmd+h":  "Hide Application",
	"cmd+m":  "Minimize",
	"cmd+n":  "New Window/Document",
	"cmd+o":  "Open",
	"cmd+p":  "Print",
	"cmd+t":  "New Tab",
	"cmd+shift+t": "Reopen Closed Tab",
	"cmd+l":  "Focus Location Bar",
	"cmd+r":  "Reload",
	"cmd+shift+r": "Hard Reload",
	"cmd+u":  "View Source",
	"cmd+d":  "Bookmark",
	"cmd+i":  "Page Info",
	"cmd+j":  "Downloads/Information",
	"cmd+k":  "Search/Focus Search",
	"cmd+y":  "History/Redo",
	"cmd+0":  "Reset Zoom",
	"cmd+1":  "Navigate to Tab 1",
	"cmd+2":  "Navigate to Tab 2",
	"cmd+3":  "Navigate to Tab 3",
	"cmd+4":  "Navigate to Tab 4",
	"cmd+5":  "Navigate to Tab 5",
	"cmd+6":  "Navigate to Tab 6",
	"cmd+7":  "Navigate to Tab 7",
	"cmd+8":  "Navigate to Tab 8",
	"cmd+9":  "Navigate to Tab 9",
	"cmd+=": "Zoom In",
	"cmd+-": "Zoom Out",
	"cmd+[": "Back",
	"cmd+]": "Forward",
	"cmd+~": "Switch Window",
	"cmd+shift+~": "Switch Window Reverse",
	"cmd+?": "Help",
	"cmd+shift+a": "Applications",
	"cmd+option+esc": "Force Quit",
	"cmd+shift+3": "Screenshot",
	"cmd+shift+4": "Screenshot Selection",
	"cmd+shift+5": "Screenshot Options",
	"cmd+space": "Spotlight Search",
	"cmd+option+space": "Spotlight in Finder",
	"cmd+ctrl+space": "Character Viewer",
	"cmd+shift+option+esc": "Force Quit Front App",
}

// Suggested alternatives for unstable keys
var unstableKeyAlternatives = map[string]string{
	"cmd+6":  "cmd+f6",
	"cmd+7":  "cmd+f7",
	"cmd+8":  "cmd+f8",
	"cmd+9":  "cmd+f9",
	"cmd+shift+6": "cmd+shift+f6",
	"cmd+shift+7": "cmd+shift+f7",
	"cmd+shift+8": "cmd+shift+f8",
	"cmd+shift+9": "cmd+shift+f9",
}

// ValidateKeyCombo validates a key combination and returns detailed results
func ValidateKeyCombo(keyCombo string, config *KeyValidationConfig) *ValidationResult {
	result := &ValidationResult{
		Valid:     true,
		Stability: KeyStabilityStable,
		Warnings:  []string{},
		Errors:    []string{},
	}

	if config == nil {
		config = DefaultKeyValidationConfig("darwin")
	}

	// Normalize the key combination
	normalizedCombo := normalizeKeyCombo(keyCombo)
	parts := splitKeyCombo(normalizedCombo)

	// Check if combination is empty
	if len(parts) == 0 {
		result.Valid = false
		result.Errors = append(result.Errors, "key combination is empty")
		return result
	}

	// Extract the main key (non-modifier)
	var mainKey string
	var hasCmd bool
	for _, part := range parts {
		normalized := toLower(part)
		if isModifier(normalized) {
			if normalized == "cmd" || normalized == "command" || normalized == "meta" {
				hasCmd = true
			}
		} else {
			mainKey = normalized
		}
	}

	// Check if main key is unstable (only when Cmd is pressed)
	if hasCmd && unstableKeys[mainKey] {
		result.Valid = false
		result.Stability = KeyStabilityUnstable
		result.Errors = append(result.Errors, fmt.Sprintf(
			"key '%s' is unstable when combined with Cmd modifier on %s",
			mainKey, config.Platform))

		// Provide alternative suggestion
		normalizedComboLower := toLower(normalizedCombo)
		if alt, exists := unstableKeyAlternatives[normalizedComboLower]; exists {
			result.Suggestion = alt
		} else {
			// Generate a generic alternative
			result.Suggestion = fmt.Sprintf("cmd+%s", strings.ToUpper(mainKey))
		}

		return result
	}

	// Check if main key is stable
	if hasCmd && !stableKeys[mainKey] {
		result.Warnings = append(result.Warnings, fmt.Sprintf(
			"key '%s' has not been extensively tested with Cmd modifier",
			mainKey))
		result.Stability = KeyStabilityMostlyStable
	}

	// Check for system-reserved shortcuts
	if hasCmd {
		normalizedComboLower := toLower(normalizedCombo)
		if systemFunction, isReserved := systemReservedDarwin[normalizedComboLower]; isReserved {
			if !config.AllowSystemReserved {
				result.Warnings = append(result.Warnings, fmt.Sprintf(
					"Cmd+%s is reserved by the system for: %s",
					mainKey, systemFunction))
				result.Stability = KeyStabilitySystemReserved
			}
		}
	}

	// Validate modifier key combinations
	modifierCount := 0
	for _, part := range parts {
		if isModifier(toLower(part)) {
			modifierCount++
		}
	}

	// Warn about complex modifier combinations
	if modifierCount >= 3 {
		result.Warnings = append(result.Warnings,
			"key combination uses 3 or more modifiers - may be difficult to use")
		result.Stability = KeyStabilityMostlyStable
	}

	// Check for Shift+Letter combinations (redundant)
	if modifierCount == 1 && hasShiftModifier(parts) && isLetterKey(mainKey) {
		result.Warnings = append(result.Warnings,
			"Shift+Letter combinations are redundant - use uppercase letter directly")
	}

	// Check for problematic combinations on specific platforms
	if config.Platform == "darwin" {
		result = validateDarwinSpecific(normalizedCombo, result)
	}

	return result
}

// isModifier checks if a key is a modifier key
func isModifier(key string) bool {
	switch key {
	case "ctrl", "control", "shift", "alt", "option", "cmd", "command", "meta":
		return true
	default:
		return false
	}
}

// isLetterKey checks if a key is a letter key (A-Z)
func isLetterKey(key string) bool {
	if len(key) != 1 {
		return false
	}
	c := key[0]
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// hasShiftModifier checks if the combination includes Shift
func hasShiftModifier(parts []string) bool {
	for _, part := range parts {
		normalized := toLower(part)
		if normalized == "shift" {
			return true
		}
	}
	return false
}

// validateDarwinSpecific performs macOS-specific validation
func validateDarwinSpecific(keyCombo string, result *ValidationResult) *ValidationResult {
	parts := splitKeyCombo(keyCombo)

	// Check for F1-F4 which have system functions
	for _, part := range parts {
		normalized := toLower(part)
		if normalized == "f1" || normalized == "f2" ||
		   normalized == "f3" || normalized == "f4" {
			result.Warnings = append(result.Warnings,
				fmt.Sprintf("%s is reserved for system functions (brightness, Mission Control)",
				strings.ToUpper(normalized)))
			if result.Stability > KeyStabilitySystemReserved {
				result.Stability = KeyStabilitySystemReserved
			}
		}
	}

	return result
}

// GetStabilityDescription returns a human-readable description of stability level
func GetStabilityDescription(stability KeyStability) string {
	switch stability {
	case KeyStabilityStable:
		return "完全稳定 - 此快捷键组合经过测试，可以可靠使用"
	case KeyStabilityMostlyStable:
		return "基本稳定 - 此快捷键组合通常可用，但可能存在边缘情况"
	case KeyStabilityUnstable:
		return "不稳定 - 此快捷键组合存在已知问题，应避免使用"
	case KeyStabilitySystemReserved:
		return "系统保留 - 此快捷键组合被操作系统保留，可能与系统功能冲突"
	default:
		return "未知"
	}
}

// GetRecommendedShortcuts returns a list of recommended shortcut combinations
func GetRecommendedShortcuts(count int, platform string) []string {
	if platform != "darwin" {
		// For non-macOS platforms, return basic recommendations
		return getBasicRecommendations(count)
	}

	// macOS-specific recommendations
	recommendations := []string{
		// Letter keys (most stable)
		"cmd+d", "cmd+j", "cmd+k", "cmd+l", "cmd+u",
		"cmd+shift+d", "cmd+shift+j", "cmd+shift+k",
		// Function keys (avoid F1-F4 system functions)
		"cmd+f5", "cmd+f6", "cmd+f7", "cmd+f8", "cmd+f9",
		"cmd+shift+f5", "cmd+shift+f6",
		// Number keys 0-5 (stable range)
		"cmd+0", "cmd+1", "cmd+2", "cmd+3", "cmd+4", "cmd+5",
	}

	if count > len(recommendations) {
		count = len(recommendations)
	}

	return recommendations[:count]
}

// getBasicRecommendations returns basic recommendations for non-macOS platforms
func getBasicRecommendations(count int) []string {
	recommendations := []string{
		"ctrl+d", "ctrl+j", "ctrl+k", "ctrl+l",
		"ctrl+shift+d", "ctrl+shift+j",
		"f5", "f6", "f7", "f8", "f9",
	}

	if count > len(recommendations) {
		count = len(recommendations)
	}

	return recommendations[:count]
}

// FormatValidationResult formats a validation result for display
func FormatValidationResult(result *ValidationResult) string {
	if result.Valid {
		var output string
		output += "✓ 快捷键有效\n"
		output += fmt.Sprintf("稳定性: %s\n", GetStabilityDescription(result.Stability))

		if len(result.Warnings) > 0 {
			output += "\n警告:\n"
			for _, warning := range result.Warnings {
				output += fmt.Sprintf("  ⚠ %s\n", warning)
			}
		}

		return output
	} else {
		var output string
		output += "✗ 快捷键无效\n"

		if len(result.Errors) > 0 {
			output += "\n错误:\n"
			for _, err := range result.Errors {
				output += fmt.Sprintf("  ✗ %s\n", err)
			}
		}

		if result.Suggestion != "" {
			output += fmt.Sprintf("\n建议替代: %s\n", result.Suggestion)
		}

		return output
	}
}
