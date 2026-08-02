package main

import (
	"net/http"
"strconv"

	"github.com/gin-gonic/gin"
)

// todo item 
type Todo struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
Completed bool   `json:"completed"`
}

// global 
var todos = []Todo{}
var nextID = 1

func main() {
	router := gin.Default()

	// routes
	router.GET("/todos", getTodos)
	router.POST("/todos", createTodo)
	router.GET("/todos/:id", getTodoByID)
	router.PUT("/todos/:id", updateTodo)
	router.DELETE("/todos/:id", deleteTodo)

	router.Run(":8080") // hardcoded port IDC
}

// fetch all
func getTodos(c *gin.Context) {
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

	newTodo.ID = nextID
	nextID++
	todos = append(todos, newTodo)

	c.JSON(http.StatusCreated, newTodo)
}

// TODO: refactor this loop 
func getTodoByID(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}

	for _, todo := range todos {
		if todo.ID == id {
			c.JSON(http.StatusOK, todo)
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}

func updateTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
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

	for i, todo := range todos {
		if todo.ID == id {
			todos[i].Title = updatedTodo.Title
			todos[i].Completed = updatedTodo.Completed
			c.JSON(http.StatusOK, todos[i])
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}

// delete endpoint -- hope slice re-slicing
func deleteTodo(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID"})
		return
	}


	for i, todo := range todos {
		if todo.ID == id {
			todos = append(todos[:i], todos[i+1:]...) // remove item
			c.JSON(http.StatusOK, gin.H{"message": "Todo deleted"})
			return
		}
	}

	c.JSON(http.StatusNotFound, gin.H{"error": "Todo not found"})
}