package types

import "time"

type UserRole string

const (
	Customer UserRole = "customer"
	Admin    UserRole = "admin"
)

type UserType struct {
	Id        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Password  string    `json:"password_hash"`
	Role      UserRole  `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UserRegisterType struct {
	Name     string   `json:"name"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Role     UserRole `json:"role"`
}

type UserLoginType struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
