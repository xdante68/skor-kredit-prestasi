package repo

import (
	"fiber/skp/app/model"
	"fiber/skp/db"
	"fmt"

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

func (r *UserRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.DB.Preload("Role").Where("is_active = ?", true)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Model(&model.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if sortBy != "" && order != "" {
		query = query.Order(fmt.Sprintf("%s %s", sortBy, order))
	} else {
		query = query.Order("created_at desc")
	}

	err := query.Offset(offset).Limit(limit).Find(&users).Error
	return users, total, err
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
