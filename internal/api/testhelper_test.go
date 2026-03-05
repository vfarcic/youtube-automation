package api

import (
	"os"
	"path/filepath"
	"testing"

	"devopstoolkit/youtube-automation/internal/filesystem"
	"devopstoolkit/youtube-automation/internal/service"
	"devopstoolkit/youtube-automation/internal/video"

	"gopkg.in/yaml.v3"

	"devopstoolkit/youtube-automation/internal/storage"
)

// testEnv bundles a test server with its temp directory for cleanup.
type testEnv struct {
	server *Server
	tmpDir string
}

// setupTestEnv creates a temporary manuscript directory with an index.yaml
// and returns a fully wired Server for testing.
func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()

	tmpDir := t.TempDir()

	// Create manuscript base dir and a category
	catDir := filepath.Join(tmpDir, "manuscript", "devops")
	if err := os.MkdirAll(catDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Create empty index.yaml
	indexPath := filepath.Join(tmpDir, "index.yaml")
	if err := os.WriteFile(indexPath, []byte("[]"), 0644); err != nil {
		t.Fatal(err)
	}

	// We need to chdir so that the service layer can find "manuscript/" and "index.yaml"
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(origDir) })

	fsOps := filesystem.NewOperations()
	videoManager := video.NewManager(fsOps.GetFilePath)
	videoService := service.NewVideoService(indexPath, fsOps, videoManager)

	srv := NewServer(videoService, videoManager)
	return &testEnv{
		server: srv,
		tmpDir: tmpDir,
	}
}

// seedVideo writes a video YAML file and updates the index so the service can find it.
func seedVideo(t *testing.T, env *testEnv, v storage.Video) {
	t.Helper()

	catDir := filepath.Join(env.tmpDir, "manuscript", v.Category)
	if err := os.MkdirAll(catDir, 0755); err != nil {
		t.Fatal(err)
	}

	// Write video YAML
	videoPath := filepath.Join(catDir, v.Name+".yaml")
	data, err := yaml.Marshal(&v)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(videoPath, data, 0644); err != nil {
		t.Fatal(err)
	}

	// Also write an empty .md file so delete can find it
	mdPath := filepath.Join(catDir, v.Name+".md")
	if err := os.WriteFile(mdPath, []byte("# test"), 0644); err != nil {
		t.Fatal(err)
	}

	// Update index
	indexPath := filepath.Join(env.tmpDir, "index.yaml")
	indexData, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatal(err)
	}
	var index []storage.VideoIndex
	yaml.Unmarshal(indexData, &index)
	index = append(index, storage.VideoIndex{Name: v.Name, Category: v.Category})
	newData, _ := yaml.Marshal(index)
	if err := os.WriteFile(indexPath, newData, 0644); err != nil {
		t.Fatal(err)
	}
}
