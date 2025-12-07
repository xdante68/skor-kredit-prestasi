package repo

import (
	"fiber/skp/app/model"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

var userSortWhitelist = map[string]string{
	"created_at": "users.created_at",
	"username":   "users.username",
	"email":      "users.email",
	"full_name":  "users.full_name",
}

type UserRepository interface {
	Create(user *model.User) error
	FindByUsername(username string) (*model.User, error)
	FindByUserID(id uuid.UUID) (*model.User, error)
	FindAll(page, limit int, search, sortBy, order string) ([]model.User, int64, error)
	Update(user *model.User) error
	Delete(id uuid.UUID) error
	UpdateRole(userID uuid.UUID, roleID uuid.UUID) error
	AddBlacklistToken(token model.BlacklistedToken) error
	ClearRefreshToken(userID uuid.UUID) error
	FindRoleByName(name string) (*model.Role, error)
}

type UserRepo struct {
	DB *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{
		DB: db,
	}
}

func (r *UserRepo) Create(user *model.User) error {
	return r.DB.Create(user).Error
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.DB.Preload("Role.Permissions").Where("username = ? AND is_active = ?", username, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindByUserID(id uuid.UUID) (*model.User, error) {
	var user model.User
	err := r.DB.Preload("Role.Permissions").Where("id = ? AND is_active = ?", id, true).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.DB.Model(&model.User{}).Preload("Role").Where("is_active = ?", true)

	if search != "" {
		searchPattern := "%" + search + "%"
		query = query.Where("username ILIKE ? OR email ILIKE ? OR full_name ILIKE ?", searchPattern, searchPattern, searchPattern)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if order != "asc" && order != "desc" {
		order = "desc"
	}

	if sortColumn, ok := userSortWhitelist[sortBy]; ok {
		query = query.Order(fmt.Sprintf("%s %s", sortColumn, order))
	} else {
		query = query.Order("users.created_at desc")
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
	result := r.DB.Model(&model.User{}).
		Where("id = ? AND is_active = ?", userID, true).
		Updates(map[string]interface{}{"role_id": roleID})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user not found or user is inactive")
	}
	return nil
}

func (r *UserRepo) AddBlacklistToken(token model.BlacklistedToken) error {
	return r.DB.Create(&token).Error
}

func (r *UserRepo) ClearRefreshToken(userID uuid.UUID) error {
	return r.DB.Model(&model.User{}).Where("id = ?", userID).Update("refresh_token", "").Error
}

func (r *UserRepo) FindRoleByName(name string) (*model.Role, error) {
	var role model.Role
	err := r.DB.First(&role, "name = ?", name).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}
