package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/gilabs/gims/api/internal/core/infrastructure/config"
	"github.com/gilabs/gims/api/internal/core/infrastructure/database"
	"github.com/gilabs/gims/api/internal/core/logger"
	roleModels "github.com/gilabs/gims/api/internal/role/data/models"
	userModels "github.com/gilabs/gims/api/internal/user/data/models"
)

func main() {
	logger.Init()

	// Load config
	if err := config.Load(); err != nil {
		log.Printf("Warning: Failed to load config from .env: %v", err)
	}

	// Connect to DB
	if err := database.Connect(); err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer database.Close()

	targetEmail := os.Getenv("AGENT_AUTH_USERNAME")
	
	if targetEmail == "" {
		targetEmail = runInteractiveMode()
	}

	if targetEmail == "" {
		log.Fatal("No user selected")
	}

	authenticate(targetEmail)
}

func runInteractiveMode() string {
	var roles []roleModels.Role
	if err := database.DB.Find(&roles).Error; err != nil {
		log.Fatal("Failed to fetch roles:", err)
	}

	fmt.Println("Available Roles:")
	for i, r := range roles {
		fmt.Printf("[%d] %s (%s)\n", i+1, r.Name, r.Code)
	}

	fmt.Print("Select Role (ID): ")
	reader := bufio.NewReader(os.Stdin)
	input, _ := reader.ReadString('\n')
	input = strings.TrimSpace(input)
	
	idx, err := strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(roles) {
		log.Fatal("Invalid selection")
	}
	selectedRole := roles[idx-1]

	var users []userModels.User
	if err := database.DB.Where("role_id = ?", selectedRole.ID).Find(&users).Error; err != nil {
		log.Fatal("Failed to fetch users:", err)
	}

	if len(users) == 0 {
		log.Fatal("No users found for this role")
	}

	fmt.Printf("\nUsers in %s role:\n", selectedRole.Name)
	for i, u := range users {
		fmt.Printf("[%d] %s (%s)\n", i+1, u.Name, u.Email)
	}

	fmt.Print("Select User (ID): ")
	input, _ = reader.ReadString('\n')
	input = strings.TrimSpace(input)

	idx, err = strconv.Atoi(input)
	if err != nil || idx < 1 || idx > len(users) {
		log.Fatal("Invalid selection")
	}

	return users[idx-1].Email
}

func authenticate(email string) {
	password := os.Getenv("SEED_DEFAULT_PASSWORD")
	if password == "" {
		password = "admin123" // Default fallback
	}

	port := config.AppConfig.Server.Port
	if port == "" {
		port = "8080"
	}

	baseURL := fmt.Sprintf("http://localhost:%s/api/v1", port)
	client := &http.Client{}

	// 1. Get CSRF Token
	csrfURL := fmt.Sprintf("%s/auth/csrf", baseURL)
	respCSRF, err := client.Get(csrfURL)
	if err != nil {
		log.Fatal("Failed to get CSRF token. Is the API server running?\nError: ", err)
	}
	defer respCSRF.Body.Close()

	var csrfToken string
	cookieJar := make(map[string]string)
	for _, cookie := range respCSRF.Cookies() {
		cookieJar[cookie.Name] = cookie.Value
		if cookie.Name == "gims_csrf_token" {
			csrfToken = cookie.Value
		}
	}

	if csrfToken == "" {
		log.Fatal("Failed to receive CSRF token from server")
	}

	// 2. Perform Login
	loginURL := fmt.Sprintf("%s/auth/login", baseURL)
	payload := map[string]string{
		"email":    email,
		"password": password,
	}
	jsonPayload, _ := json.Marshal(payload)

	req, err := http.NewRequest("POST", loginURL, bytes.NewBuffer(jsonPayload))
	if err != nil {
		log.Fatal("Failed to create request:", err)
	}
	
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF-Token", csrfToken)
	
	// Add all previous cookies
	var cookieStrings []string
	for name, val := range cookieJar {
		cookieStrings = append(cookieStrings, fmt.Sprintf("%s=%s", name, val))
	}
	req.Header.Set("Cookie", strings.Join(cookieStrings, "; "))

	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Failed to send login request:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// Try to read error body
		var errorBody map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorBody)
		errJSON, _ := json.MarshalIndent(errorBody, "", "  ")
		log.Fatalf("Login failed with status: %s\nResponse: %s", resp.Status, string(errJSON))
	}

	var authCookies []string
	// The response might include the CSRF token again or update it, but we care about auth tokens
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "gims_access_token" || cookie.Name == "gims_refresh_token" {
			authCookies = append(authCookies, fmt.Sprintf("%s=%s", cookie.Name, cookie.Value))
		}
	}

	if len(authCookies) == 0 {
		log.Fatal("Login successful but no authentication cookies received")
	}

	// We also need the CSRF token for subsequent requests
	finalCookies := append(authCookies, fmt.Sprintf("gims_csrf_token=%s", csrfToken))
	cookieString := strings.Join(finalCookies, "; ")
	
	fmt.Printf("\nLogin Successful!\n")
	fmt.Printf("User: %s\n", email)
	fmt.Printf("X-CSRF-Token: %s\n", csrfToken)
	fmt.Printf("Cookie Header:\nCookie: %s\n", cookieString)
	
	// Export command hint
	fmt.Printf("\nFor curl:\nexport COOKIE=\"%s\"\nexport CSRF=\"%s\"\n", cookieString, csrfToken)
	fmt.Printf("curl -H \"Cookie: $COOKIE\" -H \"X-CSRF-Token: $CSRF\" %s/users/me\n", baseURL)
}
