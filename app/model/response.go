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

type PaginationData[T any] struct {
	Items []T      `json:"items"`
	Meta  MetaInfo `json:"meta"`
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

type ProfileResponse struct {
	Success bool        `json:"success"`
	Data    ProfileData `json:"data"`
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

type SwaggerLoginResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message,omitempty"`
	Data    LoginResponse `json:"data"`
}

type SwaggerRefreshResponse struct {
	Success bool                 `json:"success"`
	Message string               `json:"message,omitempty"`
	Data    RefreshTokenResponse `json:"data"`
}

type SwaggerProfileResponse struct {
	Success bool        `json:"success"`
	Data    ProfileData `json:"data"`
}

type SwaggerUserListResponse struct {
	Success bool           `json:"success"`
	Data    []UserResponse `json:"data"`
	Meta    MetaInfo       `json:"meta,omitempty"`
}

type SwaggerUserResponse struct {
	Success bool         `json:"success"`
	Data    UserResponse `json:"data"`
}

type SwaggerStudentListResponse struct {
	Success bool                  `json:"success"`
	Data    []StudentListResponse `json:"data"`
	Meta    MetaInfo              `json:"meta,omitempty"`
}

type SwaggerStudentDetailResponse struct {
	Success bool                  `json:"success"`
	Data    StudentDetailResponse `json:"data"`
}

type SwaggerLecturerListResponse struct {
	Success bool                   `json:"success"`
	Data    []LecturerListResponse `json:"data"`
	Meta    MetaInfo               `json:"meta,omitempty"`
}

type SwaggerAdviseesResponse struct {
	Success bool                  `json:"success"`
	Data    []StudentListResponse `json:"data"`
}

type SwaggerAchievementListResponse struct {
	Success bool                  `json:"success"`
	Data    []AchievementResponse `json:"data"`
	Meta    MetaInfo              `json:"meta,omitempty"`
}

type SwaggerAchievementResponse struct {
	Success bool                `json:"success"`
	Data    AchievementResponse `json:"data"`
}

type SwaggerAchievementHistoryResponse struct {
	Success bool          `json:"success"`
	Data    []interface{} `json:"data"`
}

type SwaggerAttachmentResponse struct {
	Success bool               `json:"success"`
	Data    AttachmentResponse `json:"data"`
}

type SwaggerStatsResponse struct {
	Success bool          `json:"success"`
	Data    StatsResponse `json:"data"`
}

type SwaggerStudentStatsResponse struct {
	Success bool                 `json:"success"`
	Data    StudentStatsResponse `json:"data"`
}

type AssignAdvisorRequest struct {
	LecturerID string `json:"lecturer_id"`
}

type AttachmentResponse struct {
	Filename   string `json:"filename"`
	URL        string `json:"url"`
	MimeType   string `json:"mime_type"`
	Size       int64  `json:"size"`
	UploadedAt string `json:"uploaded_at"`
}
