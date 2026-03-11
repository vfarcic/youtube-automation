package ai

import (
	"testing"
)

func TestExtractJSONFromResponse(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
	}{
		{
			name:  "Direct JSON object",
			input: `{"key": "value"}`,
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON in markdown code block with json tag",
			input: "```json\n{\"key\": \"value\"}\n```",
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON in markdown code block without json tag",
			input: "```\n{\"key\": \"value\"}\n```",
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON in markdown with CRLF line endings",
			input: "```json\r\n{\"key\": \"value\"}\r\n```",
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON in markdown with trailing space after json tag",
			input: "```json \n{\"key\": \"value\"}\n```",
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON with text before code block",
			input: "Here is the result:\n```json\n{\"key\": \"value\"}\n```",
			want:  `{"key": "value"}`,
		},
		{
			name:  "JSON with text after code block",
			input: "```json\n{\"key\": \"value\"}\n```\nSome trailing text",
			want:  `{"key": "value"}`,
		},
		{
			name:  "Multiline JSON in code block",
			input: "```json\n{\n  \"key\": \"value\",\n  \"num\": 42\n}\n```",
			want:  "{\n  \"key\": \"value\",\n  \"num\": 42\n}",
		},
		{
			name:  "Multiline JSON with CRLF",
			input: "```json\r\n{\r\n  \"key\": \"value\"\r\n}\r\n```",
			want:  "{\n  \"key\": \"value\"\n}",
		},
		{
			name:  "Direct JSON with surrounding whitespace",
			input: "  {\"key\": \"value\"}  ",
			want:  `{"key": "value"}`,
		},
		{
			name:  "Empty input",
			input: "",
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractJSONFromResponse(tt.input)
			if got != tt.want {
				t.Errorf("ExtractJSONFromResponse() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestParseJSONResponse(t *testing.T) {
	type testStruct struct {
		Key string `json:"key"`
		Num int    `json:"num"`
	}

	tests := []struct {
		name    string
		input   string
		want    testStruct
		wantErr bool
	}{
		{
			name:  "Direct JSON",
			input: `{"key": "value", "num": 42}`,
			want:  testStruct{Key: "value", Num: 42},
		},
		{
			name:  "JSON in markdown code block",
			input: "```json\n{\"key\": \"value\", \"num\": 42}\n```",
			want:  testStruct{Key: "value", Num: 42},
		},
		{
			name:  "JSON in markdown with CRLF",
			input: "```json\r\n{\"key\": \"value\", \"num\": 42}\r\n```",
			want:  testStruct{Key: "value", Num: 42},
		},
		{
			name:    "Invalid JSON",
			input:   "not json at all",
			wantErr: true,
		},
		{
			name:    "Empty input",
			input:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got testStruct
			err := ParseJSONResponse(tt.input, &got)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseJSONResponse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && got != tt.want {
				t.Errorf("ParseJSONResponse() got = %+v, want %+v", got, tt.want)
			}
		})
	}
}
