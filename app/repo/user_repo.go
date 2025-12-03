package repo

import (
	"fiber/skp/app/model"
	"fiber/skp/db"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo() *UserRepo {
	return &UserRepo{
		DB: db.GetDB(),
	}
}

func (r *UserRepo) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Preload("Role").Where("username = ? AND is_active = ?", username, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.DB.Preload("Role").Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindAll() ([]model.User, error) {
	var users []model.User
	err := r.DB.Preload("Role").Where("is_active = ?", true).Find(&users).Error
	return users, err
}

func (r *UserRepo) Update(user *model.User) error {
	return r.DB.Save(user).Error
}

func (r *UserRepo) Delete(id uuid.UUID) error {
	return r.DB.Model(&model.User{}).Where("id = ?", id).Update("is_active", false).Error
}

func (r *UserRepo) UpdateRole(userID uuid.UUID, roleID uuid.UUID) error {
	return r.DB.Model(&model.User{}).Where("id = ?", userID).Update("role_id", roleID).Error
}
