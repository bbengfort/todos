package todos

type registerUserForm struct {
	Username string `json:"username" xml:"username" binding:"required"`
	Email string `json:"email" xml:"email" binding:"required"`
	Password string `json:"password" xml:"password" binding:"required"`
}

type loginUserForm struct {
	Username string `json:"username" xml:"username" binding:"required"`
	Password string `json:"password" xml:"password" binding:"required"`
}
