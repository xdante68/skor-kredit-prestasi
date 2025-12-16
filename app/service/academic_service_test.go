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

type MockStudentRepository struct {
	mock.Mock
}

func (m *MockStudentRepository) Create(student *model.Student) error {
	args := m.Called(student)
	return args.Error(0)
}

func (m *MockStudentRepository) FindAll(page, limit int, search, sortBy, order string) ([]model.Student, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Student), args.Get(1).(int64), args.Error(2)
}

func (m *MockStudentRepository) FindByID(id uuid.UUID) (*model.Student, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepository) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	args := m.Called(userID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Student), args.Error(1)
}

func (m *MockStudentRepository) UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	args := m.Called(studentID, advisorID)
	return args.Error(0)
}

func (m *MockStudentRepository) ExistsByStudentID(studentID string) (bool, error) {
	args := m.Called(studentID)
	return args.Bool(0), args.Error(1)
}

func (m *MockStudentRepository) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockLecturerRepository struct {
	mock.Mock
}

func (m *MockLecturerRepository) Create(lecturer *model.Lecturer) error {
	args := m.Called(lecturer)
	return args.Error(0)
}

func (m *MockLecturerRepository) FindAll(page, limit int, search, sortBy, order string) ([]model.Lecturer, int64, error) {
	args := m.Called(page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.Lecturer), args.Get(1).(int64), args.Error(2)
}

func (m *MockLecturerRepository) FindByID(id uuid.UUID) (*model.Lecturer, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.Lecturer), args.Error(1)
}

func (m *MockLecturerRepository) GetAdvisees(advisorID uuid.UUID) ([]model.Student, error) {
	args := m.Called(advisorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]model.Student), args.Error(1)
}

func (m *MockLecturerRepository) ExistsByLecturerID(lecturerID string) (bool, error) {
	args := m.Called(lecturerID)
	return args.Bool(0), args.Error(1)
}

func (m *MockLecturerRepository) DeleteByUserID(userID uuid.UUID) error {
	args := m.Called(userID)
	return args.Error(0)
}

type MockAchievementRepoForAcademic struct {
	mock.Mock
}

func (m *MockAchievementRepoForAcademic) Create(studentID uuid.UUID, req model.CreateAchievementRequest) (*model.AchievementResponse, error) {
	args := m.Called(studentID, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementResponse), args.Error(1)
}

func (m *MockAchievementRepoForAcademic) FindByAchievementID(id uuid.UUID) (*model.AchievementResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementResponse), args.Error(1)
}

func (m *MockAchievementRepoForAcademic) FindAll(role string, userID uuid.UUID, page, limit int, search, sortBy, order string) ([]model.AchievementResponse, int64, error) {
	args := m.Called(role, userID, page, limit, search, sortBy, order)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]model.AchievementResponse), args.Get(1).(int64), args.Error(2)
}

func (m *MockAchievementRepoForAcademic) Update(id uuid.UUID, req model.UpdateAchievementRequest) (*model.AchievementResponse, error) {
	args := m.Called(id, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementResponse), args.Error(1)
}

func (m *MockAchievementRepoForAcademic) UpdateStatus(id uuid.UUID, status string, verifierID *uuid.UUID, note string, points int) error {
	args := m.Called(id, status, verifierID, note, points)
	return args.Error(0)
}

func (m *MockAchievementRepoForAcademic) Delete(id uuid.UUID) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockAchievementRepoForAcademic) AddAttachment(id uuid.UUID, attachment model.Attachment) error {
	args := m.Called(id, attachment)
	return args.Error(0)
}

func (m *MockAchievementRepoForAcademic) GetOwnerID(id uuid.UUID) (uuid.UUID, error) {
	args := m.Called(id)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockAchievementRepoForAcademic) IsAdvisor(advisorID uuid.UUID, achievementID uuid.UUID) (bool, error) {
	args := m.Called(advisorID, achievementID)
	return args.Bool(0), args.Error(1)
}

func (m *MockAchievementRepoForAcademic) GetStatus(id uuid.UUID) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockAchievementRepoForAcademic) GetHistory(id uuid.UUID) (*model.AchievementHistoryResponse, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*model.AchievementHistoryResponse), args.Error(1)
}

func TestGetAllStudents_Success(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	studentID := uuid.New()
	userID := uuid.New()

	mockStudents := []model.Student{
		{
			ID:           studentID,
			UserID:       userID,
			StudentID:    "123456",
			ProgramStudy: "Informatika",
			AcademicYear: "2024",
			User: model.User{
				FullName: "Test Student",
			},
		},
	}

	mockStudentRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockStudents, int64(1), nil)

	app.Get("/students", academicService.GetAllStudents)

	req := httptest.NewRequest("GET", "/students?page=1&limit=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)

	respBody, _ := io.ReadAll(resp.Body)
	var result map[string]interface{}
	json.Unmarshal(respBody, &result)

	assert.True(t, result["success"].(bool))
	mockStudentRepo.AssertExpectations(t)
}

func TestGetAllStudents_Error(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	mockStudentRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("database error"))

	app.Get("/students", academicService.GetAllStudents)

	req := httptest.NewRequest("GET", "/students?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
	mockStudentRepo.AssertExpectations(t)
}

func TestGetStudentDetail_Success(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	studentID := uuid.New()
	userID := uuid.New()

	mockStudent := &model.Student{
		ID:           studentID,
		UserID:       userID,
		StudentID:    "123456",
		ProgramStudy: "Informatika",
		AcademicYear: "2024",
		CreatedAt:    time.Now(),
		User: model.User{
			FullName: "Test Student",
			Email:    "test@test.com",
		},
	}

	mockStudentRepo.On("FindByID", studentID).Return(mockStudent, nil)

	app.Get("/students/:id", academicService.GetStudentDetail)

	req := httptest.NewRequest("GET", "/students/"+studentID.String(), nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockStudentRepo.AssertExpectations(t)
}

func TestGetStudentDetail_InvalidID(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	app.Get("/students/:id", academicService.GetStudentDetail)

	req := httptest.NewRequest("GET", "/students/invalid-uuid", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetStudentDetail_NotFound(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	studentID := uuid.New()
	mockStudentRepo.On("FindByID", studentID).Return(nil, errors.New("not found"))

	app.Get("/students/:id", academicService.GetStudentDetail)

	req := httptest.NewRequest("GET", "/students/"+studentID.String(), nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
	mockStudentRepo.AssertExpectations(t)
}

func TestAssignAdvisor_InvalidStudentID(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	app.Put("/students/:id/advisor", academicService.AssignAdvisor)

	body := []byte(`{"lecturer_id": "some-id"}`)
	req := httptest.NewRequest("PUT", "/students/invalid-uuid/advisor", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetAllLecturers_Success(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	lecturerID := uuid.New()
	userID := uuid.New()

	mockLecturers := []model.Lecturer{
		{
			ID:         lecturerID,
			UserID:     userID,
			LecturerID: "DSN001",
			Department: "Informatika",
			User: model.User{
				FullName: "Dr. Lecturer",
			},
		},
	}

	mockLecturerRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(mockLecturers, int64(1), nil)

	app.Get("/lecturers", academicService.GetAllLecturers)

	req := httptest.NewRequest("GET", "/lecturers?page=1&limit=10", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockLecturerRepo.AssertExpectations(t)
}

func TestGetAllLecturers_Error(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	mockLecturerRepo.On("FindAll", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, int64(0), errors.New("db error"))

	app.Get("/lecturers", academicService.GetAllLecturers)

	req := httptest.NewRequest("GET", "/lecturers?page=1&limit=10", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 500, resp.StatusCode)
	mockLecturerRepo.AssertExpectations(t)
}

func TestGetAdvisees_Success(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	lecturerID := uuid.New()
	studentID := uuid.New()

	mockLecturer := &model.Lecturer{
		ID:         lecturerID,
		LecturerID: "DSN001",
	}

	mockStudents := []model.Student{
		{
			ID:           studentID,
			StudentID:    "MHS001",
			ProgramStudy: "Informatika",
			User: model.User{
				FullName: "Student 1",
			},
		},
	}

	mockLecturerRepo.On("FindByID", lecturerID).Return(mockLecturer, nil)
	mockLecturerRepo.On("GetAdvisees", lecturerID).Return(mockStudents, nil)

	app.Get("/lecturers/:id/advisees", academicService.GetAdvisees)

	req := httptest.NewRequest("GET", "/lecturers/"+lecturerID.String()+"/advisees", nil)
	resp, err := app.Test(req)

	assert.NoError(t, err)
	assert.Equal(t, 200, resp.StatusCode)
	mockLecturerRepo.AssertExpectations(t)
}

func TestGetAdvisees_InvalidID(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	app.Get("/lecturers/:id/advisees", academicService.GetAdvisees)

	req := httptest.NewRequest("GET", "/lecturers/invalid-id/advisees", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 400, resp.StatusCode)
}

func TestGetAdvisees_LecturerNotFound(t *testing.T) {
	app := fiber.New()
	mockStudentRepo := new(MockStudentRepository)
	mockLecturerRepo := new(MockLecturerRepository)
	mockAchievementRepo := new(MockAchievementRepoForAcademic)

	academicService := NewAcademicService(mockStudentRepo, mockLecturerRepo, mockAchievementRepo)

	lecturerID := uuid.New()
	mockLecturerRepo.On("FindByID", lecturerID).Return(nil, errors.New("not found"))

	app.Get("/lecturers/:id/advisees", academicService.GetAdvisees)

	req := httptest.NewRequest("GET", "/lecturers/"+lecturerID.String()+"/advisees", nil)
	resp, _ := app.Test(req)

	assert.Equal(t, 404, resp.StatusCode)
	mockLecturerRepo.AssertExpectations(t)
}
