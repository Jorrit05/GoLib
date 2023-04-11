package GoLib

import (
	"strings"
	"testing"
)

func TestGenerateGuid(t *testing.T) {
	testCases := []struct {
		name     string
		parts    int
		expected int
	}{
		{
			name:     "TestFullUUID",
			parts:    0,
			expected: 4,
		},
		{
			name:     "TestOnePart",
			parts:    1,
			expected: 0,
		},
		{
			name:     "TestTwoParts",
			parts:    2,
			expected: 1,
		},
		{
			name:     "TestThreeParts",
			parts:    3,
			expected: 2,
		},
		{
			name:     "TestOutOfRange",
			parts:    5,
			expected: 4,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			guid := GenerateGuid(tc.parts)
			dashCount := strings.Count(guid, "-")

			if dashCount != tc.expected {
				t.Errorf("Expected %d dashes but got %d for parts = %d", tc.expected, dashCount, tc.parts)
			}
		})
	}
}
