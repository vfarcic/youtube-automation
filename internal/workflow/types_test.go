package workflow

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDirectoryStruct(t *testing.T) {
	dirName := "Test Directory"
	dirPath := "/path/to/test/directory"

	d := Directory{
		Name: dirName,
		Path: dirPath,
	}

	assert.Equal(t, dirName, d.Name, "Directory Name should match the input")
	assert.Equal(t, dirPath, d.Path, "Directory Path should match the input")
}
