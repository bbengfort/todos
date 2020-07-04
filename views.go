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

// Overview returns statistics for the authenticated user, e.g. how many tasks and lists
// are currently open, completed, and archived. Although this is the root view of the
// API, this view requires an authenticated user in the context.
func (s *API) Overview(c *gin.Context) {
	user := c.Value(ctxUserKey).(User)
	c.JSON(http.StatusOK, OverviewResponse{Success: true, User: user.Username})
}

//===========================================================================
// Viewset for Task objects
//===========================================================================

// ListTasks returns all tasks for the authenticated user, sorted and filtered by the
// specified input parameters (e.g. by list or by most recent).
// TODO: add filtering by list
// TODO: add pagination
func (s *API) ListTasks(c *gin.Context) {
	var req ListTasksRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	user := c.Value(ctxUserKey).(User)
	var tasks []Task
	if err := s.db.Where("user_id = ?", user.ID).Find(&tasks).Error; err != nil {
		logger.Printf("could not fetch tasks: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	c.JSON(http.StatusOK, ListTasksResponse{Success: true, Tasks: tasks})
}

// CreateTask creates a new task assigned to the authenticated user in the database.
func (s *API) CreateTask(c *gin.Context) {
	// Parse the user input
	task := Task{}
	if err := c.ShouldBind(&task); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	// Add the user to the task
	user := c.Value(ctxUserKey).(User)
	task.UserID = user.ID

	// Create the task in the database
	if err := s.db.Create(&task).Error; err != nil {
		logger.Printf("could not create task: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	c.JSON(http.StatusCreated, CreateTaskResponse{Success: true, TaskID: task.ID})
}

// DetailTask returns as much information about the task as possible.
// TODO: ensure that the task belongs to the user!
func (s *API) DetailTask(c *gin.Context) {
	var task Task
	if err := s.db.Where("id = ?", c.Param("id")).First(&task).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, notFound)
			return
		}
		logger.Printf("could not find task: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	c.JSON(http.StatusOK, DetailTaskResponse{Success: true, Task: task})
}

// UpdateTask allows the user to modify a task.
// TODO: ensure that the task belongs to the user!
func (s *API) UpdateTask(c *gin.Context) {
	// Fetch the task to update
	task := Task{}
	if err := s.db.Where("id = ?", c.Param("id")).First(&task).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, notFound)
			return
		}
		logger.Printf("could not find task: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	// Parse the user input
	// In order to set zero values (e.g. completed/archived as false) input needs to be a map not a struct
	var input map[string]interface{}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	if err := s.db.Model(&task).Update(input).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, UpdateTaskResponse{Success: true})
}

// DeleteTask removes the task from the database.
// TODO: ensure that the task belongs to the user!
func (s *API) DeleteTask(c *gin.Context) {
	var task Task
	if err := s.db.Where("id = ?", c.Param("id")).First(&task).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, notFound)
			return
		}
		logger.Printf("could not find task: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	if err := s.db.Delete(&task).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, DeleteTaskResponse{Success: true})
}

//===========================================================================
// Viewset for List objects
//===========================================================================

// ListChecklists returns all checklists that belong to the authenticated user.
// TODO: add pagination
func (s *API) ListChecklists(c *gin.Context) {
	var req ListChecklistsRequest
	if err := c.ShouldBind(&req); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	var lists []Checklist
	user := c.Value(ctxUserKey).(User)

	if err := s.db.Where("user_id = ?", user.ID).Find(&lists).Error; err != nil {
		logger.Printf("could not fetch checklists: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	c.JSON(http.StatusOK, ListChecklistsResponse{Success: true, Checklists: lists})
}

// CreateChecklist creates a new grouping of tasks for the user.
func (s *API) CreateChecklist(c *gin.Context) {
	// Parse the user input
	list := Checklist{}
	if err := c.ShouldBind(&list); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	// Add the user to the list
	user := c.Value(ctxUserKey).(User)
	list.UserID = user.ID

	// Create the checklist in the database
	if err := s.db.Create(&list).Error; err != nil {
		logger.Printf("could not create list: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	c.JSON(http.StatusCreated, CreateChecklistResponse{Success: true, ChecklistID: list.ID})
}

// DetailChecklist gives as many details about the checklist as possible.
// TODO: ensure that the list belongs to the user!
func (s *API) DetailChecklist(c *gin.Context) {
	var list Checklist
	if err := s.db.Where("id = ?", c.Param("id")).First(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, notFound)
			return
		}
		logger.Printf("could not find list: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	c.JSON(http.StatusOK, DetailChecklistResponse{Success: true, Checklist: list})
}

// UpdateChecklist modifies the database checklist.
// TODO: ensure that the list belongs to the user!
func (s *API) UpdateChecklist(c *gin.Context) {
	// Fetch the list to update
	list := Checklist{}
	if err := s.db.Where("id = ?", c.Param("id")).First(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, notFound)
			return
		}
		logger.Printf("could not find checklist: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	// Parse user input
	// In order to set zero values, input needs to be a map, not a struct
	var input map[string]interface{}
	if err := c.ShouldBind(&input); err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	if err := s.db.Model(&list).Update(input).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, UpdateChecklistResponse{Success: true})
}

// DeleteChecklist removes the checklist from the database and all associated tasks.
// TODO: ensure the list belongs to the user!
func (s *API) DeleteChecklist(c *gin.Context) {
	var list Checklist
	if err := s.db.Where("id = ?", c.Param("id")).First(&list).Error; err != nil {
		if gorm.IsRecordNotFoundError(err) {
			c.JSON(http.StatusNotFound, notFound)
			return
		}
		logger.Printf("could not find checklist: %s", err)
		c.JSON(http.StatusInternalServerError, ErrorResponse(nil))
		return
	}

	if err := s.db.Delete(&list).Error; err != nil {
		c.JSON(http.StatusBadRequest, ErrorResponse(err))
		return
	}

	c.JSON(http.StatusOK, DeleteChecklistResponse{Success: true})
}
