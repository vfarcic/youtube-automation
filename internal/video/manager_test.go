package video

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewManager(t *testing.T) {
	mockFilePathFunc := func(category, name, extension string) string {
		return "mock/path"
	}
	manager := NewManager(mockFilePathFunc)
	assert.NotNil(t, manager, "NewManager should return a non-nil Manager")
	assert.NotNil(t, manager.filePathFunc, "filePathFunc should be set in Manager")
}
