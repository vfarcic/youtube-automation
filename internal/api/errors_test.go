package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRespondJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		payload    interface{}
		wantStatus int
		wantBody   string
	}{
		{
			name:       "ok with payload",
			status:     http.StatusOK,
			payload:    map[string]string{"key": "value"},
			wantStatus: http.StatusOK,
			wantBody:   `{"key":"value"}`,
		},
		{
			name:       "created with nil payload",
			status:     http.StatusCreated,
			payload:    nil,
			wantStatus: http.StatusCreated,
			wantBody:   "",
		},
		{
			name:       "no content",
			status:     http.StatusNoContent,
			payload:    nil,
			wantStatus: http.StatusNoContent,
			wantBody:   "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondJSON(w, tt.status, tt.payload)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			if tt.wantBody != "" {
				var got, want interface{}
				json.Unmarshal(w.Body.Bytes(), &got)
				json.Unmarshal([]byte(tt.wantBody), &want)
				gotJSON, _ := json.Marshal(got)
				wantJSON, _ := json.Marshal(want)
				if string(gotJSON) != string(wantJSON) {
					t.Errorf("body = %s, want %s", gotJSON, wantJSON)
				}
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		errMsg     string
		detail     string
		wantStatus int
	}{
		{
			name:       "bad request with detail",
			status:     http.StatusBadRequest,
			errMsg:     "invalid input",
			detail:     "name is required",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "not found without detail",
			status:     http.StatusNotFound,
			errMsg:     "not found",
			detail:     "",
			wantStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondError(w, tt.status, tt.errMsg, tt.detail)

			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}

			var resp ErrorResponse
			json.NewDecoder(w.Body).Decode(&resp)
			if resp.Error != tt.errMsg {
				t.Errorf("error = %q, want %q", resp.Error, tt.errMsg)
			}
			if resp.Detail != tt.detail {
				t.Errorf("detail = %q, want %q", resp.Detail, tt.detail)
			}
		})
	}
}
