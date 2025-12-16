package service

import (
	"bytes"
	"encoding/json"
	"errors"
	"fiber/skp/app/model"
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockAchievementRepository struct {
	mock.Mock
}

func (m *MockAchievementRepository) Create(studentID uuid.UUID, req model.CreateAchievementRequest) (*model.AchievementResponse, error) {
	args := m.Called(studentID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementResponse), args.Error(1)
}

func (m *MockAchievementRepository) FindByAchievementID(id uuid.UUID) (*model.AchievementResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementResponse), args.Error(1)
}

func (m *MockAchievementRepository) FindAll(role string, userID uuid.UUID, page, limit int, search, sortBy, order string) ([]model.AchievementResponse, int64, error) {
	args := m.Called(role, userID, page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.AchievementResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockAchievementRepository) Update(id uuid.UUID, req model.UpdateAchievementRequest) (*model.AchievementResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementResponse), args.Error(1)
}

func (m *MockAchievementRepository) UpdateStatus(id uuid.UUID, status string, verifierID *uuid.UUID, note string, points int) error {
	args := m.Called(id, status, verifierID, note, points)
	return args.Error(0)
}

func (m *MockAchievementRepository) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAchievementRepository) AddAttachment(id uuid.UUID, attachment model.Attachment) error {
	args := m.Called(id, attachment)
	return args.Error(0)
}

func (m *MockAchievementRepository) GetOwnerID(id uuid.UUID) (uuid.UUID, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return uuid.Nil, args.Error(1)
	}
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAchievementRepository) IsAdvisor(advisorID uuid.UUID, achievementID uuid.UUID) (bool, error) {
	args := m.Called(advisorID, achievementID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAchievementRepository) GetStatus(id uuid.UUID) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockAchievementRepository) GetHistory(id uuid.UUID) (*model.AchievementHistoryResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementHistoryResponse), args.Error(1)
}

type MockStudentRepoForAchievement struct {
	mock.Mock
}

func (m *MockStudentRepoForAchievement) Create(student *model.Student) error {
	args := m.Called(student)
	return args.Error(0)
}

func (m *MockStudentRepoForAchievement) FindAll(page, limit int, search, sortBy, order string) ([]model.Student, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Student), args.Get(1).(int64), args.Error(2)
}

func (m *MockStudentRepoForAchievement) FindByID(id uuid.UUID) (*model.Student, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepoForAchievement) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepoForAchievement) UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	args := m.Called(studentID, advisorID)
	return args.Error(0)
}

func (m *MockStudentRepoForAchievement) ExistsByStudentID(studentID string) (bool, error) {
	args := m.Called(studentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStudentRepoForAchievement) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockLecturerRepoForAchievement struct {
	mock.Mock
}

func (m *MockLecturerRepoForAchievement) Create(lecturer *model.Lecturer) error {
	args := m.Called(lecturer)
	return args.Error(0)
}

func (m *MockLecturerRepoForAchievement) FindAll(page, limit int, search, sortBy, order string) ([]model.Lecturer, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Lecturer), args.Get(1).(int64), args.Error(2)
}

func (m *MockLecturerRepoForAchievement) FindByID(id uuid.UUID) (*model.Lecturer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

func (m *MockLecturerRepoForAchievement) GetAdvisees(advisorID uuid.UUID) ([]model.Student, error) {
	args := m.Called(advisorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Student), args.Error(1)
}

func (m *MockLecturerRepoForAchievement) ExistsByLecturerID(lecturerID string) (bool, error) {
	args := m.Called(lecturerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLecturerRepoForAchievement) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

func TestListAchievements_Success(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	userID := uuid.New()

	mockAchievements := []model.AchievementResponse{
		{
			ID:              uuid.New(),
			Title:           "Test Achievement",
			AchievementType: "competition",
			Status:          "verified",
		},
	}

	mockAchievementRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockAchievements, int64(1), nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/achievements", achievementService.List)

	req := httptest.NewRequest("GET", "/achievements?page=1&limit=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	assert.True(t, result["success"].(bool))
	mockAchievementRepo.AssertExpectations(t)
}

func TestListAchievements_Error(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	userID := uuid.New()

	mockAchievementRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/achievements", achievementService.List)

	req := httptest.NewRequest("GET", "/achievements?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
	mockAchievementRepo.AssertExpectations(t)
}

func TestGetAchievement_Success(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	achievementID := uuid.New()
	userID := uuid.New()

	mockAchievement := &model.AchievementResponse{
		ID:              achievementID,
		Title:           "Test Achievement",
		AchievementType: "competition",
		Status:          "verified",
		CreatedAt:       time.Now(),
	}

	mockAchievementRepo.On("FindByAchievementID", achievementID).Return(mockAchievement, nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/achievements/:id", achievementService.Get)

	req := httptest.NewRequest("GET", "/achievements/"+achievementID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockAchievementRepo.AssertExpectations(t)
}

func TestGetAchievement_InvalidID(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/achievements/:id", achievementService.Get)

	req := httptest.NewRequest("GET", "/achievements/invalid-id", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetAchievement_NotFound(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	achievementID := uuid.New()

	mockAchievementRepo.On("FindByAchievementID", achievementID).Return(nil, errors.New("not found"))

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/achievements/:id", achievementService.Get)

	req := httptest.NewRequest("GET", "/achievements/"+achievementID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
	mockAchievementRepo.AssertExpectations(t)
}

func TestCreateAchievement_InvalidInput(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Post("/achievements", achievementService.Create)

	body := []byte(`{invalid json}`)
	req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestCreateAchievement_InvalidType(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Post("/achievements", achievementService.Create)

	createReq := model.CreateAchievementRequest{
		Title:           "Test",
		AchievementType: "invalid_type",
	}
	body, _ := json.Marshal(createReq)

	req := httptest.NewRequest("POST", "/achievements", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestDeleteAchievement_InvalidID(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Delete("/achievements/:id", achievementService.Delete)

	req := httptest.NewRequest("DELETE", "/achievements/invalid-id", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestSubmitAchievement_InvalidID(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Post("/achievements/:id/submit", achievementService.Submit)

	req := httptest.NewRequest("POST", "/achievements/invalid-id/submit", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestVerifyAchievement_InvalidID(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "dosen_wali")
		return c.Next()
	})

	app.Post("/achievements/:id/verify", achievementService.Verify)

	req := httptest.NewRequest("POST", "/achievements/invalid-id/verify", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestRejectAchievement_InvalidID(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "dosen_wali")
		return c.Next()
	})

	app.Post("/achievements/:id/reject", achievementService.Reject)

	req := httptest.NewRequest("POST", "/achievements/invalid-id/reject", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetHistory_Success(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	achievementID := uuid.New()

	mockHistory := &model.AchievementHistoryResponse{
		ID:     achievementID,
		Title:  "Test Achievement",
		Status: "verified",
	}

	mockAchievementRepo.On("GetHistory", achievementID).Return(mockHistory, nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/achievements/:id/history", achievementService.GetHistory)

	req := httptest.NewRequest("GET", "/achievements/"+achievementID.String()+"/history", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockAchievementRepo.AssertExpectations(t)
}

func TestGetHistory_InvalidID(t *testing.T) {
	app := fiber.New()
	mockAchievementRepo := new(MockAchievementRepository)
	mockStudentRepo := new(MockStudentRepoForAchievement)
	mockLecturerRepo := new(MockLecturerRepoForAchievement)

	achievementService := NewAchievementService(mockAchievementRepo, mockStudentRepo, mockLecturerRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/achievements/:id/history", achievementService.GetHistory)

	req := httptest.NewRequest("GET", "/achievements/invalid-id/history", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}
