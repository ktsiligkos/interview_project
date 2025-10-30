package domain

// This is the user with the passowrd and its hashed password
type User struct {
	ID       string
	Name     string
	Email    string
	Password string
}

// This is the user login request
type UserLoginRequest struct {
	Email    string `json:"email" binding:"required"`
	Password string `json:"password" binding:"required"`
}
