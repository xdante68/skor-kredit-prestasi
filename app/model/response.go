package model

type MetaInfo struct {
	Page   int    `json:"page"`
	Limit  int    `json:"limit"`
	Total  int64  `json:"total"`
	Pages  int    `json:"pages"`
	SortBy string `json:"sortBy"`
	Order  string `json:"order"`
	Search string `json:"search"`
}

type SuccessMessageResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Error   string `json:"error,omitempty"`
}

type SuccessResponse[T any] struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    T      `json:"data,omitempty"`
}

type LoginUser struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
}

type LoginResponse struct {
	User         LoginUser `json:"user"`
	Token        string    `json:"token"`
	RefreshToken string    `json:"refreshToken"`
}

type LoginSuccessResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Data    LoginResponse `json:"data"`
}

type ProfileData struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type ProfileResponse struct {
	Success bool        `json:"success"`
	Data    ProfileData `json:"data"`
}
