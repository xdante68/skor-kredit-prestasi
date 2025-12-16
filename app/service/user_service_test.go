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
)

type MockUserRepo struct {
	mock.Mock
}

func (m *MockUserRepo) Create(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) FindByUsername(username string) (*model.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByUserID(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindByUserIDSimple(id uuid.UUID) (*model.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.User), args.Error(1)
}

func (m *MockUserRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.User, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.User), args.Get(1).(int64), args.Error(2)
}

func (m *MockUserRepo) Update(user *model.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserRepo) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserRepo) UpdateRole(userID uuid.UUID, roleID uuid.UUID) error {
	args := m.Called(userID, roleID)
	return args.Error(0)
}

func (m *MockUserRepo) AddBlacklistToken(token model.BlacklistedToken) error {
	args := m.Called(token)
	return args.Error(0)
}

func (m *MockUserRepo) ClearRefreshToken(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func (m *MockUserRepo) FindRoleByName(name string) (*model.Role, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Role), args.Error(1)
}

type MockStudentRepo struct {
	mock.Mock
}

func (m *MockStudentRepo) Create(student *model.Student) error {
	args := m.Called(student)
	return args.Error(0)
}

func (m *MockStudentRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.Student, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Student), args.Get(1).(int64), args.Error(2)
}

func (m *MockStudentRepo) FindByID(id uuid.UUID) (*model.Student, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepo) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepo) UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	args := m.Called(studentID, advisorID)
	return args.Error(0)
}

func (m *MockStudentRepo) ExistsByStudentID(studentID string) (bool, error) {
	args := m.Called(studentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStudentRepo) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockLecturerRepo struct {
	mock.Mock
}

func (m *MockLecturerRepo) Create(lecturer *model.Lecturer) error {
	args := m.Called(lecturer)
	return args.Error(0)
}

func (m *MockLecturerRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.Lecturer, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Lecturer), args.Get(1).(int64), args.Error(2)
}

func (m *MockLecturerRepo) FindByID(id uuid.UUID) (*model.Lecturer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

func (m *MockLecturerRepo) GetAdvisees(advisorID uuid.UUID) ([]model.Student, error) {
	args := m.Called(advisorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Student), args.Error(1)
}

func (m *MockLecturerRepo) ExistsByLecturerID(lecturerID string) (bool, error) {
	args := m.Called(lecturerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLecturerRepo) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func setupUserTestApp() *fiber.App {
	return fiber.New()
}

func TestGetAllUsers_Success(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	roleID := uuid.New()
	mockUsers := []model.User{
		{ID: uuid.New(), Username: "user1", Email: "user1@test.com", FullName: "User One", RoleID: &roleID, Role: model.Role{Name: "admin"}},
		{ID: uuid.New(), Username: "user2", Email: "user2@test.com", FullName: "User Two", RoleID: &roleID, Role: model.Role{Name: "mahasiswa"}},
	}

	mockUserRepo.On("FindAll", 1, 10, "", "created_at", "desc").Return(mockUsers, int64(2), nil)

	app.Get("/users", userService.GetAllUsers)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	assert.True(t, result["success"].(bool))
	mockUserRepo.AssertExpectations(t)
}

func TestGetAllUsers_Error(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	mockUserRepo.On("FindAll", 1, 10, "", "created_at", "desc").Return(nil, int64(0), errors.New("db error"))

	app.Get("/users", userService.GetAllUsers)

	req := httptest.NewRequest("GET", "/users", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
	mockUserRepo.AssertExpectations(t)
}

func TestGetUser_Success(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	userID := uuid.New()
	roleID := uuid.New()
	mockUser := &model.User{
		ID:       userID,
		Username: "testuser",
		Email:    "test@test.com",
		FullName: "Test User",
		RoleID:   &roleID,
		Role:     model.Role{Name: "admin"},
	}

	mockUserRepo.On("FindByUserIDSimple", userID).Return(mockUser, nil)

	app.Get("/users/:id", userService.GetUser)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockUserRepo.AssertExpectations(t)
}

func TestGetUser_NotFound(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	userID := uuid.New()
	mockUserRepo.On("FindByUserIDSimple", userID).Return(nil, errors.New("not found"))

	app.Get("/users/:id", userService.GetUser)

	req := httptest.NewRequest("GET", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
	mockUserRepo.AssertExpectations(t)
}

func TestGetUser_InvalidUUID(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	app.Get("/users/:id", userService.GetUser)

	req := httptest.NewRequest("GET", "/users/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestDeleteUser_Success(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	userID := uuid.New()
	mockUserRepo.On("Delete", userID).Return(nil)
	mockStudentRepo.On("DeleteByUserID", userID).Return(nil)
	mockLecturerRepo.On("DeleteByUserID", userID).Return(nil)

	app.Delete("/users/:id", userService.DeleteUser)

	req := httptest.NewRequest("DELETE", "/users/"+userID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockUserRepo.AssertExpectations(t)
}

func TestDeleteUser_Error(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	userID := uuid.New()
	mockUserRepo.On("Delete", userID).Return(errors.New("db error"))

	app.Delete("/users/:id", userService.DeleteUser)

	req := httptest.NewRequest("DELETE", "/users/"+userID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
	mockUserRepo.AssertExpectations(t)
}

func TestDeleteUser_InvalidUUID(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	app.Delete("/users/:id", userService.DeleteUser)

	req := httptest.NewRequest("DELETE", "/users/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestCreateUser_InvalidInput(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	app.Post("/users", userService.CreateUser)

	body := []byte(`{"invalid": json}`)

	req := httptest.NewRequest("POST", "/users", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestUpdateUser_InvalidUUID(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	app.Put("/users/:id", userService.UpdateUser)

	req := httptest.NewRequest("PUT", "/users/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestUpdateUser_NotFound(t *testing.T) {
	app := setupUserTestApp()
	mockUserRepo := new(MockUserRepo)
	mockStudentRepo := new(MockStudentRepo)
	mockLecturerRepo := new(MockLecturerRepo)
	userService := NewUserService(mockUserRepo, mockStudentRepo, mockLecturerRepo)

	userID := uuid.New()
	mockUserRepo.On("FindByUserID", userID).Return(nil, errors.New("not found"))

	app.Put("/users/:id", userService.UpdateUser)

	updateReq := map[string]interface{}{"full_name": "Updated Name"}
	body, _ := json.Marshal(updateReq)

	req := httptest.NewRequest("PUT", "/users/"+userID.String(), bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
	mockUserRepo.AssertExpectations(t)
}
