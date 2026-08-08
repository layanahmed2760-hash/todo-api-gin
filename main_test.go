package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

// setupTestRouter creates a fresh in-memory database and router for each test
func setupTestRouter() *gin.Engine {
	testDB, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		panic("failed to connect to test database")
	}
	testDB.AutoMigrate(&Todo{})
	db = testDB // overwrite the global db with our test one

	gin.SetMode(gin.TestMode)
	router := gin.Default()

	router.GET("/todos", getTodos)
	router.POST("/todos", createTodo)
	router.GET("/todos/search", searchTodos)
	router.DELETE("/todos", deleteAllTodos)
	router.GET("/todos/:id", getTodoByID)
	router.PUT("/todos/:id", updateTodo)
	router.DELETE("/todos/:id", deleteTodo)
	router.GET("/todos/category/:category", getTodosByCategory)
	router.PUT("/todos/category/:category", updateTodosByCategory)
	router.GET("/todos/status/:status", getTodosByStatus)

	return router
}

func TestCreateTodo_Success(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"title":    "Test todo",
		"priority": "Low",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Errorf("expected status 201, got %d", w.Code)
	}
}

func TestCreateTodo_EmptyTitle(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"title":    "",
		"priority": "Low",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestCreateTodo_InvalidPriority(t *testing.T) {
	router := setupTestRouter()

	body := map[string]interface{}{
		"title":    "Test",
		"priority": "Urgent",
	}
	jsonBody, _ := json.Marshal(body)

	req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", w.Code)
	}
}

func TestGetTodoByID_NotFound(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("GET", "/todos/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}
func TestUpdateTodo_SetsCompletedAt(t *testing.T) {
	router := setupTestRouter()

	// First create a todo
	createBody := map[string]interface{}{
		"title":    "Finish task",
		"priority": "Medium",
	}
	jsonBody, _ := json.Marshal(createBody)
	req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var created Todo
	json.Unmarshal(w.Body.Bytes(), &created)

	// Now mark it completed
	updateBody := map[string]interface{}{
		"title":     "Finish task",
		"priority":  "Medium",
		"completed": true,
	}
	jsonBody2, _ := json.Marshal(updateBody)
	req2, _ := http.NewRequest("PUT", "/todos/1", bytes.NewBuffer(jsonBody2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	var updated Todo
	json.Unmarshal(w2.Body.Bytes(), &updated)

	if updated.CompletedAt == nil {
		t.Error("expected completedAt to be set, got nil")
	}
}

func TestDeleteTodo_NotFound(t *testing.T) {
	router := setupTestRouter()

	req, _ := http.NewRequest("DELETE", "/todos/999", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", w.Code)
	}
}

func TestSearchTodos_CaseInsensitive(t *testing.T) {
	router := setupTestRouter()

	// Create a todo with a specific title
	createBody := map[string]interface{}{
		"title":    "Write Report",
		"priority": "Low",
	}
	jsonBody, _ := json.Marshal(createBody)
	req, _ := http.NewRequest("POST", "/todos", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Search using different casing
	req2, _ := http.NewRequest("GET", "/todos/search?q=REPORT", nil)
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w2.Code)
	}

	var results []Todo
	json.Unmarshal(w2.Body.Bytes(), &results)

	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}