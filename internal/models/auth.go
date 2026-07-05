package models

type SignupRequest struct {
	Name            string
	Email           string `binding:"required,email"`
	Password        string `binding:"required,min=8"`
	ConfirmPassword string `binding:"required,min=8"`
}

type SigninRequest struct {
	Email    string `binding:"required,email"`
	Password string `binding:"required,min=8"`
}

type User struct {
	ID        string `db:"id"`
	Name      string `db:"name"`
	Email     string `db:"email"`
	Password  string `json:"-" db:"password"`
	CreatedAt int64  `db:"created_at"`
	UpdatedAt int64  `db:"updated_at"`
}

type SignupResponse struct {
	User User
}

type UpdateProfileRequest struct {
	Name     *string `binding:"omitempty,min=2"`
	Email    *string `binding:"omitempty,email"`
	Password *string `binding:"omitempty,min=8"`
}
