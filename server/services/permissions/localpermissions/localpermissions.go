// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package localpermissions

import (
	"strings"

	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/services/permissions"

	mmModel "github.com/mattermost/mattermost/server/public/model"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

type userStore interface {
	GetUserByID(userID string) (*model.User, error)
	GetUserPreferences(userID string) (mmModel.Preferences, error)
}

type Service struct {
	store  permissions.Store
	logger mlog.LoggerIFace
}

func New(store permissions.Store, logger mlog.LoggerIFace) *Service {
	return &Service{
		store:  store,
		logger: logger,
	}
}

func (s *Service) HasPermissionTo(userID string, permission *mmModel.Permission) bool {
	if userID == "" || permission == nil {
		return false
	}

	if permission.Id == model.PermissionManageSystem.Id {
		if uStore, ok := s.store.(userStore); ok {
			user, err := uStore.GetUserByID(userID)
			if err == nil && user != nil {
				if user.Username == "admin" || strings.Contains(user.Roles, "system_admin") || strings.Contains(user.Roles, "admin") {
					return true
				}
				prefs, err := uStore.GetUserPreferences(userID)
				if err == nil {
					for _, pref := range prefs {
						if (pref.Name == "roles" && (pref.Value == "system_admin" || pref.Value == "admin")) ||
							(pref.Name == "isAdmin" && pref.Value == "true") {
							return true
						}
					}
				}
			}
		}
	}

	return false
}

func (s *Service) HasPermissionToTeam(userID, teamID string, permission *mmModel.Permission) bool {
	if userID == "" || teamID == "" || permission == nil {
		return false
	}
	if permission.Id == model.PermissionManageTeam.Id {
		return false
	}
	return true
}

func (s *Service) HasPermissionToChannel(userID, channelID string, permission *mmModel.Permission) bool {
	if userID == "" || channelID == "" || permission == nil {
		return false
	}
	return true
}

func (s *Service) HasPermissionToBoard(userID, boardID string, permission *mmModel.Permission) bool {
	if userID == "" || boardID == "" || permission == nil {
		return false
	}

	member, err := s.store.GetMemberForBoard(boardID, userID)
	if model.IsErrNotFound(err) {
		return false
	}
	if err != nil {
		s.logger.Error("error getting member for board",
			mlog.String("boardID", boardID),
			mlog.String("userID", userID),
			mlog.Err(err),
		)
		return false
	}

	switch member.MinimumRole {
	case "admin":
		member.SchemeAdmin = true
	case "editor":
		member.SchemeEditor = true
	case "commenter":
		member.SchemeCommenter = true
	case "viewer":
		member.SchemeViewer = true
	}

	switch permission {
	case model.PermissionManageBoardType, model.PermissionDeleteBoard, model.PermissionManageBoardRoles, model.PermissionShareBoard, model.PermissionDeleteOthersComments:
		return member.SchemeAdmin
	case model.PermissionManageBoardCards, model.PermissionManageBoardProperties:
		return member.SchemeAdmin || member.SchemeEditor
	case model.PermissionCommentBoardCards:
		return member.SchemeAdmin || member.SchemeEditor || member.SchemeCommenter
	case model.PermissionViewBoard:
		return member.SchemeAdmin || member.SchemeEditor || member.SchemeCommenter || member.SchemeViewer
	default:
		return false
	}
}
