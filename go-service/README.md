# Student Report PDF Generator Microservice

A Go microservice that generates PDF reports for students by consuming data from the Node.js backend API.

## Overview

This microservice provides an API endpoint to generate PDF reports for students. It fetches student data from the existing Node.js backend API and converts it into a well-formatted PDF document.

## Features

- RESTful API endpoint for generating student reports
- Fetches student data from the Node.js backend API
- Generates professional PDF reports with student information
- Health check endpoint for monitoring
- Automatic authentication with the backend API

## Prerequisites

- Go 1.16 or higher
- Node.js backend running (for data source)
- PostgreSQL database with student data

## Installation

1. Clone the repository
2. Navigate to the `go-service` directory
3. Install dependencies:

```bash
go mod download
```

## Configuration

The service can be configured using environment variables or a `.env` file:

- `PORT`: The port on which the service will run (default: 5008)
- `BACKEND_URL`: The URL of the Node.js backend API (default: http://localhost:5007/api/v1)
- `ADMIN_EMAIL`: The email for backend authentication (default: admin@school-admin.com)
- `ADMIN_PASSWORD`: The password for backend authentication (default: 3OU4zn3q6Zh9)

## Running the Service

Using the Makefile:

```bash
# Build and run the service
make run

# Just build the service
make build

# Run the service after building
./student-report-service
```

Or directly with Go:

```bash
go run main.go
```

## Testing

### Running Tests

You can run the tests using the Makefile or directly with Go:

```bash
# Using Makefile
make test

# Using Go directly
go test -v ./...
```

### Writing Tests

The project follows standard Go testing practices. Here's how to write tests from scratch:

1. Create a test file with the naming convention `*_test.go` (e.g., `handlers_test.go`)
2. Import the testing package and any other required packages
3. Write test functions with the naming convention `TestXxx` where `Xxx` is the function you're testing

Example of a basic test:

```go
package main

import (
	"testing"
)

func TestSomething(t *testing.T) {
	// Test setup
	result := someFunction()
	
	// Assertions
	if result != expectedValue {
		t.Errorf("Expected %v, got %v", expectedValue, result)
	}
}
```

### HTTP Handler Testing

For testing HTTP handlers, use the `net/http/httptest` package:

```go
func TestMyHandler(t *testing.T) {
	// Create a request
	req, err := http.NewRequest("GET", "/some-endpoint", nil)
	if err != nil {
		t.Fatal(err)
	}
	
	// Create a response recorder
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(myHandler)
	
	// Serve the request
	handler.ServeHTTP(rr, req)
	
	// Check status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
	
	// Check response body
	expected := `{"message":"success"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v", rr.Body.String(), expected)
	}
}
```

### Mocking External Dependencies

For mocking the backend API, you can use a test server:

```go
// Create a mock server
mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Mock response based on the request path
	if r.URL.Path == "/api/v1/students/1" {
		// Return mock student data
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, `{"id":1,"name":"Test Student","email":"test@example.com"}`) 
	} else {
		w.WriteHeader(http.StatusNotFound)
	}
}))
defer mockServer.Close()

// Use the mock server URL in your tests
config.BackendURL = mockServer.URL
```

### Test Coverage

To check test coverage:

```bash
go test -cover ./...
```

For a detailed coverage report:

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Manual Testing

Follow these steps to manually test the service without using scripts:

#### Prerequisites for Manual Testing

1. Ensure the Node.js backend is running on port 5007 (or as configured in your `.env` file)
2. Ensure you have a student record in the database (note the student ID for testing)
3. Have a tool for making HTTP requests (like cURL, Postman, or a web browser)

#### Testing the Health Endpoint

1. Start the service:
   ```bash
   make run
   ```

2. Send a GET request to the health endpoint:
   ```bash
   curl http://localhost:5008/health
   ```

3. Verify you receive a response with status "ok":
   ```json
   {"status":"ok"}
   ```

#### Testing the Report Generation

1. Ensure the service is running

2. Replace `{student_id}` with an actual student ID from your database and send a GET request:
   ```bash
   curl -o student_report.pdf http://localhost:5008/api/v1/students/{student_id}/report
   ```

3. Verify that:
   - The request returns a PDF file
   - The file downloads correctly
   - The PDF contains the correct student information when opened

#### Testing Error Scenarios

1. Test with an invalid student ID:
   ```bash
   curl -v http://localhost:5008/api/v1/students/invalid/report
   ```
   Verify you receive a 400 Bad Request response

2. Test with a non-existent student ID:
   ```bash
   curl -v http://localhost:5008/api/v1/students/9999/report
   ```
   Verify you receive an appropriate error response

3. Test with the backend service down:
   - Stop the Node.js backend service
   - Send a request to the report endpoint
   - Verify you receive a 500 Internal Server Error response

#### Testing Authentication

1. Modify the `.env` file with incorrect credentials:
   ```
   ADMIN_EMAIL=wrong@example.com
   ADMIN_PASSWORD=wrongpassword
   ```

2. Restart the service and attempt to generate a report

3. Check the logs to verify authentication failure and retry behavior

#### Browser-Based Testing

1. Start the service

2. Open a web browser and navigate to:
   ```
   http://localhost:5008/health
   ```
   Verify you see the JSON response with status "ok"

3. To test report generation, navigate to (replace {student_id} with a valid ID):
   ```
   http://localhost:5008/api/v1/students/{student_id}/report
   ```
   The browser should download the PDF file automatically

#### Testing with Postman

1. Open Postman and create a new GET request to:
   ```
   http://localhost:5008/api/v1/students/{student_id}/report
   ```

2. Send the request and verify that:
   - The response status is 200 OK
   - The response headers include `Content-Type: application/pdf`
   - The response body contains the PDF data

3. Save the response as a file with a .pdf extension and open it to verify the content

#### Performance Testing

1. Test response time for a single request:
   ```bash
   time curl -o /dev/null http://localhost:5008/api/v1/students/{student_id}/report
   ```

2. Test multiple concurrent requests (using a tool like Apache Bench):
   ```bash
   ab -n 10 -c 2 http://localhost:5008/api/v1/students/{student_id}/report
   ```
   This sends 10 requests with 2 concurrent connections

#### Security Testing

1. Test SQL Injection protection by using special characters in the student ID:
   ```bash
   curl -v "http://localhost:5008/api/v1/students/1'%20OR%20'1'='1/report"
   ```
   Verify that the service properly rejects this request

2. Test Cross-Site Scripting (XSS) protection:
   ```bash
   curl -v "http://localhost:5008/api/v1/students/<script>alert('XSS')</script>/report"
   ```
   Verify that the service properly rejects this request

3. Test for proper error handling without information leakage:
   - Send various malformed requests
   - Verify that error responses don't expose sensitive information about the system

4. Test rate limiting (if implemented):
   - Send many requests in quick succession
   - Verify that the service implements appropriate rate limiting

#### End-to-End Testing

1. Start both the backend and the Go service
2. Create a new student in the backend system
3. Generate a report for the newly created student
4. Verify that the PDF contains all the correct information
5. Delete the student from the backend
6. Attempt to generate a report for the deleted student
7. Verify that an appropriate error is returned

## API Endpoints

### Generate Student Report

```
GET /api/v1/students/:id/report
```

Generates a PDF report for the student with the specified ID.

**Parameters:**
- `id`: Student ID (required)

**Response:**
- Content-Type: application/pdf
- A downloadable PDF file containing the student's report

### Health Check

```
GET /health
```

Returns the health status of the service.

**Response:**
```json
{
  "status": "ok"
}
```

## Architecture

This microservice follows a simple architecture:

1. The client makes a request to the Go service for a student report
2. The Go service authenticates with the Node.js backend API
3. The service fetches student data from the backend API
4. The service generates a PDF report using the fetched data
5. The PDF is returned to the client as a downloadable file

## Dependencies

- [gorilla/mux](https://github.com/gorilla/mux): HTTP router and URL matcher
- [jung-kurt/gofpdf](https://github.com/jung-kurt/gofpdf): PDF generation library
- [joho/godotenv](https://github.com/joho/godotenv): Environment variable loader

## Error Handling

The service handles various error scenarios:

- Invalid student ID: Returns 400 Bad Request
- Backend API unavailable: Returns 500 Internal Server Error
- Authentication failure: Retries authentication and logs the error
- Student not found: Returns the error from the backend API
- PDF generation failure: Returns 500 Internal Server Error

## Docker

This service can be containerized using Docker.

### Building the Docker Image

```bash
# Using Makefile
make docker-build

# Or using Docker directly
docker build -t student-report-service .
```

### Running the Docker Container

```bash
# Using Makefile
make docker-run

# Or using Docker directly
docker run -p 5008:5008 student-report-service
```

### Docker Compose (with Backend)

Create a `docker-compose.yml` file to run both the Go service and the Node.js backend:

```yaml
version: '3'

services:
  backend:
    image: node-backend-image
    ports:
      - "5007:5007"
    environment:
      - NODE_ENV=production
    networks:
      - app-network

  report-service:
    build: ./go-service
    ports:
      - "5008:5008"
    environment:
      - BACKEND_URL=http://backend:5007/api/v1
    depends_on:
      - backend
    networks:
      - app-network

networks:
  app-network:
    driver: bridge
```

Run with:

```bash
docker-compose up
```