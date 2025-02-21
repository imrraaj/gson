package gson

import (
	"fmt"
	"testing"
)

func TestJSONParser(t *testing.T) {
	testCases := []struct {
		name        string
		input       string
		expected    interface{} // Expected output type
		expectError bool
	}{
		{
			name:        "Empty Object",
			input:       "{}",
			expected:    map[string]interface{}{},
			expectError: false,
		},
		{
			name:        "Basic Object with One Key-Value Pair",
			input:       `{"key": "value"}`,
			expected:    map[string]interface{}{"key": "value"},
			expectError: false,
		},
		{
			name:        "Object with Number Value",
			input:       `{"key": 123}`,
			expected:    map[string]interface{}{"key": 123.0},
			expectError: false,
		},
		{
			name:        "Object with Boolean Value",
			input:       `{"key": true}`,
			expected:    map[string]interface{}{"key": true},
			expectError: false,
		},
		{
			name:        "Object with Null Value",
			input:       `{"key": null}`,
			expected:    map[string]interface{}{"key": nil},
			expectError: false,
		},
		{
			name:        "Object with Multiple Key-Value Pairs",
			input:       `{"key1": "value1", "key2": 123, "key3": true}`,
			expected:    map[string]interface{}{"key1": "value1", "key2": 123.0, "key3": true},
			expectError: false,
		},
		{
			name:        "Nested Object",
			input:       `{"key1": {"nestedKey": "nestedValue"}}`,
			expected:    map[string]interface{}{"key1": map[string]interface{}{"nestedKey": "nestedValue"}},
			expectError: false,
		},
		{
			name:        "Empty Array",
			input:       "[]",
			expected:    []interface{}{},
			expectError: false,
		},
		{
			name:        "Array with String Values",
			input:       `["value1", "value2", "value3"]`,
			expected:    []interface{}{"value1", "value2", "value3"},
			expectError: false,
		},
		{
			name:        "Array with Mixed Values",
			input:       `[123, "value", true, null]`,
			expected:    []interface{}{123.0, "value", true, nil},
			expectError: false,
		},
		{
			name:        "Array with Nested Object",
			input:       `[{"nestedKey": "nestedValue"}]`,
			expected:    []interface{}{map[string]interface{}{"nestedKey": "nestedValue"}},
			expectError: false,
		},
		{
			name:        "Invalid JSON - Missing Closing Brace",
			input:       `{"key": "value"`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Invalid JSON - Missing Colon",
			input:       `{"key" "value"}`,
			expected:    nil,
			expectError: true,
		},
		{
			name:        "Invalid JSON - Missing Comma",
			input:       `{"key1": "value1" "key2": "value2"}`,
			expected:    nil,
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result, err := Parse(tc.input)

			if tc.expectError {
				if err == nil {
					t.Fatalf("Expected error but got none for test case %s", tc.name)
				}
			} else {
				if err != nil {
					t.Fatalf("Unexpected error for test case %s : %v", tc.name, err)
				}

				if !compareResults(result, tc.expected) {
					t.Errorf("Results do not match for test case %s\nExpected:%v\nGot:%v", tc.name, tc.expected, result)
				}
			}
		})
	}
}

// compareResults checks if two interfaces are equal.
func compareResults(result, expected interface{}) bool {
	return fmt.Sprintf("%v", result) == fmt.Sprintf("%v", expected) // Simple comparison; replace with deep comparison if necessary
}
