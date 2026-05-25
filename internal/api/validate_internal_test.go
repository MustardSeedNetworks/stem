package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type stemValidateFixture struct {
	Name string `json:"name" validate:"required"`
	Code string `json:"code" validate:"required,numeric,len=6"`
	Mode string `json:"mode" validate:"required,oneof=reflector test_master"`
	Port int    `json:"port" validate:"required,gte=1,lte=65535"`
}

func TestValidateStruct_Valid(t *testing.T) {
	w := httptest.NewRecorder()
	dto := stemValidateFixture{Name: "alpha", Code: "123456", Mode: "reflector", Port: 8080}

	if !validateStruct(w, &dto) {
		t.Fatalf("expected validation to pass, got %d: %s", w.Code, w.Body.String())
	}
}

func TestValidateStruct_MissingRequired(t *testing.T) {
	w := httptest.NewRecorder()
	dto := stemValidateFixture{Code: "123456", Mode: "reflector", Port: 8080}

	if validateStruct(w, &dto) {
		t.Fatal("expected validation to fail for missing required field")
	}
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}

	var env map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &env); err != nil {
		t.Fatalf("response not valid JSON: %v", err)
	}
	msg, _ := env["message"].(string)
	if !strings.Contains(msg, "name") {
		t.Errorf("message should mention `name`: %q", msg)
	}
	if !strings.Contains(msg, "required") {
		t.Errorf("message should mention `required` rule: %q", msg)
	}
}

func TestValidateStruct_InvalidEnum(t *testing.T) {
	w := httptest.NewRecorder()
	dto := stemValidateFixture{Name: "alpha", Code: "123456", Mode: "passthrough", Port: 8080}

	if validateStruct(w, &dto) {
		t.Fatal("expected validation to fail for invalid mode")
	}
	if !strings.Contains(w.Body.String(), "mode") {
		t.Errorf("response should mention mode: %s", w.Body.String())
	}
}

func TestValidateStruct_OutOfRange(t *testing.T) {
	w := httptest.NewRecorder()
	dto := stemValidateFixture{Name: "alpha", Code: "123456", Mode: "reflector", Port: 99999}

	if validateStruct(w, &dto) {
		t.Fatal("expected validation to fail for port > 65535")
	}
}

func TestValidateStruct_MultipleFailures(t *testing.T) {
	w := httptest.NewRecorder()
	dto := stemValidateFixture{} // every field invalid

	if validateStruct(w, &dto) {
		t.Fatal("expected validation to fail for empty DTO")
	}
	body := w.Body.String()
	for _, field := range []string{"name", "code", "mode", "port"} {
		if !strings.Contains(body, field) {
			t.Errorf("expected response to mention %q: %s", field, body)
		}
	}
}

func TestValidateStruct_AuthLoginRequest_Real(t *testing.T) {
	// Smoke test against an actual tagged DTO to confirm the wiring.
	w := httptest.NewRecorder()
	if validateStruct(w, &AuthLoginRequest{}) {
		t.Fatal("expected empty AuthLoginRequest to fail required")
	}
	body := w.Body.String()
	for _, field := range []string{"username", "password"} {
		if !strings.Contains(body, field) {
			t.Errorf("expected message to mention %q: %s", field, body)
		}
	}
}
