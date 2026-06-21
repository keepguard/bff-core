package valueobjects

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewUserType(t *testing.T) {
	tests := []struct {
		name        string
		userType    string
		expectError bool
		expected    string
	}{
		{
			name:        "valid person type",
			userType:    "PERSON",
			expectError: false,
			expected:    "PERSON",
		},
		{
			name:        "valid company type",
			userType:    "COMPANY",
			expectError: false,
			expected:    "COMPANY",
		},
		{
			name:        "valid type with lowercase",
			userType:    "person",
			expectError: false,
			expected:    "PERSON",
		},
		{
			name:        "valid type with mixed case",
			userType:    "Company",
			expectError: false,
			expected:    "COMPANY",
		},
		{
			name:        "type with spaces should be trimmed",
			userType:    " PERSON ",
			expectError: false,
			expected:    "PERSON",
		},
		{
			name:        "invalid admin type",
			userType:    "ADMIN",
			expectError: true,
			expected:    "",
		},
		{
			name:        "invalid user type",
			userType:    "USER",
			expectError: true,
			expected:    "",
		},
		{
			name:        "invalid manager type",
			userType:    "MANAGER",
			expectError: true,
			expected:    "",
		},
		{
			name:        "invalid viewer type",
			userType:    "VIEWER",
			expectError: true,
			expected:    "",
		},
		{
			name:        "empty type",
			userType:    "",
			expectError: true,
			expected:    "",
		},
		{
			name:        "invalid type",
			userType:    "INVALID",
			expectError: true,
			expected:    "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			userType, err := NewUserType(tt.userType)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, UserType{}, userType)
			} else {
				assert.NoError(t, err)
				assert.NotEqual(t, UserType{}, userType)
				assert.Equal(t, tt.expected, userType.Value())
			}
		})
	}
}

func TestUserType_Value(t *testing.T) {
	userType, err := NewUserType("PERSON")
	assert.NoError(t, err)
	assert.Equal(t, "PERSON", userType.Value())

	companyType, err := NewUserType("COMPANY")
	assert.NoError(t, err)
	assert.Equal(t, "COMPANY", companyType.Value())
}

func TestUserType_String(t *testing.T) {
	userType, err := NewUserType("PERSON")
	assert.NoError(t, err)
	assert.Equal(t, "PERSON", userType.String())

	companyType, err := NewUserType("COMPANY")
	assert.NoError(t, err)
	assert.Equal(t, "COMPANY", companyType.String())
}

func TestUserType_Equals(t *testing.T) {
	userType1, err := NewUserType("PERSON")
	assert.NoError(t, err)

	userType2, err := NewUserType("PERSON")
	assert.NoError(t, err)

	companyType, err := NewUserType("COMPANY")
	assert.NoError(t, err)

	assert.True(t, userType1.Equals(userType2))
	assert.False(t, userType1.Equals(companyType))
}

func TestUserType_IsPerson(t *testing.T) {
	personType, err := NewUserType("PERSON")
	assert.NoError(t, err)
	assert.True(t, personType.IsPerson())

	companyType, err := NewUserType("COMPANY")
	assert.NoError(t, err)
	assert.False(t, companyType.IsPerson())
}

func TestUserType_IsCompany(t *testing.T) {
	companyType, err := NewUserType("COMPANY")
	assert.NoError(t, err)
	assert.True(t, companyType.IsCompany())

	personType, err := NewUserType("PERSON")
	assert.NoError(t, err)
	assert.False(t, personType.IsCompany())
}

func TestUserType_HasPermission(t *testing.T) {
	personType, err := NewUserType("PERSON")
	assert.NoError(t, err)

	companyType, err := NewUserType("COMPANY")
	assert.NoError(t, err)

	// Test permissions for PERSON
	assert.True(t, personType.HasPermission("READ"))
	assert.True(t, personType.HasPermission("WRITE"))
	assert.False(t, personType.HasPermission("ADMIN"))

	// Test permissions for COMPANY
	assert.True(t, companyType.HasPermission("READ"))
	assert.True(t, companyType.HasPermission("WRITE"))
	assert.False(t, companyType.HasPermission("ADMIN"))
}
