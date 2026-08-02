package main

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Todo struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	Title     string `json:"title" gorm:"not null"`
	Completed bool   `json:"completed" gorm:"default:false"`
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

	db.Create(&newTodo)
	c.JSON(http.StatusCreated, newTodo)
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

	todo.Title = updatedTodo.Title
	todo.Completed = updatedTodo.Completed
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