package repo

import (
	"database/sql"
	"fiber/skp/app/model"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type LecturerRepository interface {
	Create(lecturer *model.Lecturer) error
	FindAll(page, limit int, search, sortBy, order string) ([]model.Lecturer, int64, error)
	FindByID(id uuid.UUID) (*model.Lecturer, error)
	GetAdvisees(advisorID uuid.UUID) ([]model.Student, error)
	ExistsByLecturerID(lecturerID string) (bool, error)
	DeleteByUserID(userID uuid.UUID) error
}

type LecturerRepo struct {
	DB *sql.DB
}

func NewLecturerRepo(db *sql.DB) *LecturerRepo {
	return &LecturerRepo{
		DB: db,
	}
}

func (r *LecturerRepo) Create(lecturer *model.Lecturer) error {
	query := `
		INSERT INTO lecturers (user_id, lecturer_id, department, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id`

	now := time.Now()
	return r.DB.QueryRow(
		query,
		lecturer.UserID,
		lecturer.LecturerID,
		lecturer.Department,
		now,
	).Scan(&lecturer.ID)
}

var lecturerSortWhitelist = map[string]string{
	"created_at":  "l.created_at",
	"full_name":   "u.full_name",
	"lecturer_id": "l.lecturer_id",
	"department":  "l.department",
}

func (r *LecturerRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.Lecturer, int64, error) {
	var total int64

	countQuery := `SELECT COUNT(*) FROM lecturers l JOIN users u ON u.id = l.user_id WHERE u.is_active = true`
	countArgs := []interface{}{}
	argIndex := 1

	if search != "" {
		countQuery += fmt.Sprintf(" AND (u.full_name ILIKE $%d OR l.lecturer_id ILIKE $%d OR l.department ILIKE $%d)", argIndex, argIndex, argIndex)
		countArgs = append(countArgs, "%"+search+"%")
		argIndex++
	}

	err := r.DB.QueryRow(countQuery, countArgs...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
		       u.username, u.email, u.full_name
		FROM lecturers l
		JOIN users u ON u.id = l.user_id
		WHERE u.is_active = true`

	selectArgs := []interface{}{}
	selectArgIndex := 1

	if search != "" {
		query += fmt.Sprintf(" AND (u.full_name ILIKE $%d OR l.lecturer_id ILIKE $%d OR l.department ILIKE $%d)", selectArgIndex, selectArgIndex, selectArgIndex)
		selectArgs = append(selectArgs, "%"+search+"%")
		selectArgIndex++
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}
	if sortColumn, ok := lecturerSortWhitelist[sortBy]; ok {
		query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, order)
	} else {
		query += " ORDER BY l.created_at DESC"
	}

	offset := (page - 1) * limit
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", selectArgIndex, selectArgIndex+1)
	selectArgs = append(selectArgs, limit, offset)

	rows, err := r.DB.Query(query, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var lecturers []model.Lecturer
	for rows.Next() {
		var l model.Lecturer
		var userName, userEmail, userFullName sql.NullString

		if err := rows.Scan(
			&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt,
			&userName, &userEmail, &userFullName,
		); err != nil {
			return nil, 0, err
		}

		l.User.ID = l.UserID
		if userName.Valid {
			l.User.Username = userName.String
			l.User.Email = userEmail.String
			l.User.FullName = userFullName.String
		}

		lecturers = append(lecturers, l)
	}

	return lecturers, total, nil
}

func (r *LecturerRepo) FindByID(id uuid.UUID) (*model.Lecturer, error) {
	query := `
		SELECT l.id, l.user_id, l.lecturer_id, l.department, l.created_at,
		       u.username, u.email, u.full_name
		FROM lecturers l
		JOIN users u ON u.id = l.user_id
		WHERE l.id = $1 AND u.is_active = true`

	var l model.Lecturer
	var userName, userEmail, userFullName sql.NullString

	err := r.DB.QueryRow(query, id).Scan(
		&l.ID, &l.UserID, &l.LecturerID, &l.Department, &l.CreatedAt,
		&userName, &userEmail, &userFullName,
	)
	if err != nil {
		return nil, err
	}

	l.User.ID = l.UserID
	if userName.Valid {
		l.User.Username = userName.String
		l.User.Email = userEmail.String
		l.User.FullName = userFullName.String
	}

	return &l, nil
}

func (r *LecturerRepo) GetAdvisees(advisorID uuid.UUID) ([]model.Student, error) {
	query := `
		SELECT s.id, s.user_id, s.student_id, s.program_study, s.academic_year, s.advisor_id, s.created_at,
		       u.id, u.username, u.email, u.full_name
		FROM students s
		JOIN users u ON u.id = s.user_id
		WHERE s.advisor_id = $1 AND u.is_active = true`

	rows, err := r.DB.Query(query, advisorID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var students []model.Student
	for rows.Next() {
		var s model.Student
		var userID, userName, userEmail, userFullName sql.NullString

		if err := rows.Scan(
			&s.ID, &s.UserID, &s.StudentID, &s.ProgramStudy, &s.AcademicYear, &s.AdvisorID, &s.CreatedAt,
			&userID, &userName, &userEmail, &userFullName,
		); err != nil {
			return nil, err
		}

		if userID.Valid {
			s.User.ID, _ = uuid.Parse(userID.String)
			s.User.Username = userName.String
			s.User.Email = userEmail.String
			s.User.FullName = userFullName.String
		}

		students = append(students, s)
	}

	return students, nil
}

func (r *LecturerRepo) ExistsByLecturerID(lecturerID string) (bool, error) {
	var count int64
	query := `SELECT COUNT(*) FROM lecturers WHERE lecturer_id = $1`
	err := r.DB.QueryRow(query, lecturerID).Scan(&count)
	return count > 0, err
}

func (r *LecturerRepo) DeleteByUserID(userID uuid.UUID) error {
	query := `DELETE FROM lecturers WHERE user_id = $1`
	_, err := r.DB.Exec(query, userID)
	return err
}
