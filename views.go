package todos

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
)

// Context keys for middleware lookups
const (
	ctxUserKey = "user"
)

// Overview returns the state of the todos (e.g. the number of tasks and lists for the
// given user). This request must be authenticated.
func (s *API) Overview(c *gin.Context) {
	user := c.Value(ctxUserKey).(User)
	c.JSON(http.StatusOK, gin.H{"tasks": 0, "lists": 0, "user": user.Username})
}

//===========================================================================
// Viewset for Todo objects
//===========================================================================

// FindTodos returns a all todos for the user, sorted and filtered by the specified
// input parameters (e.g. by list or by most recent).
// TODO: add filtering by list
// TODO: add pagination
func (s *API) FindTodos(c *gin.Context) {
	user := c.Value(ctxUserKey).(User)
	var todos []Todo
	if err := s.db.Where("user_id = ?", user.ID).Find(&todos).Error; err != nil {
		logger.Printf("could not fetch todos: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"todos": todos})
}

// CreateTodo creates a new todo in the database.
func (s *API) CreateTodo(c *gin.Context) {
	// Parse the user input
	todo := Todo{}
	if err := c.ShouldBind(&todo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Add the user to the todo
	user := c.Value(ctxUserKey).(User)
	todo.UserID = user.ID

	// Create the TODO in the database
	if err := s.db.Create(&todo).Error; err != nil {
		logger.Printf("could not create todo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "todo": todo.ID})
}

// DetailTodo returns as much information about the todo as possible.
func (s *API) DetailTodo(c *gin.Context) {
	var todo Todo
	if err := s.db.Where("id = ?", c.Param("id")).First(&todo).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		logger.Printf("could not find todo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "todo": todo})
}

// UpdateTodo allows the user to modify a todo.
func (s *API) UpdateTodo(c *gin.Context) {
	// Parse the user input
	todo := Todo{}
	input := updateTodo{}

	if err := s.db.Where("id = ?", c.Param("id")).First(&todo).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		logger.Printf("could not find todo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	// TODO: right now this does not allow us to set completed/archived as false
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := s.db.Model(&todo).Update(input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteTodo removes the todo from the database.
func (s *API) DeleteTodo(c *gin.Context) {
	var todo Todo
	if err := s.db.Where("id = ?", c.Param("id")).First(&todo).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		logger.Printf("could not find todo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	if err := s.db.Delete(&todo).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

//===========================================================================
// Viewset for List objects
//===========================================================================

// FindLists returns all lists that belong to the user.
func (s *API) FindLists(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// CreateList creates a new grouping of todos for the user.
func (s *API) CreateList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DetailList gives as many details about the todo list as possible.
func (s *API) DetailList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// UpdateList modifies the database list.
func (s *API) UpdateList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteList removes the database and all associated todos.
func (s *API) DeleteList(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"success": true})
}
