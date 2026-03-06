package update

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		name     string
		v1       string
		v2       string
		expected int
	}{
		// 相等版本
		{"equal versions", "0.1.0", "0.1.0", 0},
		{"equal with v prefix", "v0.1.0", "0.1.0", 0},
		{"equal both with v", "v0.1.0", "v0.1.0", 0},

		// v1 > v2
		{"major version higher", "1.0.0", "0.1.0", 1},
		{"minor version higher", "0.2.0", "0.1.0", 1},
		{"patch version higher", "0.1.1", "0.1.0", 1},
		{"v prefix comparison", "v0.2.0", "0.1.0", 1},

		// v1 < v2
		{"major version lower", "0.1.0", "1.0.0", -1},
		{"minor version lower", "0.1.0", "0.2.0", -1},
		{"patch version lower", "0.1.0", "0.1.1", -1},

		// 缺少部分版本号
		{"missing patch", "0.1", "0.1.0", 0},
		{"missing minor and patch", "1", "1.0.0", 0},
		{"compare incomplete versions", "0.2", "0.1.9", 1},

		// 带后缀的版本号
		{"with beta suffix", "0.2.0-beta", "0.1.0", 1},
		{"with rc suffix", "0.2.0-rc1", "0.1.0", 1},
		{"with build metadata", "0.2.0+build123", "0.1.0", 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareVersions(tt.v1, tt.v2)
			if result != tt.expected {
				t.Errorf("compareVersions(%s, %s) = %d, expected %d", tt.v1, tt.v2, result, tt.expected)
			}
		})
	}
}

func TestParseVersionPart(t *testing.T) {
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"1", 1, false},
		{"10", 10, false},
		{"0", 0, false},
		{"1-beta", 1, false},
		{"2-rc1", 2, false},
		{"3+build", 3, false},
		{"abc", 0, true},
		{"", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result, err := parseVersionPart(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("parseVersionPart(%s) expected error, got none", tt.input)
				}
			} else {
				if err != nil {
					t.Errorf("parseVersionPart(%s) unexpected error: %v", tt.input, err)
				}
				if result != tt.expected {
					t.Errorf("parseVersionPart(%s) = %d, expected %d", tt.input, result, tt.expected)
				}
			}
		})
	}
}
