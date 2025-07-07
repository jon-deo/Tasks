package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
)

func TestHealthCheckHandler(t *testing.T) {
	// Create a request to pass to our handler
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(healthCheckHandler)

	// Call the handler
	handler.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body contains expected content
	var response map[string]string
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse response JSON: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("handler returned unexpected status: got %v want %v",
			response["status"], "ok")
	}
}

func TestGenerateReportHandlerInvalidID(t *testing.T) {
	// Create a new router
	r := mux.NewRouter()

	// Register the handler
	config := Config{
		Port:       "5008",
		BackendURL: "http://localhost:5007/api/v1",
	}
	r.HandleFunc("/api/v1/students/{id}/report", generateReportHandler(config))

	// Create a request with an invalid ID
	req, err := http.NewRequest("GET", "/api/v1/students/invalid/report", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder
	rr := httptest.NewRecorder()

	// Serve the request
	r.ServeHTTP(rr, req)

	// Check the status code is 400 Bad Request
	if status := rr.Code; status != http.StatusBadRequest {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusBadRequest)
	}
}