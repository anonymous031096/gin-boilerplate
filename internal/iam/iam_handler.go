package iam

import (
	"context"
	"errors"
	"strconv"

	"gin-boilerplate/internal/shared/auth"
	httpresp "gin-boilerplate/internal/shared/http"
	"gin-boilerplate/internal/shared/middleware"
	"gin-boilerplate/internal/shared/security"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Handler struct {
	db         *pgxpool.Pool
	jwtManager *auth.JWTManager
}

type authUserRow struct {
	ID           uuid.UUID
	Email        string
	FullName     string
	PasswordHash string
}

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailAlreadyExists = errors.New("email already exists")
)

func NewHandler(db *pgxpool.Pool, jwtManager *auth.JWTManager) *Handler {
	return &Handler{db: db, jwtManager: jwtManager}
}

// SignUp godoc
// @Summary Sign up user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignUpRequest true "Sign up payload"
// @Success 201 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Failure 409 {object} http.APIResponse
// @Router /auth/signup [post]
func (h *Handler) SignUp(c *gin.Context) {
	var req SignUpRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}

	var existingUserID uuid.UUID
	err := h.db.QueryRow(c, `SELECT id FROM users WHERE email=$1`, req.Email).Scan(&existingUserID)
	if err == nil {
		httpresp.Conflict(c, ErrEmailAlreadyExists.Error())
		return
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		httpresp.BadRequest(c, err.Error())
		return
	}

	passwordHash, err := security.HashPassword(req.Password)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}

	tx, err := h.db.Begin(c)
	if err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	defer tx.Rollback(c)

	var userID uuid.UUID
	err = tx.QueryRow(c, `
		INSERT INTO users (email,full_name,password_hash)
		VALUES ($1,$2,$3) RETURNING id
	`, req.Email, req.FullName, passwordHash).Scan(&userID)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}

	ct, err := tx.Exec(c, `
	INSERT INTO user_roles (user_id, role_id)
	SELECT $1, r.id FROM roles r WHERE r.code = 'USER'
	ON CONFLICT DO NOTHING
	`, userID)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	if ct.RowsAffected() == 0 {
		httpresp.Internal(c, "failed to assign default role")
		return
	}
	if err := tx.Commit(c); err != nil {
		httpresp.Internal(c, err.Error())
		return
	}

	httpresp.Created(c, nil)
}

// SignIn godoc
// @Summary Sign in user
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body SignInRequest true "Sign in payload"
// @Success 200 {object} http.APIResponse
// @Failure 401 {object} http.APIResponse
// @Failure 500 {object} http.APIResponse
// @Router /auth/signin [post]
func (h *Handler) SignIn(c *gin.Context) {
	var req SignInRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	var u authUserRow
	err := h.db.QueryRow(c, `SELECT id,email,full_name,password_hash FROM users WHERE email=$1`, req.Email).
		Scan(&u.ID, &u.Email, &u.FullName, &u.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httpresp.Unauthorized(c, ErrInvalidCredentials.Error())
			return
		}
		httpresp.Internal(c, err.Error())
		return
	}
	if err := security.ComparePassword(u.PasswordHash, req.Password); err != nil {
		httpresp.Unauthorized(c, ErrInvalidCredentials.Error())
		return
	}

	tokens, err := h.jwtManager.GenerateTokenPair(u.ID)
	if err != nil {
		httpresp.Internal(c, err.Error())
		return
	}

	resp := AuthSignInUpResponse{
		User: UserResponse{
			ID:       u.ID.String(),
			Email:    u.Email,
			FullName: u.FullName,
		},
		Tokens: toTokensResponse(tokens),
	}
	httpresp.OK(c, resp)
}

// Refresh godoc
// @Summary Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token payload"
// @Success 200 {object} http.APIResponse
// @Failure 401 {object} http.APIResponse
// @Router /auth/refresh [post]
func (h *Handler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	claims, err := h.jwtManager.ParseRefreshToken(req.RefreshToken)
	if err != nil {
		httpresp.Unauthorized(c, err.Error())
		return
	}
	userID, err := uuid.Parse(claims.UserID)
	if err != nil {
		httpresp.Unauthorized(c, err.Error())
		return
	}
	tokens, err := h.jwtManager.GenerateTokenPair(userID)
	if err != nil {
		httpresp.Unauthorized(c, err.Error())
		return
	}
	httpresp.OK(c, toTokensResponse(tokens))
}

// GetMe godoc
// @Summary Get current user profile
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Failure 401 {object} http.APIResponse
// @Failure 404 {object} http.APIResponse
// @Router /users/me [get]
func (h *Handler) GetMe(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		httpresp.Unauthorized(c, "unauthorized")
		return
	}
	uid, err := uuid.Parse(userID)
	if err != nil {
		httpresp.Unauthorized(c, "invalid user id")
		return
	}
	var u UserResponse
	err = h.db.QueryRow(c, `SELECT id,email,full_name FROM users WHERE id=$1`, uid).
		Scan(&u.ID, &u.Email, &u.FullName)
	if err != nil {
		httpresp.NotFound(c, "user not found")
		return
	}
	httpresp.OK(c, u)
}

// ListUsers godoc
// @Summary List users
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param limit query int false "Limit"
// @Param offset query int false "Offset"
// @Success 200 {object} http.APIResponse
// @Failure 500 {object} http.APIResponse
// @Router /users [get]
func (h *Handler) ListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	rows, err := h.db.Query(c, `SELECT id,email,full_name FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	defer rows.Close()
	items := make([]UserResponse, 0)
	for rows.Next() {
		var item UserResponse
		if err := rows.Scan(&item.ID, &item.Email, &item.FullName); err != nil {
			httpresp.Internal(c, err.Error())
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	httpresp.OK(c, items)
}

// GetUserByID godoc
// @Summary Get user by id
// @Tags Users
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Failure 404 {object} http.APIResponse
// @Router /users/{id} [get]
func (h *Handler) GetUserByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	var u UserResponse
	err = h.db.QueryRow(c, `SELECT id,email,full_name FROM users WHERE id=$1`, id).
		Scan(&u.ID, &u.Email, &u.FullName)
	if err != nil {
		httpresp.NotFound(c, "user not found")
		return
	}
	httpresp.OK(c, u)
}

// UpdateUser godoc
// @Summary Update user
// @Tags Users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "User ID"
// @Param request body UpdateUserRequest true "Update user payload"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Failure 404 {object} http.APIResponse
// @Failure 500 {object} http.APIResponse
// @Router /users/{id} [put]
func (h *Handler) UpdateUser(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	var req UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	ct, err := h.db.Exec(c, `UPDATE users SET full_name=$2, updated_at=now() WHERE id=$1`, id, req.FullName)
	if err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	if ct.RowsAffected() == 0 {
		httpresp.NotFound(c, "user not found")
		return
	}
	httpresp.OK(c, gin.H{"id": id})
}

// CreateRole godoc
// @Summary Create role
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateRoleRequest true "Create role payload"
// @Success 201 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Router /roles [post]
func (h *Handler) CreateRole(c *gin.Context) {
	var req CreateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	permIDs, err := parseUUIDs(req.PermissionIDs)
	if err != nil {
		httpresp.BadRequest(c, "invalid permission_ids")
		return
	}
	var roleID uuid.UUID
	var role RoleResponse
	err = h.db.QueryRow(c, `
		INSERT INTO roles (code,name)
		VALUES ($1,$2) RETURNING id,code,name
	`, req.Code, req.Name).Scan(&roleID, &role.Code, &role.Name)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	role.ID = roleID.String()

	tx, err := h.db.Begin(c)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	defer tx.Rollback(c)
	if _, err := tx.Exec(c, `DELETE FROM role_permissions WHERE role_id=$1`, roleID); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	for _, pid := range permIDs {
		if _, err := tx.Exec(c, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, roleID, pid); err != nil {
			httpresp.BadRequest(c, err.Error())
			return
		}
	}
	if err := tx.Commit(c); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	httpresp.Created(c, role)
}

// ListRoles godoc
// @Summary List roles
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Failure 500 {object} http.APIResponse
// @Router /roles [get]
func (h *Handler) ListRoles(c *gin.Context) {
	rows, err := h.db.Query(c, `SELECT id,code,name FROM roles ORDER BY created_at DESC`)
	if err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	defer rows.Close()
	items := make([]RoleResponse, 0)
	for rows.Next() {
		var item RoleResponse
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			httpresp.Internal(c, err.Error())
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	httpresp.OK(c, items)
}

// UpdateRole godoc
// @Summary Update role
// @Tags Roles
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID"
// @Param request body UpdateRoleRequest true "Update role payload"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Router /roles/{id} [put]
func (h *Handler) UpdateRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	var req UpdateRoleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	permIDs, err := parseUUIDs(req.PermissionIDs)
	if err != nil {
		httpresp.BadRequest(c, "invalid permission_ids")
		return
	}
	ct, err := h.db.Exec(c, `UPDATE roles SET name=$2, updated_at=now() WHERE id=$1`, id, req.Name)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	if ct.RowsAffected() == 0 {
		httpresp.BadRequest(c, "role not found")
		return
	}

	tx, err := h.db.Begin(c)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	defer tx.Rollback(c)
	if _, err := tx.Exec(c, `DELETE FROM role_permissions WHERE role_id=$1`, id); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	for _, pid := range permIDs {
		if _, err := tx.Exec(c, `INSERT INTO role_permissions (role_id, permission_id) VALUES ($1,$2) ON CONFLICT DO NOTHING`, id, pid); err != nil {
			httpresp.BadRequest(c, err.Error())
			return
		}
	}
	if err := tx.Commit(c); err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	httpresp.OK(c, gin.H{"id": id})
}

// DeleteRole godoc
// @Summary Delete role
// @Tags Roles
// @Produce json
// @Security BearerAuth
// @Param id path string true "Role ID"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Router /roles/{id} [delete]
func (h *Handler) DeleteRole(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	ct, err := h.db.Exec(c, `DELETE FROM roles WHERE id=$1`, id)
	if err != nil {
		httpresp.BadRequest(c, err.Error())
		return
	}
	if ct.RowsAffected() == 0 {
		httpresp.BadRequest(c, "role not found")
		return
	}
	httpresp.OK(c, gin.H{"id": id})
}

// ListPermissions godoc
// @Summary List permissions
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Success 200 {object} http.APIResponse
// @Failure 500 {object} http.APIResponse
// @Router /permissions [get]
func (h *Handler) ListPermissions(c *gin.Context) {
	rows, err := h.db.Query(c, `SELECT id,code,name FROM permissions ORDER BY code ASC`)
	if err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	defer rows.Close()
	items := make([]PermissionResponse, 0)
	for rows.Next() {
		var item PermissionResponse
		if err := rows.Scan(&item.ID, &item.Code, &item.Name); err != nil {
			httpresp.Internal(c, err.Error())
			return
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		httpresp.Internal(c, err.Error())
		return
	}
	httpresp.OK(c, items)
}

// GetPermissionByID godoc
// @Summary Get permission by id
// @Tags Permissions
// @Produce json
// @Security BearerAuth
// @Param id path string true "Permission ID"
// @Success 200 {object} http.APIResponse
// @Failure 400 {object} http.APIResponse
// @Failure 404 {object} http.APIResponse
// @Router /permissions/{id} [get]
func (h *Handler) GetPermissionByID(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		httpresp.BadRequest(c, "invalid id")
		return
	}
	var item PermissionResponse
	err = h.db.QueryRow(c, `SELECT id,code,name FROM permissions WHERE id=$1`, id).
		Scan(&item.ID, &item.Code, &item.Name)
	if err != nil {
		httpresp.NotFound(c, "permission not found")
		return
	}
	httpresp.OK(c, item)
}

func (h *Handler) HasPermission(ctx context.Context, userID string, permissionCode string) (bool, error) {
	uid, err := uuid.Parse(userID)
	if err != nil {
		return false, err
	}

	var exists bool
	err = h.db.QueryRow(ctx, `
	SELECT EXISTS(
		SELECT 1
		FROM permissions p
		JOIN user_permissions up ON up.permission_id = p.id
		WHERE up.user_id = $1 AND p.code = $2
		UNION
		SELECT 1
		FROM permissions p
		JOIN role_permissions rp ON rp.permission_id = p.id
		JOIN user_roles ur ON ur.role_id = rp.role_id
		WHERE ur.user_id = $1 AND p.code = $2
	)
	`, uid, permissionCode).Scan(&exists)
	return exists, err
}

func toTokensResponse(tokens auth.TokenPair) AuthTokensResponse {
	return AuthTokensResponse{
		AccessToken:  tokens.AccessToken,
		RefreshToken: tokens.RefreshToken,
	}
}

func parseUUIDs(values []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(values))
	for _, v := range values {
		id, err := uuid.Parse(v)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
