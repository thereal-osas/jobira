package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJSON_Success(t *testing.T) {
	rec := httptest.NewRecorder()

	data := map[string]any{
		"id":      10,
		"message": "success",
	}

	JSON(
		rec,
		http.StatusCreated,
		data,
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"expected 201, got %d",
			rec.Code,
		)
	}

	if got := rec.Header().Get(
		"Content-Type",
	); got != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			got,
		)
	}

	var body map[string]any

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&body,
	); err != nil {
		t.Fatalf(
			"failed to decode JSON: %v",
			err,
		)
	}

	if body["message"] != "success" {
		t.Fatalf(
			"expected success, got %v",
			body["message"],
		)
	}
}

func TestJSON_NilData(t *testing.T) {
	rec := httptest.NewRecorder()

	JSON(
		rec,
		http.StatusNoContent,
		nil,
	)

	if rec.Code != http.StatusNoContent {
		t.Fatalf(
			"expected 204, got %d",
			rec.Code,
		)
	}

	if rec.Body.Len() != 0 {
		t.Fatalf(
			"expected empty body, got %q",
			rec.Body.String(),
		)
	}

	if got := rec.Header().Get(
		"Content-Type",
	); got != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			got,
		)
	}
}

func TestJSON_String(t *testing.T) {
	rec := httptest.NewRecorder()

	JSON(
		rec,
		http.StatusOK,
		"hello",
	)

	if rec.Code != http.StatusOK {
		t.Fatalf(
			"expected 200, got %d",
			rec.Code,
		)
	}

	var body string

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&body,
	); err != nil {
		t.Fatalf(
			"failed to decode JSON: %v",
			err,
		)
	}

	if body != "hello" {
		t.Fatalf(
			"expected hello, got %q",
			body,
		)
	}
}

func TestError(t *testing.T) {
	rec := httptest.NewRecorder()

	Error(
		rec,
		http.StatusBadRequest,
		"invalid input",
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"expected 400, got %d",
			rec.Code,
		)
	}

	if got := rec.Header().Get(
		"Content-Type",
	); got != "application/json" {
		t.Fatalf(
			"expected application/json, got %q",
			got,
		)
	}

	var body ErrorBody

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&body,
	); err != nil {
		t.Fatalf(
			"failed to decode body: %v",
			err,
		)
	}

	if body.Error != "invalid input" {
		t.Fatalf(
			"expected invalid input, got %q",
			body.Error,
		)
	}
}

func TestError_InternalServerError(
	t *testing.T,
) {
	rec := httptest.NewRecorder()

	Error(
		rec,
		http.StatusInternalServerError,
		"internal server error",
	)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf(
			"expected 500, got %d",
			rec.Code,
		)
	}

	var body ErrorBody

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&body,
	); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if body.Error != "internal server error" {
		t.Fatalf(
			"unexpected error %q",
			body.Error,
		)
	}
}

func TestError_Unauthorized(t *testing.T) {
	rec := httptest.NewRecorder()

	Error(
		rec,
		http.StatusUnauthorized,
		"unauthorized",
	)

	var body ErrorBody

	if err := json.Unmarshal(
		rec.Body.Bytes(),
		&body,
	); err != nil {
		t.Fatalf(
			"decode response: %v",
			err,
		)
	}

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf(
			"expected 401, got %d",
			rec.Code,
		)
	}

	if body.Error != "unauthorized" {
		t.Fatalf(
			"unexpected error %q",
			body.Error,
		)
	}
}
