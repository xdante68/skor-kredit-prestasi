package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fiber/skp/app/model"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type MockUserRepository struct {
	mock.Mock
}

func (m *MockUserRepository) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUserID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindByUserIDSimple(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepository) FindAll(page, limit int, search, sortBy, order string) ([]model.User, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepository) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepository) UpdateRole(userID uuid.UUID, roleID uuid.UUID) error {
	args := m.Called(userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepository) AddBlacklistToken(token model.BlacklistedToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepository) ClearRefreshToken(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepository) FindRoleByName(name string) (*model.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

func setupTestApp() *fiber.App {
	return fiber.New()
}

func TestLogin_Success(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	roleID := uuid.New()

	mockUser := &model.User{
		ID:           uuid.New(),
		Username:     "testuser",
		Email:        "test@example.com",
		PasswordHash: string(hashedPassword),
		FullName:     "Test User",
		RoleID:       &roleID,
		IsActive:     true,
		Role: model.Role{
			Name: "mahasiswa",
		},
	}

	mockRepo.On("FindByUsername", "testuser").Return(mockUser, nil)
	mockRepo.On("Update", mock.AnythingOfType("*model.User")).Return(nil)

	app.Post("/auth/login", authService.Login)

	loginReq := model.LoginRequest{
		Username: "testuser",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	assert.True(t, result["success"].(bool))
	assert.NotNil(t, result["data"])

	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("correctpassword"), bcrypt.DefaultCost)
	roleID := uuid.New()

	mockUser := &model.User{
		ID:           uuid.New(),
		Username:     "testuser",
		PasswordHash: string(hashedPassword),
		IsActive:     true,
		RoleID:       &roleID,
	}

	mockRepo.On("FindByUsername", "testuser").Return(mockUser, nil)

	app.Post("/auth/login", authService.Login)

	loginReq := model.LoginRequest{
		Username: "testuser",
		Password: "wrongpassword",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 401, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	mockRepo.On("FindByUsername", "nonexistent").Return(nil, errors.New("user not found"))

	app.Post("/auth/login", authService.Login)

	loginReq := model.LoginRequest{
		Username: "nonexistent",
		Password: "password123",
	}
	body, _ := json.Marshal(loginReq)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 401, resp.StatusCode)
	mockRepo.AssertExpectations(t)
}

func TestLogin_InvalidInput(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	app.Post("/auth/login", authService.Login)

	body := []byte(`{"invalid": json}`)

	req := httptest.NewRequest("POST", "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestLogout_NoToken(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	app.Post("/auth/logout", authService.Logout)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestLogout_InvalidTokenFormat(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	app.Post("/auth/logout", authService.Logout)

	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "Bearer")
	resp, _ := app.Test(req)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestLogout_ShortBearer(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	app.Post("/auth/logout", authService.Logout)
	req := httptest.NewRequest("POST", "/auth/logout", nil)
	req.Header.Set("Authorization", "abc")
	resp, _ := app.Test(req)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestRefresh_InvalidInput(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	app.Post("/auth/refresh", authService.Refresh)

	body := []byte(`{"invalid": json}`)

	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestRefresh_EmptyToken(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	app.Post("/auth/refresh", authService.Refresh)

	refreshReq := model.RefreshTokenRequest{
		RefreshToken: "",
	}
	body, _ := json.Marshal(refreshReq)

	req := httptest.NewRequest("POST", "/auth/refresh", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 401, resp.StatusCode)
}

func TestProfile_WithValidContext(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	userID := uuid.New()

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("username", "testuser")
		c.Locals("email", "test@example.com")
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Get("/auth/profile", authService.Profile)

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	assert.True(t, result["success"].(bool))
	data := result["data"].(map[string]interface{})
	assert.Equal(t, "testuser", data["username"])
	assert.Equal(t, "mahasiswa", data["role"])
}

func TestProfile_StringUserID(t *testing.T) {
	app := setupTestApp()
	mockRepo := new(MockUserRepository)
	authService := NewAuthService(mockRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", "550e8400-e29b-41d4-a716-446655440000")
		c.Locals("username", "testuser")
		c.Locals("email", "test@example.com")
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/auth/profile", authService.Profile)

	req := httptest.NewRequest("GET", "/auth/profile", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 200, resp.StatusCode)
}
