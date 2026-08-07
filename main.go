package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Todo struct {
	ID          uint       `json:"id" gorm:"primaryKey"`
	Title       string     `json:"title" gorm:"not null"`
	Completed   bool       `json:"completed" gorm:"default:false"`
	Category    string     `json:"category"`
	Priority    string     `json:"priority"`
	CompletedAt *time.Time `json:"completedAt"`
	DueDate     *time.Time `json:"dueDate"`
}

var db *gorm.DB

func main() {
	dsn := "host=localhost user=postgres password=123456 dbname=todo_app port=5432 sslmode=disable"

	var err error
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("failed to connect to database: " + err.Error())
	}

	db.AutoMigrate(&Todo{})

	router := gin.Default()

	router.GET("/todos", getTodos)
	router.POST("/todos", createTodo)
	router.GET("/todos/:id", getTodoByID)
	router.PUT("/todos/:id", updateTodo)
	router.DELETE("/todos/:id", deleteTodo)

	// New endpoint
	router.GET("/todos/category/:category", getTodosByCategory)
	router.GET("/todos/status/:status", getTodosByStatus)

	// Start the server
	router.Run(":8080")
}

func getTodos(c *gin.Context) {
	var todos []Todo
	db.Find(&todos)
	c.JSON(http.StatusOK, todos)
}

func createTodo(c *gin.Context) {
	var newTodo Todo

	if err := c.ShouldBindJSON(&newTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if newTodo.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	if !isValidPriority(newTodo.Priority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Priority must be Low, Medium, or High"})
		return
	}

	if newTodo.DueDate != nil && newTodo.DueDate.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Due date cannot be in the past"})
		return
	}

	db.Create(&newTodo)
	c.JSON(http.StatusCreated, newTodo)
}

func isValidPriority(p string) bool {
	return p == "Low" || p == "Medium" || p == "High"
}
func getTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var todo Todo
	result := db.First(&todo, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	c.JSON(http.StatusOK, todo)
}

func updateTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var todo Todo
	if result := db.First(&todo, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	var updatedTodo Todo
	if err := c.ShouldBindJSON(&updatedTodo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if updatedTodo.Title == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Title cannot be empty"})
		return
	}

	if !isValidPriority(updatedTodo.Priority) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Priority must be Low, Medium, or High"})
		return
	}

	if updatedTodo.DueDate != nil && updatedTodo.DueDate.Before(time.Now()) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Due date cannot be in the past"})
		return
	}

	// Handle completedAt based on completed status changing
	if updatedTodo.Completed && !todo.Completed {
		now := time.Now().UTC()
		todo.CompletedAt = &now
	} else if !updatedTodo.Completed {
		todo.CompletedAt = nil
	}

	todo.Title = updatedTodo.Title
	todo.Completed = updatedTodo.Completed
	todo.Category = updatedTodo.Category
	todo.Priority = updatedTodo.Priority
	todo.DueDate = updatedTodo.DueDate

	db.Save(&todo)
	c.JSON(http.StatusOK, todo)
}
func deleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	var todo Todo
	if result := db.First(&todo, id); result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
		return
	}

	db.Delete(&todo)
	c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
}
func getTodosByCategory(c *gin.Context) {
	category := c.Param("category")

	var todos []Todo

	result := db.Where("category = ?", category).Find(&todos)

	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database error",
		})
		return
	}

	c.JSON(http.StatusOK, todos)
}
func getTodosByStatus(c *gin.Context) {
	status := c.Param("status")

	var todos []Todo

	switch status {
	case "completed":
		db.Where("completed = ?", true).Find(&todos)

	case "pending":
		db.Where("completed = ?", false).Find(&todos)

	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Status must be completed or pending",
		})
		return
	}

	c.JSON(http.StatusOK, todos)
}
