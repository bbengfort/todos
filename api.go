package todos

import "time"

//===========================================================================
// Top Level Requests and Responses
//===========================================================================

// Response contains standard fields that are embedded in most API responses
type Response struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// StatusResponse is returned on status requests. Note that no request is needed.
type StatusResponse struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp,omitempty"`
	Version   string    `json:"version,omitempty"`
	Error     string    `json:"error,omitempty"`
}

// OverviewResponse is returned on an overview request.
type OverviewResponse struct {
	Success    bool   `json:"success"`
	Error      string `json:"error,omitempty"`
	User       string `json:"user"`
	Tasks      int    `json:"tasks"`
	Checklists int    `json:"checklists"`
}

//===========================================================================
// Authentication
//===========================================================================

// RegisterRequest allows a administrative users to create new users.
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
	IsAdmin  bool   `json:"is_admin"`
}

// RegisterResponse returns the status of a a Register request.
type RegisterResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Username string `json:"username"`
}

// LoginRequest to authenticate a user with the service and return tokens
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	NoCookie bool   `json:"no_cookie"`
}

// LoginResponse is returned on a successful login
type LoginResponse struct {
	Success      bool   `json:"success"`
	Error        string `json:"error,omitempty"`
	AccessToken  string `json:"access_token,omitempty"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// LogoutRequest to logout the current user and optionally revoke all logins. Must be
// authenticated to log out a user.
type LogoutRequest struct {
	RevokeAll bool `json:"revoke_all"`
}

// RefreshRequest is a reauthorization with a request token rather than username/password
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
	NoCookie     bool   `json:"no_cookie"`
}

//===========================================================================
// Tasks RESTful API
//===========================================================================

// ListTasksRequest fetches tasks with specific filters.
type ListTasksRequest struct {
	Checklist uint `json:"checklist,omitempty"`
	Page      int  `json:"page,omitempty"`
	PerPage   int  `json:"per_page,omitempty"`
}

// ListTasksResponse returns the tasks, and response info such as pagination.
type ListTasksResponse struct {
	Success  bool   `json:"success"`
	Error    string `json:"error,omitempty"`
	Tasks    []Task `json:"tasks,omitempty"`
	Page     int    `json:"page,omitempty"`
	NumPages int    `json:"num_pages,omitempty"`
}

// CreateTaskResponse returns the information about the created task. Currently the
// CreateTaskRequest is simply the task object itself.
type CreateTaskResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	TaskID  uint   `json:"task,omitempty"`
}

// DetailTaskResponse returns the detailed information about the task. Currently there
// is no DetailTaskRequest, the request is in the URL.
type DetailTaskResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
	Task    Task   `json:"task"`
}

// UpdateTaskResponse returns information about the update call. Currently there is no
// UpdateTaskRequest, because it is simply the task object itself.
type UpdateTaskResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DeleteTaskResponse returns information about the delete call. Currently there is no
// DeleteTaskRequest, because the request is in the URL.
type DeleteTaskResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

//===========================================================================
// Checklist RESTful API
//===========================================================================

// ListChecklistsRequest fetches checklists with specific filters.
type ListChecklistsRequest struct {
	Page    int `json:"page,omitempty"`
	PerPage int `json:"per_page,omitempty"`
}

// ListChecklistsResponse returns the checklists, and response info such as pagination.
type ListChecklistsResponse struct {
	Success    bool        `json:"success"`
	Error      string      `json:"error,omitempty"`
	Checklists []Checklist `json:"checklists,omitempty"`
	Page       int         `json:"page,omitempty"`
	NumPages   int         `json:"num_pages,omitempty"`
}

// CreateChecklistResponse returns the information about the created checklist.
// Currently the CreateChecklistRequest is simply the checklist object itself.
type CreateChecklistResponse struct {
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
	ChecklistID uint   `json:"checklist,omitempty"`
}

// DetailChecklistResponse returns the detailed information about the checklist.
// Currently there is no DetailChecklistRequest, the request is in the URL.
type DetailChecklistResponse struct {
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	Checklist Checklist `json:"checklist"`
}

// UpdateChecklistResponse returns information about the update call. Currently there is
// no UpdateChecklistRequest, because it is simply the checklist object itself.
type UpdateChecklistResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}

// DeleteChecklistResponse returns information about the delete call. Currently there is
// no DeleteChecklistRequest, because the request is in the URL.
type DeleteChecklistResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error,omitempty"`
}
