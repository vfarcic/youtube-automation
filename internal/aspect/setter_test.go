package aspect

import (
	"testing"

	"devopstoolkit/youtube-automation/internal/storage"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetFieldValueByJSONPath(t *testing.T) {
	tests := []struct {
		name      string
		jsonPath  string
		value     interface{}
		setup     func() *storage.Video
		verify    func(t *testing.T, v *storage.Video)
		wantErr   bool
		errSubstr string
	}{
		{
			name:     "set string field",
			jsonPath: "projectName",
			value:    "MyProject",
			setup:    func() *storage.Video { return &storage.Video{} },
			verify: func(t *testing.T, v *storage.Video) {
				assert.Equal(t, "MyProject", v.ProjectName)
			},
		},
		{
			name:     "set bool field to true",
			jsonPath: "code",
			value:    true,
			setup:    func() *storage.Video { return &storage.Video{} },
			verify: func(t *testing.T, v *storage.Video) {
				assert.True(t, v.Code)
			},
		},
		{
			name:     "set bool field to false",
			jsonPath: "screen",
			value:    false,
			setup:    func() *storage.Video { return &storage.Video{Screen: true} },
			verify: func(t *testing.T, v *storage.Video) {
				assert.False(t, v.Screen)
			},
		},
		{
			name:     "set nested field - sponsorship.amount",
			jsonPath: "sponsorship.amount",
			value:    "5000",
			setup:    func() *storage.Video { return &storage.Video{} },
			verify: func(t *testing.T, v *storage.Video) {
				assert.Equal(t, "5000", v.Sponsorship.Amount)
			},
		},
		{
			name:     "set nested field - sponsorship.name",
			jsonPath: "sponsorship.name",
			value:    "ACME Corp",
			setup:    func() *storage.Video { return &storage.Video{} },
			verify: func(t *testing.T, v *storage.Video) {
				assert.Equal(t, "ACME Corp", v.Sponsorship.Name)
			},
		},
		{
			name:     "set slice field - titles",
			jsonPath: "titles",
			value: []interface{}{
				map[string]interface{}{"index": float64(1), "text": "Title One"},
				map[string]interface{}{"index": float64(2), "text": "Title Two"},
			},
			setup: func() *storage.Video { return &storage.Video{} },
			verify: func(t *testing.T, v *storage.Video) {
				require.Len(t, v.Titles, 2)
				assert.Equal(t, "Title One", v.Titles[0].Text)
				assert.Equal(t, 1, v.Titles[0].Index)
				assert.Equal(t, "Title Two", v.Titles[1].Text)
			},
		},
		{
			name:      "invalid path - nonexistent field",
			jsonPath:  "nonExistentField",
			value:     "value",
			setup:     func() *storage.Video { return &storage.Video{} },
			wantErr:   true,
			errSubstr: "not found",
		},
		{
			name:      "invalid path - nonexistent nested parent",
			jsonPath:  "bogus.field",
			value:     "value",
			setup:     func() *storage.Video { return &storage.Video{} },
			wantErr:   true,
			errSubstr: "not found",
		},
		{
			name:      "type mismatch - string to bool",
			jsonPath:  "code",
			value:     "not-a-bool",
			setup:     func() *storage.Video { return &storage.Video{} },
			wantErr:   true,
			errSubstr: "cannot assign",
		},
		{
			name:      "type mismatch - bool to string",
			jsonPath:  "projectName",
			value:     true,
			setup:     func() *storage.Video { return &storage.Video{} },
			wantErr:   true,
			errSubstr: "cannot assign",
		},
		{
			name:     "nil value sets zero value",
			jsonPath: "projectName",
			value:    nil,
			setup:    func() *storage.Video { return &storage.Video{ProjectName: "existing"} },
			verify: func(t *testing.T, v *storage.Video) {
				assert.Equal(t, "", v.ProjectName)
			},
		},
		{
			name:     "preserves other fields when setting one",
			jsonPath: "projectName",
			value:    "NewProject",
			setup: func() *storage.Video {
				return &storage.Video{
					ProjectURL: "https://example.com",
					Code:       true,
				}
			},
			verify: func(t *testing.T, v *storage.Video) {
				assert.Equal(t, "NewProject", v.ProjectName)
				assert.Equal(t, "https://example.com", v.ProjectURL)
				assert.True(t, v.Code)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.setup()
			err := SetFieldValueByJSONPath(v, tt.jsonPath, tt.value)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errSubstr != "" {
					assert.Contains(t, err.Error(), tt.errSubstr)
				}
				return
			}

			require.NoError(t, err)
			if tt.verify != nil {
				tt.verify(t, v)
			}
		})
	}
}

func TestSetFieldValueByJSONPath_NonPointer(t *testing.T) {
	v := storage.Video{}
	err := SetFieldValueByJSONPath(v, "projectName", "test")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pointer to a struct")
}

func TestSetFieldValueByJSONPath_MapField(t *testing.T) {
	v := &storage.Video{}
	dubbingData := map[string]interface{}{
		"es": map[string]interface{}{
			"dubbingId":       "dub-123",
			"dubbedVideoPath": "/path/to/dubbed.mp4",
			"title":           "Título del video",
		},
		"fr": map[string]interface{}{
			"dubbingId": "dub-456",
			"title":     "Titre de la vidéo",
		},
	}

	err := SetFieldValueByJSONPath(v, "dubbing", dubbingData)
	require.NoError(t, err)
	require.Len(t, v.Dubbing, 2)
	assert.Equal(t, "dub-123", v.Dubbing["es"].DubbingID)
	assert.Equal(t, "/path/to/dubbed.mp4", v.Dubbing["es"].DubbedVideoPath)
	assert.Equal(t, "Título del video", v.Dubbing["es"].Title)
	assert.Equal(t, "dub-456", v.Dubbing["fr"].DubbingID)
}

func TestSetFieldValueByJSONPath_StructField(t *testing.T) {
	v := &storage.Video{}
	sponsorshipData := map[string]interface{}{
		"amount":  "5000",
		"emails":  "sponsor@example.com",
		"blocked": "",
		"name":    "ACME",
		"url":     "https://acme.com",
	}

	err := SetFieldValueByJSONPath(v, "sponsorship", sponsorshipData)
	require.NoError(t, err)
	assert.Equal(t, "5000", v.Sponsorship.Amount)
	assert.Equal(t, "sponsor@example.com", v.Sponsorship.Emails)
	assert.Equal(t, "ACME", v.Sponsorship.Name)
}

func TestSetFieldValueByJSONPath_FloatToInt(t *testing.T) {
	// JSON decodes numbers as float64 — verify we handle int conversion
	type testStruct struct {
		Count int `json:"count"`
	}
	s := &testStruct{}
	err := SetFieldValueByJSONPath(s, "count", float64(42))
	require.NoError(t, err)
	assert.Equal(t, 42, s.Count)
}
