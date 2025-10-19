package models

type User struct {
	Id       int64  `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserId   string `gorm:"column:user_id" json:"user_id"`
	Name     string `gorm:"column:name" json:"name"`
	Email    string `gorm:"column:email" json:"email"`
	Password string `gorm:"column:password" json:"-"`
	Role     string `gorm:"column:role" json:"role"`
	Status   string `gorm:"column:status" json:"status"`
}

type UserRequest struct {
	Name  string `gorm:"column:name" json:"name"`
	Email string `gorm:"column:email" json:"email"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type RegisterResponse struct {
	Message string `json:"message"`
}

type ErrorResponse struct {
	Message string `json:"message"`
}
