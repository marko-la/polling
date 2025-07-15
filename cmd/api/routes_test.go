package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"polling/internal/models"
	"polling/internal/repository/mocks"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
)

type TestAppConfig struct {
	ShouldFail bool
	MockUser   *models.User
	MockError  error
}

func setuptestApp(cfg TestAppConfig) *application {
	mockRepo := &mocks.MockDBRepo{
		ShouldFail: cfg.ShouldFail,
		MockUser:   cfg.MockUser,
		MockError:  cfg.MockError}

	mockAuth := Auth{
		Issuer:        "test",
		Audience:      "test",
		Secret:        "test_secret",
		TokenExpiry:   time.Hour,
		RefreshExpiry: time.Hour * 24,
		CookieDomain:  "localhost",
		CookiePath:    "/",
		CookieName:    "refresh_token",
	}

	return &application{
		DB:   mockRepo,
		auth: mockAuth,
	}
}

func generateTestJWT(auth Auth, userID int) (string, error) {
	user := &jwtUser{
		ID:        userID,
		FirstName: "Test",
		LastName:  "User",
	}
	tokenPair, err := auth.GenerateTokenPair(user)
	if err != nil {
		return "", err
	}
	return tokenPair.Token, nil
}

func addURLParamToRequest(req *http.Request, key, value string) *http.Request {
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add(key, value)
	return req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
}

func TestSignup(t *testing.T) {
	// Define test cases
	tests := []struct {
		name            string
		payload         map[string]string
		dbShouldFail    bool
		expectedStatus  int
		expectError     bool
		expectedMessage string
	}{
		{
			name: "valid signup",
			payload: map[string]string{
				"username":   "testuser",
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
			},
			dbShouldFail:    false,
			expectedStatus:  http.StatusOK,
			expectError:     false,
			expectedMessage: "User successfuly created",
		},
		{
			name: "missing username",
			payload: map[string]string{
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
			},
			dbShouldFail:    false,
			expectedStatus:  http.StatusBadRequest,
			expectError:     true,
			expectedMessage: "missing one or more required field ['username','password','first_name', 'last_name']",
		},
		{
			name: "missing password",
			payload: map[string]string{
				"username":   "testuser",
				"first_name": "John",
				"last_name":  "Doe",
			},
			dbShouldFail:    false,
			expectedStatus:  http.StatusBadRequest,
			expectError:     true,
			expectedMessage: "missing one or more required field ['username','password','first_name', 'last_name']",
		},
		{
			name: "database error",
			payload: map[string]string{
				"username":   "testuser",
				"password":   "password123",
				"first_name": "John",
				"last_name":  "Doe",
			},
			dbShouldFail:    true,
			expectedStatus:  http.StatusBadRequest,
			expectError:     true,
			expectedMessage: "database error",
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setuptestApp(TestAppConfig{ShouldFail: tt.dbShouldFail})

			// Create request body
			jsonPayload, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/signup", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			app.Signup(rr, req)

			// Assert status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check if response body contains error when expected
			if tt.expectError {
				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["error"]; !exists {
					t.Error("Expected error in response but got none")
				}

				// Check specific error message
				if errorMsg, exists := response["error"].(string); exists {
					if errorMsg != tt.expectedMessage {
						t.Errorf("Expected error message '%s', got '%s'", tt.expectedMessage, errorMsg)
					}
				}
			} else {
				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["message"]; !exists {
					t.Error("Expected success message in response")
				}

				// Check specific success message
				if successMsg, exists := response["message"].(string); exists {
					if successMsg != tt.expectedMessage {
						t.Errorf("Expected success message '%s', got '%s'", tt.expectedMessage, successMsg)
					}
				}
			}
		})
	}
}

func TestLogin(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		payload        map[string]string
		mockUser       *models.User
		mockError      error
		expectedStatus int
		expectError    bool
		checkTokens    bool
	}{
		{
			name: "valid login",
			payload: map[string]string{
				"username": "testuser",
				"password": "password123",
			},
			mockUser: &models.User{
				ID:        1,
				Username:  "testuser",
				Password:  "$2a$12$PVb44Is66aXVFtMBu5yzx.FR1QeF1vEJ9iAxIuDhs9ZltkaOUaKCy", // bcrypt hash for "password123"
				FirstName: "John",
				LastName:  "Doe",
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectError:    false,
			checkTokens:    true,
		},
		{
			name: "missing username",
			payload: map[string]string{
				"password": "password123",
			},
			mockUser:       nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			checkTokens:    false,
		},
		{
			name: "missing password",
			payload: map[string]string{
				"username": "testuser",
			},
			mockUser:       nil,
			mockError:      nil,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			checkTokens:    false,
		},
		{
			name: "user not found",
			payload: map[string]string{
				"username": "nonexistent",
				"password": "password123",
			},
			mockUser:       nil,
			mockError:      errors.New("user not found"),
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
			checkTokens:    false,
		},
		{
			name: "wrong password",
			payload: map[string]string{
				"username": "testuser",
				"password": "wrongpassword",
			},
			mockUser: &models.User{
				ID:        1,
				Username:  "testuser",
				Password:  "$2a$12$example_hashed_password",
				FirstName: "John",
				LastName:  "Doe",
			},
			mockError:      nil,
			expectedStatus: http.StatusUnauthorized,
			expectError:    true,
			checkTokens:    false,
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setuptestApp(TestAppConfig{ShouldFail: false, MockUser: tt.mockUser, MockError: tt.mockError})

			// Create request body
			jsonPayload, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/login", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			app.Login(rr, req)

			// Assert status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response based on expectations
			if tt.expectError {
				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["error"]; !exists {
					t.Error("Expected error in response but got none")
				}
			} else if tt.checkTokens {
				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				// Check if user data is present
				if _, exists := response["user"]; !exists {
					t.Error("Expected user data in response")
				}

				// Check if tokens are present
				if _, exists := response["tokens"]; !exists {
					t.Error("Expected tokens in response")
				}

				// Check if refresh cookie is set
				cookies := rr.Result().Cookies()
				refreshCookieFound := false
				for _, cookie := range cookies {
					if cookie.Name == "refresh_token" { // Adjust cookie name as needed
						refreshCookieFound = true
						break
					}
				}
				if !refreshCookieFound {
					t.Error("Expected refresh cookie to be set")
				}
			}
		})
	}
}

func TestCreatePoll(t *testing.T) {
	tests := []struct {
		name            string
		payload         map[string]any
		userID          int
		withAuth        bool
		expectedStatus  int
		expectError     bool
		expectedMessage string
	}{
		{
			name: "valid poll creation",
			payload: map[string]any{
				"title":       "Test Poll",
				"description": "This is a test poll",
			},
			userID:          1,
			withAuth:        true,
			expectedStatus:  http.StatusOK,
			expectError:     false,
			expectedMessage: "Poll created successfully",
		},
		{
			name: "missing title",
			payload: map[string]any{
				"description": "This is a test poll",
			},
			userID:          1,
			withAuth:        true,
			expectedStatus:  http.StatusBadRequest,
			expectError:     true,
			expectedMessage: "missing required fields",
		},
		{
			name: "unauthorized request",
			payload: map[string]any{
				"title":       "Test Poll",
				"description": "This is a test poll",
			},
			userID:          0,
			withAuth:        false,
			expectedStatus:  http.StatusUnauthorized,
			expectError:     true,
			expectedMessage: "unauthorized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setuptestApp(TestAppConfig{})

			// Create request body
			jsonPayload, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", "/polls/create", bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			// Add JWT token if authenticated
			if tt.withAuth {
				token, err := generateTestJWT(app.auth, tt.userID)
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			authHandler := app.authRequired(http.HandlerFunc(app.CreatePoll))
			authHandler.ServeHTTP(rr, req)

			// Assert status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response based on expectations
			if tt.expectError {
				if tt.expectedStatus == http.StatusUnauthorized {
					// Auth middleware returns empty body for unauthorized requests
					if rr.Body.Len() == 0 {
						// This is expected for auth middleware rejection
						return
					}
				}

				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["error"]; !exists {
					t.Error("Expected error in response but got none")
				}
			} else {
				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}

				// For successful poll creation, we expect the poll object back
				if response == nil {
					t.Error("Expected response data for successful poll creation")
				}

				// Check if poll data is present
				if _, exists := response["id"]; !exists {
					t.Error("Expected poll ID in response")
				}
				if title, exists := response["title"]; !exists || title != tt.payload["title"] {
					t.Errorf("Expected title '%s' in response, got '%v'", tt.payload["title"], title)
				}
			}
		})
	}
}

func TestRemovePoll(t *testing.T) {
	tests := []struct {
		name            string
		pollID          int
		userID          int
		withAuth        bool
		expectedStatus  int
		expectError     bool
		expectedMessage string
		dbShouldFail    bool
	}{
		{
			name:            "valid delete",
			pollID:          1,
			userID:          1,
			withAuth:        true,
			expectedStatus:  http.StatusOK,
			expectError:     false,
			expectedMessage: "Poll deleted",
			dbShouldFail:    false,
		}, {
			name:            "invalid poll",
			pollID:          2,
			userID:          1,
			withAuth:        true,
			expectedStatus:  http.StatusBadRequest,
			expectError:     true,
			expectedMessage: "database error",
			dbShouldFail:    false,
		}, {
			name:            "unathuroized",
			pollID:          1,
			userID:          2,
			withAuth:        true,
			expectedStatus:  http.StatusUnauthorized,
			expectError:     true,
			expectedMessage: "you are not authorized to update this poll",
			dbShouldFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup
			app := setuptestApp(TestAppConfig{ShouldFail: tt.dbShouldFail})

			// Create HTTP request
			req := httptest.NewRequest("DELETE", fmt.Sprintf("/polls/%d", tt.pollID), nil)
			req.Header.Set("Content-Type", "application/json")

			// Add poll ID to chi URL params context
			req = addURLParamToRequest(req, "pollID", fmt.Sprintf("%d", tt.pollID))

			// Add JWT token if authenticated
			if tt.withAuth {
				// Generate a real JWT token
				token, err := generateTestJWT(app.auth, tt.userID)
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			authHandler := app.authRequired(http.HandlerFunc(app.RemovePoll))
			authHandler.ServeHTTP(rr, req)

			// Assert status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response based on expectations
			if tt.expectError {
				if tt.expectedStatus == http.StatusUnauthorized {
					// Auth middleware returns empty body for unauthorized requests
					if rr.Body.Len() == 0 {
						// This is expected for auth middleware rejection
						return
					}
				}

				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["error"]; !exists {
					t.Error("Expected error in response but got none")
				}
				if _, exists := response["message"]; !exists {
					t.Error("Expected message in response but got none")
				} else if response["message"] != tt.expectedMessage {
					t.Error("Got wrong message in response")
				}

			} else {

				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["message"]; !exists {
					t.Error("Expected message in response but got none")
				} else if response["message"] != tt.expectedMessage {
					t.Error("Got wrong message in response")
				}
			}

		})

	}
}

func TestAddPollOptions(t *testing.T) {
	tests := []struct {
		name            string
		pollID          int
		userID          int
		payload         struct{ Options []models.PollOption }
		withAuth        bool
		expectedStatus  int
		expectError     bool
		expectedMessage string
		dbShouldFail    bool
	}{
		{
			name:   "valid poll options",
			pollID: 1,
			userID: 1,
			payload: struct{ Options []models.PollOption }{
				Options: []models.PollOption{
					{Text: "option1"},
					{Text: "option2"},
				},
			},
			withAuth:        true,
			expectedStatus:  http.StatusOK,
			expectError:     false,
			expectedMessage: "options added successfully!",
			dbShouldFail:    false,
		}, {
			name:   "database error",
			pollID: 2,
			userID: 1,
			payload: struct{ Options []models.PollOption }{
				Options: []models.PollOption{
					{Text: "option1"},
					{Text: "option2"},
				},
			},
			withAuth:        true,
			expectedStatus:  http.StatusBadRequest,
			expectError:     true,
			expectedMessage: "database error",
			dbShouldFail:    false,
		}, {
			name:   "invalid owner",
			pollID: 1,
			userID: 2,
			payload: struct{ Options []models.PollOption }{
				Options: []models.PollOption{
					{Text: "option1"},
					{Text: "option2"},
				},
			},
			withAuth:        true,
			expectedStatus:  http.StatusUnauthorized,
			expectError:     true,
			expectedMessage: "you are not authorized to update this poll",
			dbShouldFail:    false,
		}, {
			name:   "unauthorized",
			pollID: 1,
			userID: 2,
			payload: struct{ Options []models.PollOption }{
				Options: []models.PollOption{
					{Text: "option1"},
					{Text: "option2"},
				},
			},
			withAuth:        false,
			expectedStatus:  http.StatusUnauthorized,
			expectError:     true,
			expectedMessage: "",
			dbShouldFail:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := setuptestApp(TestAppConfig{ShouldFail: tt.dbShouldFail})

			// Create request body
			jsonPayload, err := json.Marshal(tt.payload)
			if err != nil {
				t.Fatalf("Failed to marshal payload: %v", err)
			}

			// Create HTTP request
			req := httptest.NewRequest("POST", fmt.Sprintf("/polls/%d/options", tt.pollID), bytes.NewBuffer(jsonPayload))
			req.Header.Set("Content-Type", "application/json")

			// Add poll ID to chi URL params context
			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("pollID", fmt.Sprintf("%d", tt.pollID))
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			// Add JWT token if authenticated
			if tt.withAuth {
				// Generate a real JWT token
				token, err := generateTestJWT(app.auth, tt.userID)
				if err != nil {
					t.Fatalf("Failed to generate token: %v", err)
				}
				req.Header.Set("Authorization", "Bearer "+token)
			}

			// Create response recorder
			rr := httptest.NewRecorder()

			authHandler := app.authRequired(http.HandlerFunc(app.AddPollOptions))
			authHandler.ServeHTTP(rr, req)

			// Assert status code
			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			// Check response based on expectations
			if tt.expectError {
				if tt.expectedStatus == http.StatusUnauthorized {
					// Auth middleware returns empty body for unauthorized requests
					if rr.Body.Len() == 0 {
						// This is expected for auth middleware rejection
						return
					}
				}

				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["error"]; !exists {
					t.Error("Expected error in response but got none")
				}
				if _, exists := response["message"]; !exists {
					t.Error("Expected message in response but got none")
				} else if response["message"] != tt.expectedMessage {
					t.Error("Got wrong message in response")
				}

			} else {

				var response map[string]any
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Errorf("Failed to parse response: %v", err)
				}
				if _, exists := response["message"]; !exists {
					t.Error("Expected message in response but got none")
				} else if response["message"] != tt.expectedMessage {
					t.Error("Got wrong message in response")
				}
			}

		})

	}

}
