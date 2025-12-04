package repo

import (
	"fiber/skp/app/model"
	"fiber/skp/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type StudentRepo struct {
	DB *gorm.DB
}

func NewStudentRepo() *StudentRepo {
	return &StudentRepo{
		DB: db.GetDB(),
	}
}

func (r *StudentRepo) FindByUserID(userID uuid.UUID) (*model.Student, error) {
	var student model.Student
	err := r.DB.Where("user_id = ?", userID).First(&student).Error
	if err != nil {
		return nil, err
	}
	return &student, nil
}
