package app

import (
	"fmt"

	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/services/auth"
	"github.com/mattermost/focalboard/server/utils"
	mmModel "github.com/mattermost/mattermost/server/public/model"
	"github.com/pkg/errors"
)

func (a *App) GetTeamUsers(teamID string, asGuestID string) ([]*model.User, error) {
	return a.store.GetUsersByTeam(teamID, asGuestID, a.config.ShowEmailAddress, a.config.ShowFullName)
}

func (a *App) SearchTeamUsers(teamID string, searchQuery string, asGuestID string, excludeBots bool) ([]*model.User, error) {
	users, err := a.store.SearchUsersByTeam(teamID, searchQuery, asGuestID, excludeBots, a.config.ShowEmailAddress, a.config.ShowFullName)
	if err != nil {
		return nil, err
	}

	for i, u := range users {
		if a.permissions.HasPermissionToTeam(u.ID, teamID, model.PermissionManageTeam) {
			users[i].Permissions = append(users[i].Permissions, model.PermissionManageTeam.Id)
		}
		if a.permissions.HasPermissionTo(u.ID, model.PermissionManageSystem) {
			users[i].Permissions = append(users[i].Permissions, model.PermissionManageSystem.Id)
		}
	}
	return users, nil
}

func (a *App) UpdateUserConfig(userID string, patch model.UserPreferencesPatch) ([]mmModel.Preference, error) {
	updatedPreferences, err := a.store.PatchUserPreferences(userID, patch)
	if err != nil {
		return nil, err
	}

	return updatedPreferences, nil
}

func (a *App) GetUserPreferences(userID string) ([]mmModel.Preference, error) {
	return a.store.GetUserPreferences(userID)
}

func (a *App) UserIsGuest(userID string) (bool, error) {
	user, err := a.store.GetUserByID(userID)
	if err != nil {
		return false, err
	}
	return user.IsGuest, nil
}

func (a *App) CanSeeUser(seerUser string, seenUser string) (bool, error) {
	isGuest, err := a.UserIsGuest(seerUser)
	if err != nil {
		return false, err
	}
	if isGuest {
		hasSharedChannels, err := a.store.CanSeeUser(seerUser, seenUser)
		if err != nil {
			return false, err
		}
		return hasSharedChannels, nil
	}
	return true, nil
}

func (a *App) SearchUserChannels(teamID string, userID string, query string) ([]*mmModel.Channel, error) {
	channels, err := a.store.SearchUserChannels(teamID, userID, query)
	if err != nil {
		return nil, err
	}

	var writeableChannels []*mmModel.Channel
	for _, channel := range channels {
		if a.permissions.HasPermissionToChannel(userID, channel.Id, model.PermissionCreatePost) {
			writeableChannels = append(writeableChannels, channel)
		}
	}
	return writeableChannels, nil
}

func (a *App) GetChannel(teamID string, channelID string) (*mmModel.Channel, error) {
	return a.store.GetChannel(teamID, channelID)
}

func (a *App) SanitizeProfile(user *model.User, isAdmin bool) {
	options := map[string]bool{}
	if isAdmin {
		options["fullname"] = true
		options["email"] = true
	} else {
		options["fullname"] = a.config.ShowFullName
		options["email"] = a.config.ShowEmailAddress
	}
	user.Sanitize(options)
}

// UserDeleter is an optional interface implemented by stores that support user deletion.
type UserDeleter interface {
	DeleteUser(userID string) error
}

func (a *App) DeleteUser(userID string) error {
	if userDeleter, ok := a.store.(UserDeleter); ok {
		return userDeleter.DeleteUser(userID)
	}
	return errors.New("delete user not supported by store")
}

func (a *App) GetUserByUsername(username string) (*model.User, error) {
	return a.store.GetUserByUsername(username)
}

func (a *App) GetUserByEmail(email string) (*model.User, error) {
	return a.store.GetUserByEmail(email)
}

func (a *App) GetAllUsers() ([]*model.User, error) {
	users, err := a.store.GetUsersByTeam("0", "", true, true)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		if a.permissions.HasPermissionTo(u.ID, model.PermissionManageSystem) {
			u.Roles = "system_admin"
			u.Permissions = append(u.Permissions, model.PermissionManageSystem.Id)
		} else {
			u.Roles = "member"
		}
	}
	return users, nil
}

func (a *App) UpdateUser(user *model.User) (*model.User, error) {
	return a.store.UpdateUser(user)
}

func (a *App) UpdateUserPasswordByID(userID, password string) error {
	passwordSettings := auth.PasswordSettings{
		MinimumLength: 6,
	}
	if err := auth.IsPasswordValid(password, passwordSettings); err != nil {
		return errors.Wrap(err, "invalid password")
	}
	return a.store.UpdateUserPasswordByID(userID, auth.HashPassword(password))
}

func (a *App) SetUserRole(userID string, role string) error {
	isAdmin := "false"
	if role == "system_admin" {
		isAdmin = "true"
	}
	patch := model.UserPreferencesPatch{
		UpdatedFields: map[string]string{
			"roles":   role,
			"isAdmin": isAdmin,
		},
	}
	_, err := a.store.PatchUserPreferences(userID, patch)
	return err
}

func (a *App) EnsureAdminUser() error {
	adminUser, err := a.store.GetUserByUsername("admin")
	if err != nil && !model.IsErrNotFound(err) {
		return err
	}
	if adminUser == nil {
		a.logger.Info("Creating default system admin user (admin)")
		newUser := &model.User{
			ID:          utils.NewID(utils.IDTypeUser),
			Username:    "admin",
			Email:       "admin@example.com",
			Password:    auth.HashPassword("admin123"),
			AuthService: a.config.AuthMode,
		}
		createdUser, err := a.store.CreateUser(newUser)
		if err != nil {
			return errors.Wrap(err, "unable to create admin user")
		}
		patch := model.UserPreferencesPatch{
			UpdatedFields: map[string]string{
				"roles":   "system_admin",
				"isAdmin": "true",
			},
		}
		_, _ = a.store.PatchUserPreferences(createdUser.ID, patch)
		a.logger.Info(fmt.Sprintf("Default admin user created successfully with ID %s", createdUser.ID))
	} else {
		// Ensure existing admin user has system_admin preferences
		patch := model.UserPreferencesPatch{
			UpdatedFields: map[string]string{
				"roles":   "system_admin",
				"isAdmin": "true",
			},
		}
		_, _ = a.store.PatchUserPreferences(adminUser.ID, patch)
	}
	return nil
}

