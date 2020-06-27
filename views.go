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

	// Create the todo in the database
	if err := s.db.Create(&todo).Error; err != nil {
		logger.Printf("could not create todo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "todo": todo.ID})
}

// DetailTodo returns as much information about the todo as possible.
// TODO: ensure that the todo belongs to the user!
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
// TODO: ensure that the todo belongs to the user!
func (s *API) UpdateTodo(c *gin.Context) {
	// Fetch the todo to update
	todo := Todo{}
	if err := s.db.Where("id = ?", c.Param("id")).First(&todo).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		logger.Printf("could not find todo: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	// Parse the user input
	// In order to set zero values (e.g. completed/archived as false) input needs to be a map not a struct
	var input map[string]interface{}
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
// TODO: ensure that the todo belongs to the user!
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
// TODO: add pagination
func (s *API) FindLists(c *gin.Context) {
	user := c.Value(ctxUserKey).(User)
	var lists []List
	if err := s.db.Where("user_id = ?", user.ID).Find(&lists).Error; err != nil {
		logger.Printf("could not fetch lists: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"lists": lists})
}

// CreateList creates a new grouping of todos for the user.
func (s *API) CreateList(c *gin.Context) {
	// Parse the user input
	list := List{}
	if err := c.ShouldBind(&list); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	// Add the user to the list
	user := c.Value(ctxUserKey).(User)
	list.UserID = user.ID

	// Create the list in the database
	if err := s.db.Create(&list).Error; err != nil {
		logger.Printf("could not create list: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"success": true, "list": list.ID})
}

// DetailList gives as many details about the todo list as possible.
// TODO: ensure that the list belongs to the user!
func (s *API) DetailList(c *gin.Context) {
	var list List
	if err := s.db.Where("id = ?", c.Param("id")).First(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		logger.Printf("could not find list: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true, "list": list})
}

// UpdateList modifies the database list.
// TODO: ensure that the list belongs to the user!
func (s *API) UpdateList(c *gin.Context) {
	// Fetch the list to update
	list := List{}
	if err := s.db.Where("id = ?", c.Param("id")).First(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		logger.Printf("could not find list: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	// Parse user input
	// In order to set zero values, input needs to be a map, not a struct
	var input map[string]interface{}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	if err := s.db.Model(&list).Update(input).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DeleteList removes the list from the database and all associated todos.
// TODO: ensure the list belongs to the user!
func (s *API) DeleteList(c *gin.Context) {
	var list List
	if err := s.db.Where("id = ?", c.Param("id")).First(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "not found"})
			return
		}
		logger.Printf("could not find list: %s", err)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false})
		return
	}

	if err := s.db.Delete(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
