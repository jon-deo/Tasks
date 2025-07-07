package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	"github.com/jung-kurt/gofpdf"
)

// StudentDetail represents the student data structure from the Node.js API
type StudentDetail struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	Email              string `json:"email"`
	SystemAccess       bool   `json:"systemAccess"`
	Phone              string `json:"phone"`
	Gender             string `json:"gender"`
	DOB                string `json:"dob"`
	Class              string `json:"class"`
	Section            string `json:"section"`
	Roll               string `json:"roll"`
	FatherName         string `json:"fatherName"`
	FatherPhone        string `json:"fatherPhone"`
	MotherName         string `json:"motherName"`
	MotherPhone        string `json:"motherPhone"`
	GuardianName       string `json:"guardianName"`
	GuardianPhone      string `json:"guardianPhone"`
	RelationOfGuardian string `json:"relationOfGuardian"`
	CurrentAddress     string `json:"currentAddress"`
	PermanentAddress   string `json:"permanentAddress"`
	AdmissionDate      string `json:"admissionDate"`
	ReporterName       string `json:"reporterName"`
}

// LoginResponse represents the response from the login API
type LoginResponse struct {
	AccessToken string `json:"accessToken"`
	User        User   `json:"user"`
}

// User represents the user data from the login response
type User struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// Config holds application configuration
type Config struct {
	Port          string
	BackendURL    string
	AdminEmail    string
	AdminPassword string
}

var authTokens struct {
	accessToken  string
	refreshToken string
	csrfToken    string
	lastRefresh  time.Time
}

func main() {
	// Load environment variables
	config := loadConfig()

	// Authenticate with backend
	authenticateWithBackend(config)

	// Create router
	r := mux.NewRouter()

	// Define routes
	r.HandleFunc("/api/v1/students/{id}/report", generateReportHandler(config)).Methods("GET")
	r.HandleFunc("/health", healthCheckHandler).Methods("GET")

	// Start server
	port := config.Port
	log.Printf("Server starting on port %s...", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}

func loadConfig() Config {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Set default values
	config := Config{
		Port:          getEnv("PORT", "5008"),
		BackendURL:    getEnv("BACKEND_URL", "http://localhost:5007/api/v1"),
		AdminEmail:    getEnv("ADMIN_EMAIL", "admin@school-admin.com"),
		AdminPassword: getEnv("ADMIN_PASSWORD", "3OU4zn3q6Zh9"),
	}

	return config
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

func authenticateWithBackend(config Config) {
	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Prepare login request
	loginURL := fmt.Sprintf("%s/auth/login", config.BackendURL)
	loginData := map[string]string{
		"username": config.AdminEmail,
		"password": config.AdminPassword,
	}
	loginJSON, err := json.Marshal(loginData)
	if err != nil {
		log.Fatalf("Error preparing login request: %v", err)
	}

	// Make login request
	resp, err := client.Post(loginURL, "application/json", bytes.NewBuffer(loginJSON))
	if err != nil {
		log.Fatalf("Error authenticating with backend: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Fatalf("Backend authentication failed with status %d: %s", resp.StatusCode, body)
	}

	// Extract cookies for future requests
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "accessToken" {
			authTokens.accessToken = cookie.Value
		} else if cookie.Name == "refreshToken" {
			authTokens.refreshToken = cookie.Value
		} else if cookie.Name == "csrfToken" {
			authTokens.csrfToken = cookie.Value
		}
	}

	// Parse response body to get access token
	var loginResponse LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResponse); err != nil {
		log.Fatalf("Error parsing login response: %v", err)
	}

	log.Printf("Successfully authenticated with backend as %s (role: %s)",
		loginResponse.User.Name, loginResponse.User.Role)
	authTokens.lastRefresh = time.Now()
}

func healthCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func generateReportHandler(config Config) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract student ID from URL
		vars := mux.Vars(r)
		studentID := vars["id"]

		// Validate student ID
		id, err := strconv.Atoi(studentID)
		if err != nil || id <= 0 {
			http.Error(w, "Invalid student ID", http.StatusBadRequest)
			return
		}

		// Fetch student data from backend API
		student, err := fetchStudentData(config.BackendURL, studentID)
		if err != nil {
			log.Printf("Error fetching student data: %v", err)
			http.Error(w, "Error fetching student data", http.StatusInternalServerError)
			return
		}

		// Generate PDF
		pdfBytes, err := generatePDF(student)
		if err != nil {
			log.Printf("Error generating PDF: %v", err)
			http.Error(w, "Error generating PDF", http.StatusInternalServerError)
			return
		}

		// Set response headers
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=student_%s_report.pdf", studentID))
		w.Write(pdfBytes)
	}
}

func fetchStudentData(backendURL, studentID string) (*StudentDetail, error) {
	// Create HTTP client with timeout
	client := &http.Client{Timeout: 10 * time.Second}

	// Check if tokens need refresh (tokens expire after 15 minutes)
	if time.Since(authTokens.lastRefresh) > 14*time.Minute {
		log.Println("Auth tokens might be expired, refreshing...")
		config := loadConfig()
		authenticateWithBackend(config)
	}

	// Make request to backend API
	url := fmt.Sprintf("%s/students/%s", backendURL, studentID)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	// Add cookies for authentication
	req.AddCookie(&http.Cookie{Name: "accessToken", Value: authTokens.accessToken})
	req.AddCookie(&http.Cookie{Name: "refreshToken", Value: authTokens.refreshToken})
	req.AddCookie(&http.Cookie{Name: "csrfToken", Value: authTokens.csrfToken})
	
	// Add CSRF token to headers
	req.Header.Set("X-CSRF-Token", authTokens.csrfToken)

	// Execute request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request to backend: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode == http.StatusUnauthorized {
		// Try to refresh tokens and retry the request once
		log.Println("Received 401 Unauthorized, attempting to refresh tokens and retry...")
		config := loadConfig()
		authenticateWithBackend(config)

		// Retry the request with new tokens
		req, _ = http.NewRequest("GET", url, nil)
		req.AddCookie(&http.Cookie{Name: "accessToken", Value: authTokens.accessToken})
		req.AddCookie(&http.Cookie{Name: "refreshToken", Value: authTokens.refreshToken})
		req.AddCookie(&http.Cookie{Name: "csrfToken", Value: authTokens.csrfToken})
		
		// Add CSRF token to headers
		req.Header.Set("X-CSRF-Token", authTokens.csrfToken)

		resp, err = client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("error making request to backend after token refresh: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("backend API returned non-200 status after token refresh: %d, body: %s", resp.StatusCode, body)
		}
	} else if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("student with ID %s not found", studentID)
	} else if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("backend API returned non-200 status: %d, body: %s", resp.StatusCode, body)
	}

	// Parse response body
	var student StudentDetail
	if err := json.NewDecoder(resp.Body).Decode(&student); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	return &student, nil
}

func generatePDF(student *StudentDetail) ([]byte, error) {
	// Create new PDF document
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	// Set font
	pdf.SetFont("Arial", "B", 16)

	// Add title
	pdf.Cell(40, 10, "Student Report")
	pdf.Ln(15)

	// Add school logo or header
	pdf.SetFont("Arial", "B", 14)
	pdf.Cell(40, 10, "School Management System")
	pdf.Ln(10)

	// Add generation date
	pdf.SetFont("Arial", "I", 10)
	pdf.Cell(40, 10, fmt.Sprintf("Generated on: %s", time.Now().Format("January 2, 2006")))
	pdf.Ln(15)

	// Add student information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Personal Information")
	pdf.Ln(10)

	// Add student details in a table-like format
	addInfoRow(pdf, "ID", fmt.Sprintf("%d", student.ID))
	addInfoRow(pdf, "Name", student.Name)
	addInfoRow(pdf, "Email", student.Email)
	addInfoRow(pdf, "Phone", student.Phone)
	addInfoRow(pdf, "Gender", student.Gender)
	addInfoRow(pdf, "Date of Birth", student.DOB)
	addInfoRow(pdf, "Admission Date", student.AdmissionDate)
	pdf.Ln(10)

	// Add academic information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Academic Information")
	pdf.Ln(10)

	addInfoRow(pdf, "Class", student.Class)
	addInfoRow(pdf, "Section", student.Section)
	addInfoRow(pdf, "Roll Number", student.Roll)
	pdf.Ln(10)

	// Add family information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Family Information")
	pdf.Ln(10)

	addInfoRow(pdf, "Father's Name", student.FatherName)
	addInfoRow(pdf, "Father's Phone", student.FatherPhone)
	addInfoRow(pdf, "Mother's Name", student.MotherName)
	addInfoRow(pdf, "Mother's Phone", student.MotherPhone)
	addInfoRow(pdf, "Guardian's Name", student.GuardianName)
	addInfoRow(pdf, "Guardian's Phone", student.GuardianPhone)
	addInfoRow(pdf, "Relation of Guardian", student.RelationOfGuardian)
	pdf.Ln(10)

	// Add address information
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(40, 10, "Address Information")
	pdf.Ln(10)

	addInfoRow(pdf, "Current Address", student.CurrentAddress)
	addInfoRow(pdf, "Permanent Address", student.PermanentAddress)
	pdf.Ln(10)

	// Add footer
	pdf.SetY(-30)
	pdf.SetFont("Arial", "I", 8)
	pdf.Cell(0, 10, "This is an official document of School Management System")
	pdf.Ln(5)
	pdf.Cell(0, 10, fmt.Sprintf("Report generated by: %s", student.ReporterName))

	// Output PDF to buffer
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func addInfoRow(pdf *gofpdf.Fpdf, label, value string) {
	pdf.SetFont("Arial", "B", 10)
	pdf.Cell(40, 8, label+":")
	pdf.SetFont("Arial", "", 10)
	pdf.Cell(0, 8, value)
	pdf.Ln(8)
}
