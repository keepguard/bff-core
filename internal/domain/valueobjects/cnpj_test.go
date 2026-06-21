package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewCNPJ(t *testing.T) {
	tests := []struct {
		name        string
		cnpj        string
		expectError bool
	}{
		{
			name:        "Valid CNPJ should create successfully",
			cnpj:        "11222333000181",
			expectError: false,
		},
		{
			name:        "Valid CNPJ with formatting should create successfully",
			cnpj:        "11.222.333/0001-81",
			expectError: false,
		},
		{
			name:        "Invalid CNPJ should return error",
			cnpj:        "12345678000198",
			expectError: true,
		},
		{
			name:        "Empty CNPJ should return error",
			cnpj:        "",
			expectError: true,
		},
		{
			name:        "CNPJ with wrong length should return error",
			cnpj:        "123456789",
			expectError: true,
		},
		{
			name:        "CNPJ with letters should return error",
			cnpj:        "1234567800019a",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			cnpj, err := NewCNPJ(tt.cnpj)

			// Assert
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, cnpj.Value())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, cnpj.Value())
			}
		})
	}
}

func TestCNPJ_Value(t *testing.T) {
	// Arrange
	cnpj, err := NewCNPJ("11222333000181")

	// Act & Assert
	assert.NoError(t, err)
	assert.Equal(t, "11222333000181", cnpj.Value())
}

func TestCNPJ_Formatted(t *testing.T) {
	// Arrange
	cnpj, err := NewCNPJ("11222333000181")

	// Act & Assert
	assert.NoError(t, err)
	assert.Equal(t, "11.222.333/0001-81", cnpj.Formatted())
}

func TestCNPJ_String(t *testing.T) {
	// Arrange
	cnpj, err := NewCNPJ("11222333000181")

	// Act & Assert
	assert.NoError(t, err)
	assert.Equal(t, "11222333000181", cnpj.String())
}

func TestCNPJ_Equals(t *testing.T) {
	tests := []struct {
		name     string
		cnpj1    string
		cnpj2    string
		expected bool
	}{
		{
			name:     "Same CNPJs should be equal",
			cnpj1:    "11222333000181",
			cnpj2:    "11222333000181",
			expected: true,
		},
		{
			name:     "Same CNPJs with different formatting should be equal",
			cnpj1:    "11222333000181",
			cnpj2:    "11.222.333/0001-81",
			expected: true,
		},
		{
			name:     "Different CNPJs should not be equal",
			cnpj1:    "11222333000181",
			cnpj2:    "00000000000191",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			cnpj1, err1 := NewCNPJ(tt.cnpj1)
			cnpj2, err2 := NewCNPJ(tt.cnpj2)

			// Act & Assert
			assert.NoError(t, err1)
			assert.NoError(t, err2)
			assert.Equal(t, tt.expected, cnpj1.Equals(cnpj2))
		})
	}
}

func TestCNPJ_CleanCNPJ(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CNPJ with dots and dashes should be cleaned",
			input:    "12.345.678/0001-99",
			expected: "12345678000199",
		},
		{
			name:     "CNPJ with spaces should be cleaned",
			input:    "12 345 678 0001 99",
			expected: "12345678000199",
		},
		{
			name:     "CNPJ with mixed separators should be cleaned",
			input:    "12.345.678/0001-99",
			expected: "12345678000199",
		},
		{
			name:     "Clean CNPJ should remain unchanged",
			input:    "12345678000199",
			expected: "12345678000199",
		},
		{
			name:     "Empty CNPJ should remain empty",
			input:    "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := cleanCNPJ(tt.input)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCNPJ_IsValidCNPJ(t *testing.T) {
	tests := []struct {
		name     string
		cnpj     string
		expected bool
	}{
		{
			name:     "Valid CNPJ should return true",
			cnpj:     "11222333000181",
			expected: true,
		},
		{
			name:     "Invalid CNPJ should return false",
			cnpj:     "11222333000180",
			expected: false,
		},
		{
			name:     "Empty CNPJ should return false",
			cnpj:     "",
			expected: false,
		},
		{
			name:     "CNPJ with wrong length should return false",
			cnpj:     "123456789",
			expected: false,
		},
		{
			name:     "CNPJ with letters should return false",
			cnpj:     "1234567800019a",
			expected: false,
		},
		{
			name:     "CNPJ with all same digits should return false",
			cnpj:     "11111111111111",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Act
			result := isValidCNPJ(tt.cnpj)

			// Assert
			assert.Equal(t, tt.expected, result)
		})
	}
}
