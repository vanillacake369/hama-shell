package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSh(t *testing.T) {
	// GIVEN
	// WHEN
	err := ssh()

	// THEN
	assert.NoError(t, err)
}

func TestSShWithPTY(t *testing.T) {
	// GIVEN
	// WHEN
	err := ssh_with_pty()

	// THEN
	assert.NoError(t, err)
}
