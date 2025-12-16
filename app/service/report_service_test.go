package service

import (
	"errors"
	"fiber/skp/app/model"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type MockReportRepository struct {
	mock.Mock
}

func (m *MockReportRepository) GetMongoIDs(role string, userID uuid.UUID) ([]primitive.ObjectID, error) {
	args := m.Called(role, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]primitive.ObjectID), args.Error(1)
}

func (m *MockReportRepository) GetStatsByPeriod(role string, userID uuid.UUID) ([]model.StatItem, error) {
	args := m.Called(role, userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StatItem), args.Error(1)
}

func (m *MockReportRepository) GetTopStudents() ([]model.TopStudent, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.TopStudent), args.Error(1)
}

func (m *MockReportRepository) GetStudentProfile(studentID uuid.UUID) (*model.TopStudent, error) {
	args := m.Called(studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.TopStudent), args.Error(1)
}

func (m *MockReportRepository) GetMongoIDsByStudentID(studentID uuid.UUID) ([]primitive.ObjectID, error) {
	args := m.Called(studentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]primitive.ObjectID), args.Error(1)
}

func (m *MockReportRepository) GetStatsByType(ids []primitive.ObjectID) ([]model.StatItem, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StatItem), args.Error(1)
}

func (m *MockReportRepository) GetStatsByLevel(ids []primitive.ObjectID) ([]model.StatItem, error) {
	args := m.Called(ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.StatItem), args.Error(1)
}

// MockStudentRepoForReport
type MockStudentRepoForReport struct {
	mock.Mock
}

func (m *MockStudentRepoForReport) Create(student *model.Student) error {
	args := m.Called(student)
	return args.Error(0)
}

func (m *MockStudentRepoForReport) FindAll(page, limit int, search, sortBy, order string) ([]model.Student, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Student), args.Get(1).(int64), args.Error(2)
}

func (m *MockStudentRepoForReport) FindByID(id uuid.UUID) (*model.Student, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepoForReport) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepoForReport) UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	args := m.Called(studentID, advisorID)
	return args.Error(0)
}

func (m *MockStudentRepoForReport) ExistsByStudentID(studentID string) (bool, error) {
	args := m.Called(studentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStudentRepoForReport) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}


func TestGetStatistics_AdminSuccess(t *testing.T) {
	app := fiber.New()
	mockReportRepo := new(MockReportRepository)
	mockStudentRepo := new(MockStudentRepoForReport)

	reportService := NewReportService(mockReportRepo, mockStudentRepo)

	userID := uuid.New()
	mockMongoIDs := []primitive.ObjectID{primitive.NewObjectID(), primitive.NewObjectID()}
	mockStatsByPeriod := []model.StatItem{{Label: "2024", Count: 5}}
	mockTopStudents := []model.TopStudent{{StudentID: "MHS001", StudentName: "Test", TotalAchievements: 3}}

	mockReportRepo.On("GetMongoIDs", "admin", userID).Return(mockMongoIDs, nil)
	mockReportRepo.On("GetStatsByPeriod", "admin", userID).Return(mockStatsByPeriod, nil)
	mockReportRepo.On("GetTopStudents").Return(mockTopStudents, nil)
	mockReportRepo.On("GetStatsByType", mockMongoIDs).Return([]model.StatItem{{Label: "competition", Count: 2}}, nil)
	mockReportRepo.On("GetStatsByLevel", mockMongoIDs).Return([]model.StatItem{{Label: "national", Count: 1}}, nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/reports/statistics", reportService.GetStatistics)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(respBody), "total_achievements")
	mockReportRepo.AssertExpectations(t)
}

func TestGetStatistics_MahasiswaSuccess(t *testing.T) {
	app := fiber.New()
	mockReportRepo := new(MockReportRepository)
	mockStudentRepo := new(MockStudentRepoForReport)

	reportService := NewReportService(mockReportRepo, mockStudentRepo)

	userID := uuid.New()
	mockMongoIDs := []primitive.ObjectID{primitive.NewObjectID()}
	mockStatsByPeriod := []model.StatItem{{Label: "2024", Count: 1}}

	mockReportRepo.On("GetMongoIDs", "mahasiswa", userID).Return(mockMongoIDs, nil)
	mockReportRepo.On("GetStatsByPeriod", "mahasiswa", userID).Return(mockStatsByPeriod, nil)
	mockReportRepo.On("GetStatsByType", mockMongoIDs).Return([]model.StatItem{}, nil)
	mockReportRepo.On("GetStatsByLevel", mockMongoIDs).Return([]model.StatItem{}, nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Get("/reports/statistics", reportService.GetStatistics)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockReportRepo.AssertExpectations(t)
}

func TestGetStatistics_Error(t *testing.T) {
	app := fiber.New()
	mockReportRepo := new(MockReportRepository)
	mockStudentRepo := new(MockStudentRepoForReport)

	reportService := NewReportService(mockReportRepo, mockStudentRepo)

	userID := uuid.New()

	mockReportRepo.On("GetMongoIDs", "admin", userID).Return(nil, errors.New("db error"))

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/reports/statistics", reportService.GetStatistics)

	req := httptest.NewRequest("GET", "/reports/statistics", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
	mockReportRepo.AssertExpectations(t)
}

func TestGetStudentStats_InvalidID(t *testing.T) {
	app := fiber.New()
	mockReportRepo := new(MockReportRepository)
	mockStudentRepo := new(MockStudentRepoForReport)

	reportService := NewReportService(mockReportRepo, mockStudentRepo)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/reports/student/:id", reportService.GetStudentStats)

	req := httptest.NewRequest("GET", "/reports/student/invalid-id", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetStudentStats_StudentNotFound(t *testing.T) {
	app := fiber.New()
	mockReportRepo := new(MockReportRepository)
	mockStudentRepo := new(MockStudentRepoForReport)

	reportService := NewReportService(mockReportRepo, mockStudentRepo)

	studentID := uuid.New()
	mockReportRepo.On("GetStudentProfile", studentID).Return(nil, errors.New("not found"))

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", uuid.New())
		c.Locals("role", "admin")
		return c.Next()
	})

	app.Get("/reports/student/:id", reportService.GetStudentStats)

	req := httptest.NewRequest("GET", "/reports/student/"+studentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
	mockReportRepo.AssertExpectations(t)
}

func TestGetStudentStats_Forbidden(t *testing.T) {
	app := fiber.New()
	mockReportRepo := new(MockReportRepository)
	mockStudentRepo := new(MockStudentRepoForReport)

	reportService := NewReportService(mockReportRepo, mockStudentRepo)

	studentID := uuid.New()
	userID := uuid.New()
	otherStudentID := uuid.New()

	mockStudentForUser := &model.Student{
		ID:     otherStudentID,
		UserID: userID,
	}

	mockStudentRepo.On("FindByUserID", userID).Return(mockStudentForUser, nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Get("/reports/student/:id", reportService.GetStudentStats)

	req := httptest.NewRequest("GET", "/reports/student/"+studentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 403, resp.StatusCode)
	mockStudentRepo.AssertExpectations(t)
}

func TestGetStudentStats_Success(t *testing.T) {
	app := fiber.New()
	mockReportRepo := new(MockReportRepository)
	mockStudentRepo := new(MockStudentRepoForReport)

	reportService := NewReportService(mockReportRepo, mockStudentRepo)

	studentID := uuid.New()
	userID := uuid.New()

	mockStudentForUser := &model.Student{
		ID:     studentID,
		UserID: userID,
	}

	mockProfile := &model.TopStudent{
		StudentID:         "MHS001",
		StudentName:       "Test Student",
		TotalAchievements: 5,
		TotalPoints:       100,
	}

	mockMongoIDs := []primitive.ObjectID{primitive.NewObjectID()}

	mockStudentRepo.On("FindByUserID", userID).Return(mockStudentForUser, nil)
	mockReportRepo.On("GetStudentProfile", studentID).Return(mockProfile, nil)
	mockReportRepo.On("GetMongoIDsByStudentID", studentID).Return(mockMongoIDs, nil)
	mockReportRepo.On("GetStatsByType", mockMongoIDs).Return([]model.StatItem{}, nil)
	mockReportRepo.On("GetStatsByLevel", mockMongoIDs).Return([]model.StatItem{}, nil)

	app.Use(func(c *fiber.Ctx) error {
		c.Locals("user_id", userID)
		c.Locals("role", "mahasiswa")
		return c.Next()
	})

	app.Get("/reports/student/:id", reportService.GetStudentStats)

	req := httptest.NewRequest("GET", "/reports/student/"+studentID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockStudentRepo.AssertExpectations(t)
	mockReportRepo.AssertExpectations(t)
}
