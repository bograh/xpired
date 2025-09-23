package api

type UserRequest struct {
	Email       string  `json:"email"`
	Password    string  `json:"password"`
	Name        string  `json:"name"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          string  `json:"id"`
	Email       string  `json:"email"`
	PhoneNumber *string `json:"phoneNumber,omitempty"`
	Name        string  `json:"name"`
}
