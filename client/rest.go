package client

import (
	"errors"
	"fmt"
	"net/http"
	"time"
)

// TODO: replace this entire file with something better!

func (c *Client) FindTodos() (data map[string]interface{}, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, "/todos", true, nil); err != nil {
		return nil, err
	}

	var status int
	if data, status, err = c.Do(req); err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		if errmsg, ok := data["error"]; ok {
			return nil, errors.New(errmsg.(string))
		}
		return nil, fmt.Errorf("could not fetch todos: %s", http.StatusText(status))
	}

	return data, nil
}

func (c *Client) CreateTodo(title, details string, listID uint, deadline time.Time) (err error) {
	data := make(map[string]interface{})
	data["title"] = title
	data["details"] = details

	if listID > 0 {
		data["list"] = listID
	}

	if !deadline.IsZero() {
		data["deadline"] = deadline
	}

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/todos", true, data); err != nil {
		return err
	}

	var status int
	var info map[string]interface{}
	if info, status, err = c.Do(req); err != nil {
		return err
	}

	if !(status == http.StatusOK || status == http.StatusCreated) {
		if errmsg, ok := info["error"]; ok {
			return errors.New(errmsg.(string))
		}
		return fmt.Errorf("could not create todo: %s", http.StatusText(status))
	}

	return nil
}

func (c *Client) DetailTodo(id uint) (data map[string]interface{}, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, fmt.Sprintf("/todos/%d", id), true, nil); err != nil {
		return nil, err
	}

	var status int
	if data, status, err = c.Do(req); err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		if errmsg, ok := data["error"]; ok {
			return nil, errors.New(errmsg.(string))
		}
		return nil, fmt.Errorf("could not fetch todo %d: %s", id, http.StatusText(status))
	}

	return data, nil
}

func (c *Client) UpdateTodo(id uint, title, details string, listID uint, completed, archived bool, deadline time.Time) (err error) {
	data := make(map[string]interface{})
	if title != "" {
		data["title"] = title
	}
	if details != "" {
		data["details"] = details
	}
	if listID > 0 {
		data["list"] = listID
	}
	if completed {
		data["completed"] = true
	}
	if archived {
		data["archived"] = true
	}
	if !deadline.IsZero() {
		data["deadline"] = deadline
	}

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPut, fmt.Sprintf("/todos/%d", id), true, data); err != nil {
		return err
	}

	var status int
	var info map[string]interface{}
	if info, status, err = c.Do(req); err != nil {
		return err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) {
		if errmsg, ok := info["error"]; ok {
			return errors.New(errmsg.(string))
		}
		return fmt.Errorf("could not update todo %d: %s", id, http.StatusText(status))
	}
	return nil
}

func (c *Client) DeleteTodo(id uint) (err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodDelete, fmt.Sprintf("/todos/%d", id), true, nil); err != nil {
		return err
	}

	var status int
	var data map[string]interface{}
	if data, status, err = c.Do(req); err != nil {
		return err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) {
		if errmsg, ok := data["error"]; ok {
			return errors.New(errmsg.(string))
		}
		return fmt.Errorf("could not delete todo %d: %s", id, http.StatusText(status))
	}
	return nil
}

func (c *Client) FindLists() (data map[string]interface{}, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, "/lists", true, nil); err != nil {
		return nil, err
	}

	var status int
	if data, status, err = c.Do(req); err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		if errmsg, ok := data["error"]; ok {
			return nil, errors.New(errmsg.(string))
		}
		return nil, fmt.Errorf("could not fetch lists: %s", http.StatusText(status))
	}

	return data, nil
}

func (c *Client) CreateList(title, details string, deadline time.Time) (err error) {
	data := make(map[string]interface{})
	data["title"] = title
	data["details"] = details

	if !deadline.IsZero() {
		data["deadline"] = deadline
	}

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPost, "/lists", true, data); err != nil {
		return err
	}

	var status int
	var info map[string]interface{}
	if info, status, err = c.Do(req); err != nil {
		return err
	}

	if !(status == http.StatusOK || status == http.StatusCreated) {
		if errmsg, ok := info["error"]; ok {
			return errors.New(errmsg.(string))
		}
		return fmt.Errorf("could not create list: %s", http.StatusText(status))
	}

	return nil
}

func (c *Client) DetailList(id uint) (data map[string]interface{}, err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodGet, fmt.Sprintf("/lists/%d", id), true, nil); err != nil {
		return nil, err
	}

	var status int
	if data, status, err = c.Do(req); err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		if errmsg, ok := data["error"]; ok {
			return nil, errors.New(errmsg.(string))
		}
		return nil, fmt.Errorf("could not fetch list %d: %s", id, http.StatusText(status))
	}

	return data, nil
}

func (c *Client) UpdateList(id uint, title, details string, deadline time.Time) (err error) {
	data := make(map[string]interface{})
	if title != "" {
		data["title"] = title
	}
	if details != "" {
		data["details"] = details
	}
	if !deadline.IsZero() {
		data["deadline"] = deadline
	}

	var req *http.Request
	if req, err = c.NewRequest(http.MethodPut, fmt.Sprintf("/lists/%d", id), true, data); err != nil {
		return err
	}

	var status int
	var info map[string]interface{}
	if info, status, err = c.Do(req); err != nil {
		return err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) {
		if errmsg, ok := info["error"]; ok {
			return errors.New(errmsg.(string))
		}
		return fmt.Errorf("could not update list %d: %s", id, http.StatusText(status))
	}
	return nil
}

func (c *Client) DeleteList(id uint) (err error) {
	var req *http.Request
	if req, err = c.NewRequest(http.MethodDelete, fmt.Sprintf("/lists/%d", id), true, nil); err != nil {
		return err
	}

	var status int
	var data map[string]interface{}
	if data, status, err = c.Do(req); err != nil {
		return err
	}

	if !(status == http.StatusOK || status == http.StatusNoContent) {
		if errmsg, ok := data["error"]; ok {
			return errors.New(errmsg.(string))
		}
		return fmt.Errorf("could not delete list %d: %s", id, http.StatusText(status))
	}
	return nil
}
