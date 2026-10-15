package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/mattermost/focalboard/server/model"
	"github.com/mattermost/focalboard/server/services/audit"
	"github.com/mattermost/focalboard/server/utils"
	"github.com/mattermost/mattermost/server/public/shared/mlog"
)

func (a *API) registerUsersRoutes(r *mux.Router) {
	// Users APIs
	r.HandleFunc("/users", a.sessionRequired(a.handleGetUsersList)).Methods("POST")
	r.HandleFunc("/users/me", a.sessionRequired(a.handleGetMe)).Methods("GET")
	r.HandleFunc("/users/me/memberships", a.sessionRequired(a.handleGetMyMemberships)).Methods("GET")
	r.HandleFunc("/users/{userID}", a.sessionRequired(a.handleGetUser)).Methods("GET")
	r.HandleFunc("/users/{userID}/config", a.sessionRequired(a.handleUpdateUserConfig)).Methods(http.MethodPut)
	r.HandleFunc("/users/me/config", a.sessionRequired(a.handleGetUserPreferences)).Methods(http.MethodGet)

	// Admin User Management APIs
	r.HandleFunc("/admin/users", a.systemAdminRequired(a.handleAdminGetUsers)).Methods(http.MethodGet)
	r.HandleFunc("/admin/users", a.systemAdminRequired(a.handleAdminCreateUser)).Methods(http.MethodPost)
	r.HandleFunc("/admin/users/{userID}", a.systemAdminRequired(a.handleAdminUpdateUser)).Methods(http.MethodPut)
	r.HandleFunc("/admin/users/{userID}", a.systemAdminRequired(a.handleAdminDeleteUser)).Methods(http.MethodDelete)
	r.HandleFunc("/admin/users/{userID}/password", a.systemAdminRequired(a.handleAdminResetPassword)).Methods(http.MethodPost)
	r.HandleFunc("/admin/user_boards", a.systemAdminRequired(a.handleAdminGetBoardsList)).Methods(http.MethodGet)
}

func (a *API) handleGetUsersList(w http.ResponseWriter, r *http.Request) {
	// swagger:operation POST /users getUsersList
	//
	// Returns a user[]
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: userID
	//   in: path
	//   description: User ID
	//   required: true
	//   type: string
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//     schema:
	//       "$ref": "#/definitions/User"
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var userIDs []string
	if err = json.Unmarshal(requestBody, &userIDs); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	auditRec := a.makeAuditRecord(r, "getUsersList", audit.Fail)
	defer a.audit.LogRecord(audit.LevelAuth, auditRec)

	var users []*model.User
	var error error

	if len(userIDs) == 0 {
		a.errorResponse(w, r, model.NewErrBadRequest("User IDs are empty"))
		return
	}

	if userIDs[0] == model.SingleUser {
		ws, _ := a.app.GetRootTeam()
		now := utils.GetMillis()
		user := &model.User{
			ID:       model.SingleUser,
			Username: model.SingleUser,
			Email:    model.SingleUser,
			CreateAt: ws.UpdateAt,
			UpdateAt: now,
		}
		users = append(users, user)
	} else {
		users, error = a.app.GetUsersList(userIDs)
		if error != nil {
			a.errorResponse(w, r, error)
			return
		}
	}

	ctx := r.Context()
	session := ctx.Value(sessionContextKey).(*model.Session)
	isSystemAdmin := a.permissions.HasPermissionTo(session.UserID, model.PermissionManageSystem)

	sanitizedUsers := make([]*model.User, 0)
	for _, user := range users {
		canSeeUser, err2 := a.app.CanSeeUser(session.UserID, user.ID)
		if err2 != nil {
			a.errorResponse(w, r, err2)
			return
		}
		if !canSeeUser {
			continue
		}
		if user.ID == session.UserID {
			user.Sanitize(map[string]bool{})
		} else {
			a.app.SanitizeProfile(user, isSystemAdmin)
		}
		sanitizedUsers = append(sanitizedUsers, user)
	}

	usersList, err := json.Marshal(sanitizedUsers)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonStringResponse(w, http.StatusOK, string(usersList))
	auditRec.Success()
}

func (a *API) handleGetMe(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /users/me getMe
	//
	// Returns the currently logged-in user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: teamID
	//   in: path
	//   description: Team ID
	//   required: false
	//   type: string
	// - name: channelID
	//   in: path
	//   description: Channel ID
	//   required: false
	//   type: string
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//     schema:
	//       "$ref": "#/definitions/User"
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"
	query := r.URL.Query()
	teamID := query.Get("teamID")
	channelID := query.Get("channelID")

	userID := getUserID(r)

	var user *model.User
	var err error

	auditRec := a.makeAuditRecord(r, "getMe", audit.Fail)
	defer a.audit.LogRecord(audit.LevelRead, auditRec)

	if userID == model.SingleUser {
		ws, _ := a.app.GetRootTeam()
		now := utils.GetMillis()
		user = &model.User{
			ID:       model.SingleUser,
			Username: model.SingleUser,
			Email:    model.SingleUser,
			CreateAt: ws.UpdateAt,
			UpdateAt: now,
		}
	} else {
		user, err = a.app.GetUser(userID)
		if err != nil {
			// ToDo: wrap with an invalid token error
			a.errorResponse(w, r, err)
			return
		}
	}

	if teamID != "" && a.permissions.HasPermissionToTeam(userID, teamID, model.PermissionManageTeam) {
		user.Permissions = append(user.Permissions, model.PermissionManageTeam.Id)
	}
	if a.permissions.HasPermissionTo(userID, model.PermissionManageSystem) {
		user.Permissions = append(user.Permissions, model.PermissionManageSystem.Id)
		user.Roles = "system_admin"
	}
	if channelID != "" && a.permissions.HasPermissionToChannel(userID, channelID, model.PermissionCreatePost) {
		user.Permissions = append(user.Permissions, model.PermissionCreatePost.Id)
	}

	user.Sanitize(map[string]bool{})
	userData, err := json.Marshal(user)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}
	jsonBytesResponse(w, http.StatusOK, userData)

	auditRec.AddMeta("userID", user.ID)
	auditRec.Success()
}

func (a *API) handleGetMyMemberships(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /users/me/memberships getMyMemberships
	//
	// Returns the currently users board memberships
	//
	// ---
	// produces:
	// - application/json
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//     schema:
	//       type: array
	//       items:
	//         "$ref": "#/definitions/BoardMember"
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	userID := getUserID(r)

	auditRec := a.makeAuditRecord(r, "getMyBoardMemberships", audit.Fail)
	auditRec.AddMeta("userID", userID)
	defer a.audit.LogRecord(audit.LevelRead, auditRec)

	members, err := a.app.GetMembersForUser(userID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	membersData, err := json.Marshal(members)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, membersData)

	auditRec.Success()
}

func (a *API) handleGetUser(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /users/{userID} getUser
	//
	// Returns a user
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: userID
	//   in: path
	//   description: User ID
	//   required: true
	//   type: string
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//     schema:
	//       "$ref": "#/definitions/User"
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	vars := mux.Vars(r)
	userID := vars["userID"]

	auditRec := a.makeAuditRecord(r, "postBlocks", audit.Fail)
	defer a.audit.LogRecord(audit.LevelRead, auditRec)
	auditRec.AddMeta("userID", userID)

	user, err := a.app.GetUser(userID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	ctx := r.Context()
	session := ctx.Value(sessionContextKey).(*model.Session)

	canSeeUser, err := a.app.CanSeeUser(session.UserID, userID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}
	if !canSeeUser {
		a.errorResponse(w, r, model.NewErrNotFound("user ID="+userID))
		return
	}

	if userID == session.UserID {
		user.Sanitize(map[string]bool{})
	} else {
		a.app.SanitizeProfile(user, a.permissions.HasPermissionTo(session.UserID, model.PermissionManageSystem))
	}

	userData, err := json.Marshal(user)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, userData)
	auditRec.Success()
}

func (a *API) handleUpdateUserConfig(w http.ResponseWriter, r *http.Request) {
	// swagger:operation PATCH /users/{userID}/config updateUserConfig
	//
	// Updates user config
	//
	// ---
	// produces:
	// - application/json
	// parameters:
	// - name: userID
	//   in: path
	//   description: User ID
	//   required: true
	//   type: string
	// - name: Body
	//   in: body
	//   description: User config patch to apply
	//   required: true
	//   schema:
	//     "$ref": "#/definitions/UserPreferencesPatch"
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var patch *model.UserPreferencesPatch
	err = json.Unmarshal(requestBody, &patch)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	vars := mux.Vars(r)
	userID := vars["userID"]

	ctx := r.Context()
	session := ctx.Value(sessionContextKey).(*model.Session)

	auditRec := a.makeAuditRecord(r, "updateUserConfig", audit.Fail)
	defer a.audit.LogRecord(audit.LevelModify, auditRec)

	// a user can update only own config
	if userID != session.UserID {
		a.errorResponse(w, r, model.NewErrForbidden(""))
		return
	}

	updatedConfig, err := a.app.UpdateUserConfig(userID, *patch)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	data, err := json.Marshal(updatedConfig)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, data)
	auditRec.Success()
}

func (a *API) handleGetUserPreferences(w http.ResponseWriter, r *http.Request) {
	// swagger:operation GET /users/me/config getUserConfig
	//
	// Returns an array of user preferences
	//
	// ---
	// produces:
	// - application/json
	// security:
	// - BearerAuth: []
	// responses:
	//   '200':
	//     description: success
	//     schema:
	//       "$ref": "#/definitions/Preferences"
	//   default:
	//     description: internal error
	//     schema:
	//       "$ref": "#/definitions/ErrorResponse"

	userID := getUserID(r)

	auditRec := a.makeAuditRecord(r, "getUserConfig", audit.Fail)
	defer a.audit.LogRecord(audit.LevelRead, auditRec)

	preferences, err := a.app.GetUserPreferences(userID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	data, err := json.Marshal(preferences)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, data)
	auditRec.Success()
}

func (a *API) handleAdminGetUsers(w http.ResponseWriter, r *http.Request) {
	auditRec := a.makeAuditRecord(r, "adminGetUsers", audit.Fail)
	defer a.audit.LogRecord(audit.LevelRead, auditRec)

	users, err := a.app.GetAllUsers()
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	for _, u := range users {
		a.app.SanitizeProfile(u, true)
	}

	userData, err := json.Marshal(users)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, userData)
	auditRec.Success()
}

type AdminCreateUserRequest struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Roles    string   `json:"roles"`
	BoardID  string   `json:"board_id"`
	BoardIDs []string `json:"board_ids"`
}

func (a *API) handleAdminCreateUser(w http.ResponseWriter, r *http.Request) {
	auditRec := a.makeAuditRecord(r, "adminCreateUser", audit.Fail)
	defer a.audit.LogRecord(audit.LevelAuth, auditRec)

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var req AdminCreateUserRequest
	if err = json.Unmarshal(requestBody, &req); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	req.BoardID = strings.TrimSpace(req.BoardID)

	if req.Username == "" {
		a.errorResponse(w, r, model.NewErrBadRequest("username is required"))
		return
	}
	if req.Email == "" {
		a.errorResponse(w, r, model.NewErrBadRequest("email is required"))
		return
	}
	if len(req.Password) < 6 {
		a.errorResponse(w, r, model.NewErrBadRequest("password must be at least 6 characters"))
		return
	}
	if req.BoardID == "" && len(req.BoardIDs) == 0 {
		a.errorResponse(w, r, model.NewErrBadRequest("É obrigatório selecionar um quadro para compartilhar com o usuário"))
		return
	}

	existingUser, err := a.app.GetUserByUsername(req.Username)
	if err != nil && !model.IsErrNotFound(err) {
		a.errorResponse(w, r, err)
		return
	}
	if existingUser != nil {
		a.errorResponse(w, r, model.NewErrBadRequest("The username already exists"))
		return
	}

	existingUser, err = a.app.GetUserByEmail(req.Email)
	if err != nil && !model.IsErrNotFound(err) {
		a.errorResponse(w, r, err)
		return
	}
	if existingUser != nil {
		a.errorResponse(w, r, model.NewErrBadRequest("The email already exists"))
		return
	}

	if err := a.app.RegisterUser(req.Username, req.Email, req.Password); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	createdUser, err := a.app.GetUserByUsername(req.Username)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	if req.Roles == "system_admin" || req.Roles == "member" {
		if err := a.app.SetUserRole(createdUser.ID, req.Roles); err != nil {
			a.logger.Error("Failed to set user role", mlog.Err(err))
		}
		createdUser.Roles = req.Roles
	} else {
		createdUser.Roles = "member"
	}

	// Compartilhamento obrigatório do(s) quadro(s) selecionado(s) com o novo usuário
	adminUserID := getUserID(r)
	targetBoardIDs := []string{}
	if len(req.BoardIDs) > 0 {
		targetBoardIDs = req.BoardIDs
	} else if req.BoardID != "" {
		if req.BoardID == "all" {
			boards, bErr := a.app.GetBoardsForUserAndTeam(adminUserID, "0", true)
			if bErr == nil {
				for _, b := range boards {
					targetBoardIDs = append(targetBoardIDs, b.ID)
				}
			}
		} else {
			targetBoardIDs = append(targetBoardIDs, req.BoardID)
		}
	}

	for _, bID := range targetBoardIDs {
		if bID == "" {
			continue
		}
		newBoardMember := &model.BoardMember{
			BoardID:         bID,
			UserID:          createdUser.ID,
			SchemeAdmin:     req.Roles == "system_admin",
			SchemeEditor:    true,
			SchemeCommenter: true,
			SchemeViewer:    true,
		}
		if _, err := a.app.AddMemberToBoard(newBoardMember); err != nil {
			a.logger.Error("Failed to add new user to board", mlog.String("boardID", bID), mlog.String("userID", createdUser.ID), mlog.Err(err))
		}
	}

	a.app.SanitizeProfile(createdUser, true)
	userData, err := json.Marshal(createdUser)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, userData)
	auditRec.Success()
}

func (a *API) handleAdminGetBoardsList(w http.ResponseWriter, r *http.Request) {
	adminUserID := getUserID(r)
	boards, err := a.app.GetBoardsForUserAndTeam(adminUserID, "0", true)
	if err != nil {
		boards = []*model.Board{}
	}

	teams, err := a.app.GetTeamsForUser(adminUserID)
	if err == nil {
		boardIDsMap := make(map[string]bool)
		for _, b := range boards {
			boardIDsMap[b.ID] = true
		}
		for _, t := range teams {
			if t.ID == "0" {
				continue
			}
			teamBoards, tErr := a.app.GetBoardsForUserAndTeam(adminUserID, t.ID, true)
			if tErr == nil {
				for _, tb := range teamBoards {
					if !boardIDsMap[tb.ID] {
						boardIDsMap[tb.ID] = true
						boards = append(boards, tb)
					}
				}
			}
		}
	}

	data, err := json.Marshal(boards)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, data)
}

type AdminUpdateUserRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Roles    string `json:"roles"`
}

func (a *API) handleAdminUpdateUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetUserID := vars["userID"]

	auditRec := a.makeAuditRecord(r, "adminUpdateUser", audit.Fail)
	defer a.audit.LogRecord(audit.LevelAuth, auditRec)

	user, err := a.app.GetUser(targetUserID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var req AdminUpdateUserRequest
	if err = json.Unmarshal(requestBody, &req); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)

	if user.Username == "admin" && req.Username != "" && req.Username != "admin" {
		a.errorResponse(w, r, model.NewErrBadRequest("cannot change root admin username"))
		return
	}

	if req.Username != "" && req.Username != user.Username {
		existing, err := a.app.GetUserByUsername(req.Username)
		if err != nil && !model.IsErrNotFound(err) {
			a.errorResponse(w, r, err)
			return
		}
		if existing != nil && existing.ID != user.ID {
			a.errorResponse(w, r, model.NewErrBadRequest("The username already exists"))
			return
		}
		user.Username = req.Username
	}

	if req.Email != "" && req.Email != user.Email {
		existing, err := a.app.GetUserByEmail(req.Email)
		if err != nil && !model.IsErrNotFound(err) {
			a.errorResponse(w, r, err)
			return
		}
		if existing != nil && existing.ID != user.ID {
			a.errorResponse(w, r, model.NewErrBadRequest("The email already exists"))
			return
		}
		user.Email = req.Email
	}

	updatedUser, err := a.app.UpdateUser(user)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	if req.Roles != "" {
		if user.Username == "admin" && req.Roles != "system_admin" {
			a.errorResponse(w, r, model.NewErrBadRequest("cannot remove admin role from root admin user"))
			return
		}
		if err := a.app.SetUserRole(user.ID, req.Roles); err != nil {
			a.errorResponse(w, r, err)
			return
		}
		updatedUser.Roles = req.Roles
	} else {
		if a.permissions.HasPermissionTo(user.ID, model.PermissionManageSystem) {
			updatedUser.Roles = "system_admin"
		} else {
			updatedUser.Roles = "member"
		}
	}

	a.app.SanitizeProfile(updatedUser, true)
	userData, err := json.Marshal(updatedUser)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonBytesResponse(w, http.StatusOK, userData)
	auditRec.Success()
}

func (a *API) handleAdminDeleteUser(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetUserID := vars["userID"]
	currentUserID := getUserID(r)

	auditRec := a.makeAuditRecord(r, "adminDeleteUser", audit.Fail)
	defer a.audit.LogRecord(audit.LevelAuth, auditRec)

	if targetUserID == currentUserID {
		a.errorResponse(w, r, model.NewErrBadRequest("cannot delete your own account"))
		return
	}

	targetUser, err := a.app.GetUser(targetUserID)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	if targetUser.Username == "admin" {
		a.errorResponse(w, r, model.NewErrBadRequest("cannot delete the root admin user"))
		return
	}

	if err := a.app.DeleteUser(targetUserID); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonStringResponse(w, http.StatusOK, `{"status":"ok"}`)
	auditRec.Success()
}

type AdminResetPasswordRequest struct {
	Password string `json:"password"`
}

func (a *API) handleAdminResetPassword(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	targetUserID := vars["userID"]

	auditRec := a.makeAuditRecord(r, "adminResetPassword", audit.Fail)
	defer a.audit.LogRecord(audit.LevelAuth, auditRec)

	requestBody, err := io.ReadAll(r.Body)
	if err != nil {
		a.errorResponse(w, r, err)
		return
	}

	var req AdminResetPasswordRequest
	if err = json.Unmarshal(requestBody, &req); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	if len(req.Password) < 6 {
		a.errorResponse(w, r, model.NewErrBadRequest("password must be at least 6 characters"))
		return
	}

	if err := a.app.UpdateUserPasswordByID(targetUserID, req.Password); err != nil {
		a.errorResponse(w, r, err)
		return
	}

	jsonStringResponse(w, http.StatusOK, `{"status":"ok"}`)
	auditRec.Success()
}

