package repo

import (
	"database/sql"
	"fiber/skp/app/model"
	"fmt"
	"time"

	"github.com/google/uuid"
)

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
	DB *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		DB: db,
	}
}

var userSortWhitelist = map[string]string{
	"created_at": "u.created_at",
	"username":   "u.username",
	"email":      "u.email",
	"full_name":  "u.full_name",
}

func (r *UserRepo) Create(user *model.User) error {
	query := `
		INSERT INTO users (username, email, password_hash, full_name, role_id, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id`

	now := time.Now()
	return r.DB.QueryRow(
		query,
		user.Username,
		user.Email,
		user.PasswordHash,
		user.FullName,
		user.RoleID,
		true,
		now,
		now,
	).Scan(&user.ID)
}

func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at, u.refresh_token,
		       r.id, r.name, r.description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.username = $1 AND u.is_active = true`

	var user model.User
	var roleID, roleName, roleDesc sql.NullString
	var refreshToken sql.NullString

	err := r.DB.QueryRow(query, username).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &refreshToken,
		&roleID, &roleName, &roleDesc,
	)
	if err != nil {
		return nil, err
	}

	if refreshToken.Valid {
		user.RefreshToken = refreshToken.String
	}

	if roleID.Valid {
		user.Role.ID, _ = uuid.Parse(roleID.String)
		user.Role.Name = roleName.String
		user.Role.Description = roleDesc.String
	}

	if user.RoleID != nil {
		permissions, err := r.getPermissionsForRole(*user.RoleID)
		if err == nil {
			user.Role.Permissions = permissions
		}
	}

	return &user, nil
}

func (r *UserRepo) getPermissionsForRole(roleID uuid.UUID) ([]model.Permission, error) {
	query := `
		SELECT p.id, p.name, p.resource, p.action, p.description
		FROM permissions p
		JOIN role_permissions rp ON p.id = rp.permission_id
		WHERE rp.role_id = $1`

	rows, err := r.DB.Query(query, roleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var permissions []model.Permission
	for rows.Next() {
		var p model.Permission
		var desc sql.NullString
		if err := rows.Scan(&p.ID, &p.Name, &p.Resource, &p.Action, &desc); err != nil {
			return nil, err
		}
		if desc.Valid {
			p.Description = desc.String
		}
		permissions = append(permissions, p)
	}
	return permissions, nil
}

func (r *UserRepo) FindByUserID(id uuid.UUID) (*model.User, error) {
	query := `
		SELECT u.id, u.username, u.email, u.password_hash, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at, u.refresh_token,
		       r.id, r.name, r.description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.is_active = true`

	var user model.User
	var roleID, roleName, roleDesc sql.NullString
	var refreshToken sql.NullString

	err := r.DB.QueryRow(query, id).Scan(
		&user.ID, &user.Username, &user.Email, &user.PasswordHash, &user.FullName,
		&user.RoleID, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &refreshToken,
		&roleID, &roleName, &roleDesc,
	)
	if err != nil {
		return nil, err
	}

	if refreshToken.Valid {
		user.RefreshToken = refreshToken.String
	}

	if roleID.Valid {
		user.Role.ID, _ = uuid.Parse(roleID.String)
		user.Role.Name = roleName.String
		user.Role.Description = roleDesc.String
	}

	if user.RoleID != nil {
		permissions, err := r.getPermissionsForRole(*user.RoleID)
		if err == nil {
			user.Role.Permissions = permissions
		}
	}

	return &user, nil
}

func (r *UserRepo) FindAll(page, limit int, search, sortBy, order string) ([]model.User, int64, error) {
	var total int64

	countQuery := `SELECT COUNT(*) FROM users WHERE is_active = true`
	args := []interface{}{}
	argIndex := 1

	if search != "" {
		countQuery += fmt.Sprintf(" AND (username ILIKE $%d OR email ILIKE $%d OR full_name ILIKE $%d)", argIndex, argIndex, argIndex)
		args = append(args, "%"+search+"%")
		argIndex++
	}

	err := r.DB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	query := `
		SELECT u.id, u.username, u.email, u.full_name, u.role_id, u.is_active, u.created_at, u.updated_at,
		       r.id, r.name, r.description
		FROM users u
		LEFT JOIN roles r ON u.role_id = r.id
		WHERE u.is_active = true`

	selectArgs := []interface{}{}
	selectArgIndex := 1

	if search != "" {
		query += fmt.Sprintf(" AND (u.username ILIKE $%d OR u.email ILIKE $%d OR u.full_name ILIKE $%d)", selectArgIndex, selectArgIndex, selectArgIndex)
		selectArgs = append(selectArgs, "%"+search+"%")
		selectArgIndex++
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}
	if sortColumn, ok := userSortWhitelist[sortBy]; ok {
		query += fmt.Sprintf(" ORDER BY %s %s", sortColumn, order)
	} else {
		query += " ORDER BY u.created_at DESC"
	}

	offset := (page - 1) * limit
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", selectArgIndex, selectArgIndex+1)
	selectArgs = append(selectArgs, limit, offset)

	rows, err := r.DB.Query(query, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var users []model.User
	for rows.Next() {
		var u model.User
		var roleID, roleName, roleDesc sql.NullString

		if err := rows.Scan(
			&u.ID, &u.Username, &u.Email, &u.FullName, &u.RoleID, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
			&roleID, &roleName, &roleDesc,
		); err != nil {
			return nil, 0, err
		}

		if roleID.Valid {
			u.Role.ID, _ = uuid.Parse(roleID.String)
			u.Role.Name = roleName.String
			u.Role.Description = roleDesc.String
		}

		users = append(users, u)
	}

	return users, total, nil
}

func (r *UserRepo) Update(user *model.User) error {
	query := `
		UPDATE users 
		SET username = $1, email = $2, password_hash = $3, full_name = $4, refresh_token = $5, updated_at = $6
		WHERE id = $7`

	_, err := r.DB.Exec(query, user.Username, user.Email, user.PasswordHash, user.FullName, user.RefreshToken, time.Now(), user.ID)
	return err
}

func (r *UserRepo) Delete(id uuid.UUID) error {
	query := `UPDATE users SET is_active = false WHERE id = $1`
	_, err := r.DB.Exec(query, id)
	return err
}

func (r *UserRepo) UpdateRole(userID uuid.UUID, roleID uuid.UUID) error {
	query := `UPDATE users SET role_id = $1, updated_at = $2 WHERE id = $3 AND is_active = true`
	result, err := r.DB.Exec(query, roleID, time.Now(), userID)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("User tidak ditemukan atau user tidak aktif")
	}
	return nil
}

func (r *UserRepo) AddBlacklistToken(token model.BlacklistedToken) error {
	query := `INSERT INTO blacklisted_tokens (token, expires_at, created_at) VALUES ($1, $2, $3)`
	_, err := r.DB.Exec(query, token.Token, token.ExpiresAt, time.Now())
	return err
}

func (r *UserRepo) ClearRefreshToken(userID uuid.UUID) error {
	query := `UPDATE users SET refresh_token = '' WHERE id = $1`
	_, err := r.DB.Exec(query, userID)
	return err
}

func (r *UserRepo) FindRoleByName(name string) (*model.Role, error) {
	query := `SELECT id, name, description, created_at FROM roles WHERE name = $1`
	var role model.Role
	var desc sql.NullString
	err := r.DB.QueryRow(query, name).Scan(&role.ID, &role.Name, &desc, &role.CreatedAt)
	if err != nil {
		return nil, err
	}
	if desc.Valid {
		role.Description = desc.String
	}
	return &role, nil
}
