package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "valid email",
			email:       "user@example.com",
			expectError: false,
		},
		{
			name:        "valid email with subdomain",
			email:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "valid email with numbers",
			email:       "user123@example123.com",
			expectError: false,
		},
		{
			name:        "invalid email - no @",
			email:       "userexample.com",
			expectError: true,
		},
		{
			name:        "invalid email - no domain",
			email:       "user@",
			expectError: true,
		},
		{
			name:        "invalid email - no local part",
			email:       "@example.com",
			expectError: true,
		},
		{
			name:        "invalid email - empty",
			email:       "",
			expectError: true,
		},
		{
			name:        "invalid email - spaces",
			email:       "user @example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			email, err := NewEmail(tt.email)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, Email{}, email)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, Email{}, email)
			}
		})
	}
}

func TestEmail_Value(t *testing.T) {
	email, err := NewEmail("User@Example.COM")
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", email.Value())
}

func TestEmail_Domain(t *testing.T) {
	email, err := NewEmail("user@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "example.com", email.Domain())
}

func TestEmail_LocalPart(t *testing.T) {
	email, err := NewEmail("user@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "user", email.LocalPart())
}

func TestEmail_String(t *testing.T) {
	email, err := NewEmail("user@example.com")
	assert.NoError(t, err)
	assert.Equal(t, "user@example.com", email.String())
}

func TestEmail_Equals(t *testing.T) {
	email1, err1 := NewEmail("user@example.com")
	email2, err2 := NewEmail("user@example.com")
	email3, err3 := NewEmail("other@example.com")

	assert.NoError(t, err1)
	assert.NoError(t, err2)
	assert.NoError(t, err3)

	assert.True(t, email1.Equals(email2))
	assert.False(t, email1.Equals(email3))
}
