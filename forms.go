package todos

import "time"

// POST to /register to create a new user
type registerUserForm struct {
	Username string `json:"username" xml:"username" binding:"required"`
	Email    string `json:"email" xml:"email" binding:"required"`
	Password string `json:"password" xml:"password" binding:"required"`
	IsAdmin  bool   `json:"is_admin" xml:"is_admin"`
}

// POST to /login to authenticate the user
type loginUserForm struct {
	Username string `json:"username" xml:"username" binding:"required"`
	Password string `json:"password" xml:"password" binding:"required"`
	NoCookie bool   `json:"no_cookie" xml:"no_cookie"`
}

// POST to /logout to log the user out and revoke credentials
type logoutUserForm struct {
	RevokeAll bool `json:"revoke_all" xml:"revoke_all"`
}

// POST to /refresh to reauthorize the user with the refresh token
type refreshTokenForm struct {
	RefreshToken string `json:"refresh_token" xml:"refresh_token" binding:"required"`
	NoCookie     bool   `json:"no_cookie" xml:"no_cookie"`
}

// PUT to /todos/:id to update a todo
type updateTodo struct {
	Title     string     `json:"title" xml:"title"`
	Details   string     `json:"details" xml:"details"`
	Completed bool       `json:"completed" xml:"completed"`
	Archived  bool       `json:"archived" xml:"archived"`
	ListID    *uint      `json:"list" xml:"list"`
	Deadline  *time.Time `json:"deadline" xml:"deadline"`
}
