package repo

import (
	"database/sql"
	"fiber/skp/app/model"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type StudentRepository interface {
	Create(student *model.Student) error
	FindAll(page, limit int, search, sortBy, order string) ([]model.Student, int64, error)
	FindByID(id uuid.UUID) (*model.Student, error)
	FindByUserID(userID uuid.UUID) (*model.Student, error)
	UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error
	ExistsByStudentID(studentID string) (bool, error)
	DeleteByUserID(userID uuid.UUID) error
}

type StudentRepo struct {
	DB *sql.DB
}

func NewStudentRepo(db *sql.DB) *StudentRepo {
	return &StudentRepo{
		DB: db,
	}
}

var studentSortWhitelist = map[string]string{
	"created_at":    "s.created_at",
	"full_name":     "u.full_name",
	"student_id":    "s.student_id",
	"program_study": "s.program_study",
}

func (r *StudentRepo) Create(student *model.Student) error {
	query := `
		INSERT INTO students (user_id, student_id, program_study, academic_year, advisor_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id`

	now := time.Now()
	return r.DB.QueryRow(
		query,
		student.UserID,
		student.StudentID,
		student.ProgramStudy,
		student.AcademicYear,
		student.AdvisorID,
		now,
	).Scan(&student.ID)
}

func (r *StudentRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.Student, int64, error) {
	var total int64

	countQuery := `SELECT COUNT(*) FROM students s JOIN users u ON u.id = s.user_id WHERE u.is_active = true`
	countArgs := []interface{}{}
	argIndex := 1

	if search != "" {
		countQuery += fmt.Sprintf(" AND (u.full_name ILIKE $%d OR s.student_id ILIKE $%d OR s.program_study ILIKE $%d)", argIndex, argIndex, argIndex)
		countArgs = append(countArgs, "%"+search+"%")
		argIndex++
	}

	err := r.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
		       u.username, u.email, u.full_name,
		       a.id, a.lecturer_id,
		       au.full_name
		FROM students s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN lecturers a ON a.id = s.advisor_id
		LEFT JOIN users au ON au.id = a.user_id
		WHERE u.is_active = true`

	selectArgs := []interface{}{}
	selectArgIndex := 1

	if search != "" {
		query += fmt.Sprintf(" AND (u.full_name ILIKE $%d OR s.student_id ILIKE $%d OR s.program_study ILIKE $%d)", selectArgIndex, selectArgIndex, selectArgIndex)
		selectArgs = append(selectArgs, "%"+search+"%")
		selectArgIndex++
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}
	if sortColumn, ok := studentSortWhitelist[sortBy]; ok {
		query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, order)
	} else {
		query += " ORDER BY s.created_at DESC"
	}

	offset := (page - 1) * limit
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", selectArgIndex, selectArgIndex+1)
	selectArgs = append(selectArgs, limit, offset)

	rows, err := r.DB.Query(query, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		var userName, userEmail, userFullName sql.NullString
		var advisorID, advisorLecturerID sql.NullString
		var advisorUserFullName sql.NullString

		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
			&userName, &userEmail, &userFullName,
			&advisorID, &advisorLecturerID,
			&advisorUserFullName,
		); err != nil {
			return nil, 0, err
		}

		s.User.ID = s.UserID 
		if userName.Valid {
			s.User.Username = userName.String
			s.User.Email = userEmail.String
			s.User.FullName = userFullName.String
		}

		if advisorID.Valid {
			s.Advisor = &model.Lecturer{}
			s.Advisor.ID, _ = uuid.Parse(advisorID.String)
			s.Advisor.LecturerID = advisorLecturerID.String
			if advisorUserFullName.Valid {
				s.Advisor.User.FullName = advisorUserFullName.String
			}
		}

		students = append(students, s)
	}

	return students, total, nil
}

func (r *StudentRepo) FindByID(id uuid.UUID) (*model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
		       u.username, u.email, u.full_name,
		       a.id, a.lecturer_id,
		       au.full_name
		FROM students s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN lecturers a ON a.id = s.advisor_id
		LEFT JOIN users au ON au.id = a.user_id
		WHERE s.id = $1 AND u.is_active = true`

	var s model.Student
	var userName, userEmail, userFullName sql.NullString
	var advisorID, advisorLecturerID sql.NullString
	var advisorUserFullName sql.NullString

	err := r.DB.QueryRow(query, id).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
		&userName, &userEmail, &userFullName,
		&advisorID, &advisorLecturerID,
		&advisorUserFullName,
	)
	if err != nil {
		return nil, err
	}

	s.User.ID = s.UserID
	if userName.Valid {
		s.User.Username = userName.String
		s.User.Email = userEmail.String
		s.User.FullName = userFullName.String
	}

	if advisorID.Valid {
		s.Advisor = &model.Lecturer{}
		s.Advisor.ID, _ = uuid.Parse(advisorID.String)
		s.Advisor.LecturerID = advisorLecturerID.String
		if advisorUserFullName.Valid {
			s.Advisor.User.FullName = advisorUserFullName.String
		}
	}

	return &s, nil
}

func (r *StudentRepo) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
		       u.username, u.email, u.full_name,
		       a.id, a.lecturer_id,
		       au.full_name
		FROM students s
		JOIN users u ON u.id = s.user_id
		LEFT JOIN lecturers a ON a.id = s.advisor_id
		LEFT JOIN users au ON au.id = a.user_id
		WHERE s.user_id = $1 AND u.is_active = true`

	var s model.Student
	var userName, userEmail, userFullName sql.NullString
	var advisorID, advisorLecturerID sql.NullString
	var advisorUserFullName sql.NullString

	err := r.DB.QueryRow(query, userID).Scan(
		&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
		&userName, &userEmail, &userFullName,
		&advisorID, &advisorLecturerID,
		&advisorUserFullName,
	)
	if err != nil {
		return nil, err
	}

	s.User.ID = s.UserID
	if userName.Valid {
		s.User.Username = userName.String
		s.User.Email = userEmail.String
		s.User.FullName = userFullName.String
	}

	if advisorID.Valid {
		s.Advisor = &model.Lecturer{}
		s.Advisor.ID, _ = uuid.Parse(advisorID.String)
		s.Advisor.LecturerID = advisorLecturerID.String
		if advisorUserFullName.Valid {
			s.Advisor.User.FullName = advisorUserFullName.String
		}
	}

	return &s, nil
}

func (r *StudentRepo) UpdateAdvisor(studentID uuid.UUID, advisorID uuid.UUID) error {
	query := `UPDATE students SET advisor_id = $1 WHERE id = $2`
	_, err := r.DB.Exec(query, advisorID, studentID)
	return err
}

func (r *StudentRepo) ExistsByStudentID(studentID string) (bool, error) {
	var count int64
	query := `SELECT COUNT(*) FROM students WHERE student_id = $1`
	err := r.DB.QueryRow(query, studentID).Scan(&count)
	return count > 0, err
}

func (r *StudentRepo) DeleteByUserID(userID uuid.UUID) error {
	query := `DELETE FROM students WHERE user_id = $1`
	_, err := r.DB.Exec(query, userID)
	return err
}
