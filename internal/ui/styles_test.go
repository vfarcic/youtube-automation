package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStylesAreInitialized(t *testing.T) {
	assert.NotNil(t, RedStyle, "RedStyle should be initialized")
	assert.NotNil(t, GreenStyle, "GreenStyle should be initialized")
	assert.NotNil(t, OrangeStyle, "OrangeStyle should be initialized")
	assert.NotNil(t, FarFutureStyle, "FarFutureStyle should be initialized")
	assert.NotNil(t, ConfirmationStyle, "ConfirmationStyle should be initialized")
	assert.NotNil(t, ErrorStyle, "ErrorStyle should be initialized")
}
