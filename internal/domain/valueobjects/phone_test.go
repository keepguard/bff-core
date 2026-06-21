package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewPhone(t *testing.T) {
	tests := []struct {
		name        string
		phone       string
		expectError bool
	}{
		{
			name:        "valid phone with 10 digits",
			phone:       "1199999999",
			expectError: false,
		},
		{
			name:        "valid phone with 11 digits",
			phone:       "11999999999",
			expectError: false,
		},
		{
			name:        "valid phone with formatting",
			phone:       "(11) 99999-9999",
			expectError: false,
		},
		{
			name:        "valid phone with spaces and dashes",
			phone:       "11 99999-9999",
			expectError: false,
		},
		{
			name:        "invalid phone - too short",
			phone:       "119999999",
			expectError: true,
		},
		{
			name:        "invalid phone - too long",
			phone:       "119999999999",
			expectError: true,
		},
		{
			name:        "invalid phone - empty",
			phone:       "",
			expectError: true,
		},
		{
			name:        "invalid phone - letters",
			phone:       "11abc99999",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			phone, err := NewPhone(tt.phone)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, Phone{}, phone)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, Phone{}, phone)
			}
		})
	}
}

func TestPhone_Value(t *testing.T) {
	phone, err := NewPhone("(11) 99999-9999")
	assert.NoError(t, err)
	assert.Equal(t, "11999999999", phone.Value())
}

func TestPhone_Formatted(t *testing.T) {
	phone11, err11 := NewPhone("11999999999")
	assert.NoError(t, err11)
	assert.Equal(t, "(11) 99999-9999", phone11.Formatted())

	phone10, err10 := NewPhone("1199999999")
	assert.NoError(t, err10)
	assert.Equal(t, "(11) 9999-9999", phone10.Formatted())
}

func TestPhone_String(t *testing.T) {
	phone, err := NewPhone("11999999999")
	assert.NoError(t, err)
	assert.Equal(t, "11999999999", phone.String())
}

func TestPhone_Equals(t *testing.T) {
	phone1, err1 := NewPhone("11999999999")
	phone2, err2 := NewPhone("11999999999")
	phone3, err3 := NewPhone("11888888888")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	assert.True(t, phone1.Equals(phone2))
	assert.False(t, phone1.Equals(phone3))
}

func TestPhone_AreaCode(t *testing.T) {
	phone, err := NewPhone("11999999999")
	assert.NoError(t, err)
	assert.Equal(t, "11", phone.AreaCode())
}

func TestPhone_Number(t *testing.T) {
	phone, err := NewPhone("11999999999")
	assert.NoError(t, err)
	assert.Equal(t, "999999999", phone.Number())
}
