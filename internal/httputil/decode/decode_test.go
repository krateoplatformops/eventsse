package decode

import (
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type testData struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

func TestJSONBody(t *testing.T) {
	tests := []struct {
		name        string
		body        string
		contentType string
		wantError   *MalformedRequest
	}{
		{
			name:        "Valid JSON",
			body:        `{"name": "John", "age": 30}`,
			contentType: "application/json",
			wantError:   nil,
		},
		{
			name:        "Badly formed JSON",
			body:        `{"name": "John", "age": 30`,
			contentType: "application/json",
			wantError:   &MalformedRequest{Status: http.StatusBadRequest, Msg: "Request body contains badly-formed JSON"},
		},
		{
			name:        "Invalid value type",
			body:        `{"name": "John", "age": "thirty"}`,
			contentType: "application/json",
			wantError:   &MalformedRequest{Status: http.StatusBadRequest, Msg: "Request body contains an invalid value for the \"age\" field"},
		},
		{
			name:        "Unknown field",
			body:        `{"name": "John", "age": 30, "extra": "field"}`,
			contentType: "application/json",
			wantError:   &MalformedRequest{Status: http.StatusBadRequest, Msg: "Request body contains unknown field \"extra\""},
		},
		{
			name:        "Empty body",
			body:        ``,
			contentType: "application/json",
			wantError:   &MalformedRequest{Status: http.StatusNoContent, Msg: "Request body is empty"},
		},
		{
			name:        "Multiple JSON objects",
			body:        `{"name": "John", "age": 30}{"name": "Jane", "age": 25}`,
			contentType: "application/json",
			wantError:   &MalformedRequest{Status: http.StatusBadRequest, Msg: "Request body must only contain a single JSON object"},
		},
		{
			name:        "Incorrect Content-Type",
			body:        `{"name": "John", "age": 30}`,
			contentType: "text/plain",
			wantError:   &MalformedRequest{Status: http.StatusUnsupportedMediaType, Msg: "Content-Type header is not application/json"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", tt.contentType)
			w := httptest.NewRecorder()

			var dst testData
			err := JSONBody(w, req, &dst)

			if tt.wantError != nil {
				var mr *MalformedRequest
				if !errors.As(err, &mr) {
					t.Fatalf("Expected error of type *MalformedRequest, got %v", err)
				}
				if mr.Status != tt.wantError.Status || !strings.Contains(mr.Msg, tt.wantError.Msg) {
					t.Errorf("Expected error: %v, got: %v", tt.wantError, mr)
				}
			} else if err != nil {
				t.Errorf("Did not expect an error, got: %v", err)
			}
		})
	}
}

func TestIsEmptyBodyError(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		wantIsEmpty bool
	}{
		{
			name:        "Empty body error",
			err:         &MalformedRequest{Status: http.StatusNoContent, Msg: "Request body is empty"},
			wantIsEmpty: true,
		},
		{
			name:        "Other error",
			err:         &MalformedRequest{Status: http.StatusBadRequest, Msg: "Request body contains badly-formed JSON"},
			wantIsEmpty: false,
		},
		{
			name:        "Nil error",
			err:         nil,
			wantIsEmpty: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsEmptyBodyError(tt.err)
			if got != tt.wantIsEmpty {
				t.Errorf("IsEmptyBodyError(%v) = %v; want %v", tt.err, got, tt.wantIsEmpty)
			}
		})
	}
}

func TestBodyError(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{
			name: "Empty body error",
			err:  &MalformedRequest{Status: http.StatusNoContent, Msg: "Request body is empty"},
		},
		{
			name: "Other error",
			err:  &MalformedRequest{Status: http.StatusBadRequest, Msg: "Request body contains badly-formed JSON"},
		},
		{
			name: "Nil error",
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err != nil {
				got := tt.err.Error()
				if len(got) == 0 {
					t.Errorf("Expectd error")
				}
			}
		})
	}
}
