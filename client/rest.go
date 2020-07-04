package client

import (
	"fmt"
	"net/http"

	"github.com/bbengfort/todos"
)

// ListTasks returns all tasks for the authenticated user, sorted and filtered by the
// input request. This function checks the response for errors but does not otherwise
// modify the output response. User authentication is required.
func (c *Client) ListTasks(in *todos.ListTasksRequest) (out *todos.ListTasksResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, "/tasks", true, in); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if status != http.StatusOK || !out.Success {
		return out, StatusError(status, out.Error)
	}

	return out, nil
}

// CreateTask posts the task to the server in order to create it. This function checks
// the response for errors, but does not otherwise modify the output response. User
// authentication is required.
func (c *Client) CreateTask(in *todos.Task) (out *todos.CreateTaskResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/tasks", true, in); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if !(status == http.StatusOK || status == http.StatusCreated) || !out.Success {
		return out, StatusError(status, out.Error)
	}

	return out, nil
}

// DetailTask returns as much information as possible about the specified task. This
// function checks the response for errors, but does not otherwise modify the output
// response. User authentication is required.
func (c *Client) DetailTask(id uint) (out *todos.DetailTaskResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", id), true, nil); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if status != http.StatusOK || !out.Success {
		return out, StatusError(status, out.Error)
	}

	return out, nil
}

// UpdateTask puts the task info to the specified id in order to update it. This
// function checks the response for errors, but does not otherwise modify the output
// response. User authentication is required.
func (c *Client) UpdateTask(id uint, task *todos.Task) (out *todos.UpdateTaskResponse, err error) {
	if id == 0 || (task.ID > 0 && id != task.ID) {
		return nil, fmt.Errorf("cannot update with id %d and task id %d", id, task.ID)
	}

	// Ensure that the task ID is a zero value.
	task.ID = 0

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPut, fmt.Sprintf("/tasks/%d", id), true, task); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) || !out.Success {
		return out, StatusError(status, out.Error)
	}
	return out, nil
}

// DeleteTask sends a delete request for the specified id. This function checks the
// response for errors, but does not otherwise modify the output response. User
// authentication is required.
func (c *Client) DeleteTask(id uint) (out *todos.DeleteTaskResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodDelete, fmt.Sprintf("/tasks/%d", id), true, nil); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) || !out.Success {
		return out, StatusError(status, out.Error)
	}
	return out, nil
}

// ListChecklists returns all checklists for the authenticated user, sorted and filtered
// by the input request. This function checks the response for errors but does not
// otherwise modify the output response. User authentication is required.
func (c *Client) ListChecklists(in *todos.ListChecklistsRequest) (out *todos.ListChecklistsResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, "/lists", true, in); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if status != http.StatusOK || !out.Success {
		return out, StatusError(status, out.Error)
	}

	return out, nil
}

// CreateChecklist posts the checklist to the server in order to create it. This function
// checks the response for errors, but does not otherwise modify the output response.
// User authentication is required.
func (c *Client) CreateChecklist(in *todos.Checklist) (out *todos.CreateChecklistResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/lists", true, in); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if !(status == http.StatusOK || status == http.StatusCreated) || !out.Success {
		return out, StatusError(status, out.Error)
	}

	return out, nil
}

// DetailChecklist returns as much information as possible about the specified checklist.
// This function checks the response for errors, but does not otherwise modify the
// output response. User authentication is required.
func (c *Client) DetailChecklist(id uint) (out *todos.DetailChecklistResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, fmt.Sprintf("/lists/%d", id), true, nil); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if status != http.StatusOK || !out.Success {
		return out, StatusError(status, out.Error)
	}

	return out, nil
}

// UpdateChecklist puts the checklist info to the specified id in order to update it.
// This function checks the response for errors, but does not otherwise modify the
// output response. User authentication is required.
func (c *Client) UpdateChecklist(id uint, list *todos.Checklist) (out *todos.UpdateChecklistResponse, err error) {
	if id == 0 || (list.ID > 0 && id != list.ID) {
		return nil, fmt.Errorf("cannot update with id %d and checklist id %d", id, list.ID)
	}

	// Ensure that the checklist ID is a zero value.
	list.ID = 0

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPut, fmt.Sprintf("/lists/%d", id), true, list); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, out); err != nil {
		return nil, err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) || !out.Success {
		return out, StatusError(status, out.Error)
	}
	return out, nil
}

// DeleteChecklist sends a delete request for the specified id. This function checks the
// response for errors, but does not otherwise modify the output response. User
// authentication is required.
func (c *Client) DeleteChecklist(id uint) (out *todos.DeleteChecklistResponse, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodDelete, fmt.Sprintf("/lists/%d", id), true, nil); err != nil {
		return nil, err
	}

	var status int
	if status, err = c.Do(req, &out); err != nil {
		return nil, err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) || !out.Success {
		return out, StatusError(status, out.Error)
	}
	return out, nil
}
