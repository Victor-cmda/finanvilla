// internal/domain/services/user_service.go
package services

import (
	"context"
	"finanvilla/internal/domain/entities"
	"finanvilla/internal/domain/enums"
	"finanvilla/internal/domain/repositories"
	"finanvilla/pkg/errors"

	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepo repositories.UserRepository
}

func NewUserService(userRepo repositories.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) CreateUser(ctx context.Context, user *entities.User) error {
	// Hash da senha
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(hashedPassword)

	// Configurações padrão
	user.Settings = &entities.UserSettings{
		Theme:    "light",
		Language: "pt-BR",
		Currency: "BRL",
	}

	// Definir permissões baseadas no tipo de usuário
	permissions := getDefaultPermissions(user.UserType)
	user.Permissions = permissions

	return s.userRepo.Create(ctx, user)
}

func (s *UserService) UpdateUser(ctx context.Context, user *entities.User) error {
	existingUser, err := s.userRepo.GetByID(ctx, user.ID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Se uma nova senha foi fornecida, faz o hash
	if user.Password != "" {
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.Password = string(hashedPassword)
	} else {
		// Mantém a senha antiga se nenhuma nova senha foi fornecida
		user.Password = existingUser.Password
	}

	return s.userRepo.Update(ctx, user)
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	_, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return errors.ErrUserNotFound
	}
	return s.userRepo.Delete(ctx, id)
}

func (s *UserService) GetByID(ctx context.Context, id string) (*entities.User, error) {
	user, err := s.userRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}
	return user, nil
}

func (s *UserService) List(ctx context.Context, page, limit int) ([]entities.User, int, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	return s.userRepo.List(ctx, page, limit)
}

func (s *UserService) UpdateSettings(ctx context.Context, userID string, settings *entities.UserSettings) error {
	// Verifica se o usuário existe
	_, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	settings.UserID = userID
	return s.userRepo.UpdateSettings(ctx, settings)
}

func (s *UserService) AddPermissions(ctx context.Context, userID string, permissions []string) error {
	// Verifica se o usuário existe
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Valida as permissões
	for _, p := range permissions {
		if !isValidPermission(p) {
			return errors.ErrInvalidPermission
		}
	}

	return s.userRepo.AddPermissions(ctx, user.ID, permissions)
}

func (s *UserService) RemovePermissions(ctx context.Context, userID string, permissions []string) error {
	// Verifica se o usuário existe
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return errors.ErrUserNotFound
	}

	// Valida as permissões
	for _, p := range permissions {
		if !isValidPermission(p) {
			return errors.ErrInvalidPermission
		}
	}

	return s.userRepo.RemovePermissions(ctx, user.ID, permissions)
}

// Método auxiliar para autenticação
func (s *UserService) Authenticate(ctx context.Context, email, password string) (*entities.User, error) {
	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return nil, errors.ErrInvalidPassword
	}

	return user, nil
}

// Método auxiliar para verificar se um usuário tem uma determinada permissão
func (s *UserService) HasPermission(user *entities.User, permission enums.Permission) bool {
	for _, p := range user.Permissions {
		if p.Name == permission {
			return true
		}
	}
	return false
}

// Função auxiliar para validar permissões
func isValidPermission(permission string) bool {
	validPermissions := map[string]bool{
		string(enums.CreateUser):     true,
		string(enums.UpdateUser):     true,
		string(enums.DeleteUser):     true,
		string(enums.ViewAllUsers):   true,
		string(enums.ManageRoles):    true,
		string(enums.ViewReports):    true,
		string(enums.ManageSettings): true,
	}

	return validPermissions[permission]
}

func getDefaultPermissions(userType enums.UserType) []entities.Permission {
	var permissions []entities.Permission

	// Permissões básicas que todos os usuários têm
	basePermissions := []enums.Permission{
		enums.ViewReports,
	}

	switch userType {
	case enums.Admin:
		adminPermissions := []enums.Permission{
			enums.CreateUser,
			enums.UpdateUser,
			enums.DeleteUser,
			enums.ViewAllUsers,
			enums.ManageRoles,
			enums.ViewReports,
			enums.ManageSettings,
		}
		for _, p := range adminPermissions {
			permissions = append(permissions, entities.Permission{
				Name:        p,
				Description: getPermissionDescription(p),
			})
		}

	case enums.Manager:
		managerPermissions := []enums.Permission{
			enums.ViewAllUsers,
			enums.ViewReports,
			enums.ManageSettings,
		}
		for _, p := range managerPermissions {
			permissions = append(permissions, entities.Permission{
				Name:        p,
				Description: getPermissionDescription(p),
			})
		}

	case enums.Standard:
		for _, p := range basePermissions {
			permissions = append(permissions, entities.Permission{
				Name:        p,
				Description: getPermissionDescription(p),
			})
		}
	}

	return permissions
}

func getPermissionDescription(permission enums.Permission) string {
	descriptions := map[enums.Permission]string{
		enums.CreateUser:     "Permite criar novos usuários",
		enums.UpdateUser:     "Permite atualizar informações de usuários",
		enums.DeleteUser:     "Permite deletar usuários",
		enums.ViewAllUsers:   "Permite visualizar todos os usuários",
		enums.ManageRoles:    "Permite gerenciar papéis e permissões",
		enums.ViewReports:    "Permite visualizar relatórios",
		enums.ManageSettings: "Permite gerenciar configurações do sistema",
	}
	return descriptions[permission]
}
