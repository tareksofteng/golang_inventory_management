package services

import (
	"errors"

	"inventory-api/internal/models"
	"inventory-api/internal/rbac"
	"inventory-api/internal/repositories"
	"inventory-api/pkg/auth"

	"gorm.io/gorm"
)

var (
	ErrUserNotFound   = errors.New("user not found")
	ErrUserEmailTaken = errors.New("email already exists")
	ErrInvalidRole    = errors.New("invalid role")
)

var ErrInvalidPermission = errors.New("invalid permission")

type UserService interface {
	Create(name, email, password, role string, permissions []string) (*models.User, error)
	List(search string, page, perPage int) ([]models.User, int64, error)
	Get(id uint) (*models.User, error)
	Update(id uint, name, email, role string, isActive bool, permissions []string) (*models.User, error)
	ChangePassword(id uint, newPassword string) error
	Disable(id uint) error
}

// normalizePermissions validates each permission and, when none are given,
// falls back to the role's default set.
func normalizePermissions(role string, permissions []string) ([]string, error) {
	for _, p := range permissions {
		if !rbac.IsValidPermission(p) {
			return nil, ErrInvalidPermission
		}
	}
	if len(permissions) == 0 {
		return rbac.EffectivePermissions(role, nil), nil
	}
	return permissions, nil
}

type userService struct {
	repo        repositories.UserRepository
	refreshRepo repositories.RefreshTokenRepository
}

func NewUserService(repo repositories.UserRepository, refreshRepo repositories.RefreshTokenRepository) UserService {
	return &userService{repo: repo, refreshRepo: refreshRepo}
}

func (s *userService) Create(name, email, password, role string, permissions []string) (*models.User, error) {
	if !rbac.IsValidRole(rbac.Role(role)) {
		return nil, ErrInvalidRole
	}
	perms, err := normalizePermissions(role, permissions)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsByEmail(email, 0)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserEmailTaken
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Name:        name,
		Email:       email,
		Password:    hash,
		Role:        role,
		Permissions: perms,
		IsActive:    true,
	}
	if err := s.repo.Create(user); err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userService) List(search string, page, perPage int) ([]models.User, int64, error) {
	offset := (page - 1) * perPage
	return s.repo.FindAll(search, offset, perPage)
}

func (s *userService) Get(id uint) (*models.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

func (s *userService) Update(id uint, name, email, role string, isActive bool, permissions []string) (*models.User, error) {
	if !rbac.IsValidRole(rbac.Role(role)) {
		return nil, ErrInvalidRole
	}
	perms, err := normalizePermissions(role, permissions)
	if err != nil {
		return nil, err
	}

	existing, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.ExistsByEmail(email, id)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserEmailTaken
	}

	existing.Name = name
	existing.Email = email
	existing.Role = role
	existing.Permissions = perms
	existing.IsActive = isActive

	if err := s.repo.Update(existing); err != nil {
		return nil, err
	}
	// If the account was just disabled, revoke its sessions immediately.
	if !isActive {
		_ = s.refreshRepo.DeleteByUserID(id)
	}
	return existing, nil
}

func (s *userService) ChangePassword(id uint, newPassword string) error {
	existing, err := s.Get(id)
	if err != nil {
		return err
	}

	hash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}
	existing.Password = hash
	if err := s.repo.Update(existing); err != nil {
		return err
	}
	// Changing a password logs out all existing sessions.
	_ = s.refreshRepo.DeleteByUserID(id)
	return nil
}

// Disable deactivates a user (cannot log in) and revokes their refresh tokens.
func (s *userService) Disable(id uint) error {
	existing, err := s.Get(id)
	if err != nil {
		return err
	}
	existing.IsActive = false
	if err := s.repo.Update(existing); err != nil {
		return err
	}
	_ = s.refreshRepo.DeleteByUserID(id)
	return nil
}
