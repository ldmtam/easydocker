package easydocker

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreatePool(t *testing.T) {
	pool, err := NewPool("")
	assert.Nil(t, err)
	assert.NotNil(t, pool)
}
