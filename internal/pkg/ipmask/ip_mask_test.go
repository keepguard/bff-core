package ipmask

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTruncate(t *testing.T) {
	assert.Equal(t, "203.0.113.0", Truncate("203.0.113.77"))
	assert.Equal(t, "2001:db8:1::", Truncate("2001:db8:1:2:3:4:5:6"))
	assert.Empty(t, Truncate("não é ip"))
	assert.Empty(t, Truncate(""))
}
