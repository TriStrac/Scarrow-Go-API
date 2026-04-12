package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TriStrac/Scarrow-Go-API/internal/api/controllers"
	"github.com/gin-gonic/gin"
)

// Mock services to bypass actual database/SMS calls
type mockUserService struct{}
// ... (implement needed mock methods later if we want a full test, but for now we just want to test if the DTO binding works)

func TestRegisterBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Since we only want to test the Gin binding logic for the DTO without full mocks, 
	// we will create a dummy route that uses the same binding to verify.
	router := gin.Default()
	router.POST("/test-register", func(ctx *gin.Context) {
		var req controllers.RegisterReq
		if err := ctx.ShouldBindJSON(&req); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, gin.H{"message": "binding successful"})
	})

	// Test case 1: Minimal payload (should pass)
	payload := map[string]string{
		"username": "testuser",
		"password": "password123",
		"number":   "09123456789",
	}
	body, _ := json.Marshal(payload)
	
	req, _ := http.NewRequest(http.MethodPost, "/test-register", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d with body: %s", w.Code, w.Body.String())
	}

	// Test case 2: Missing required field (should fail)
	payloadMissing := map[string]string{
		"username": "testuser",
		"password": "password123",
	}
	bodyMissing, _ := json.Marshal(payloadMissing)
	
	reqMissing, _ := http.NewRequest(http.MethodPost, "/test-register", bytes.NewBuffer(bodyMissing))
	reqMissing.Header.Set("Content-Type", "application/json")
	wMissing := httptest.NewRecorder()
	
	router.ServeHTTP(wMissing, reqMissing)

	if wMissing.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400 for missing number, got %d", wMissing.Code)
	}
}
