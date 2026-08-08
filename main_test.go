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