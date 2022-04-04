package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/dranih/go-crud-api/pkg/controller"
	"github.com/dranih/go-crud-api/pkg/database"
	"github.com/dranih/go-crud-api/pkg/record"
	"github.com/dranih/go-crud-api/pkg/utils"
	"golang.org/x/crypto/bcrypt"
)

type DbAuthMiddleware struct {
	GenericMiddleware
	reflection *database.ReflectionService
	db         *database.GenericDB
	ordering   *record.OrderingInfo
}

func NewDbAuth(responder controller.Responder, properties map[string]interface{}, reflection *database.ReflectionService, db *database.GenericDB) *DbAuthMiddleware {
	return &DbAuthMiddleware{GenericMiddleware: GenericMiddleware{Responder: responder, Properties: properties}, reflection: reflection, db: db, ordering: record.NewOrderingInfo()}
}

func (dam *DbAuthMiddleware) Process(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := utils.GetPathSegment(r, 1)
		method := r.Method
		if method == http.MethodPost && map[string]bool{"login": true, "register": true, "password": true}[path] {
			body, err := utils.GetBodyMapData(r)
			if err != nil {
				dam.Responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
				return
			}
			usernameFormFieldName := fmt.Sprint(dam.getProperty("usernameFormField", "username"))
			passwordFormFieldName := fmt.Sprint(dam.getProperty("passwordFormField", "password"))
			newPasswordFormFieldName := fmt.Sprint(dam.getProperty("newPasswordFormField", "newPassword"))
			var username, password, newPassword string
			if v, ok := body[usernameFormFieldName]; ok {
				username = fmt.Sprint(v)
			}
			if v, ok := body[passwordFormFieldName]; ok {
				password = fmt.Sprint(v)
			}
			if v, ok := body[newPasswordFormFieldName]; ok {
				newPassword = fmt.Sprint(v)
			}
			tableName := fmt.Sprint(dam.getProperty("usersTable", "users"))
			table := dam.reflection.GetTable(tableName)
			usernameColumnName := fmt.Sprint(dam.getProperty("usernameColumn", "username"))
			usernameColumn := table.GetColumn(usernameColumnName)
			passwordColumnName := fmt.Sprint(dam.getProperty("passwordColumn", "password"))
			passwordLength := dam.getIntProperty("passwordLength", 12)
			pkName := table.GetPk().GetName()
			registerUser := fmt.Sprint(dam.getProperty("registerUser", ""))
			condition := database.NewColumnCondition(usernameColumn, "eq", username)
			returnedColumns := fmt.Sprint(dam.getProperty("returnedColumns", ""))
			var columnNames []string
			if returnedColumns == "" {
				columnNames = table.GetColumnNames()
			} else {
				columnNames = strings.Split(returnedColumns, ",")
				for i, elem := range columnNames {
					columnNames[i] = strings.TrimSpace(elem)
				}
				columnNames = append(columnNames, passwordColumnName)
				columnNames = utils.RemoveDuplicateStr(columnNames)
			}
			columnOrdering := dam.ordering.GetDefaultColumnOrdering(table)
			if path == "register" {
				if registerUser == "" {
					dam.Responder.Error(record.AUTHENTICATION_FAILED, username, w, "")
					return
				}
				if len(password) < passwordLength {
					dam.Responder.Error(record.PASSWORD_TOO_SHORT, fmt.Sprintf("%d", passwordLength), w, "")
					return
				}
				users := dam.db.SelectAll(table, columnNames, condition, columnOrdering, 0, 1)
				if len(users) >= 1 {
					dam.Responder.Error(record.USER_ALREADY_EXIST, username, w, "")
					return
				}
				data := map[string]interface{}{}
				if registerUser != "1" {
					if err := json.Unmarshal([]byte(registerUser), &data); err != nil {
						dam.Responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
						return
					}
				}
				data[usernameColumnName] = username
				hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
				if err == nil {
					data[passwordColumnName] = string(hash)
				} else {
					dam.Responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
					return
				}
				if _, err := dam.db.CreateSingle(nil, table, data); err != nil {
					dam.Responder.Error(record.INTERNAL_SERVER_ERROR, "", w, "")
					return
				}
				users = dam.db.SelectAll(table, columnNames, condition, columnOrdering, 0, 1)
				for _, user := range users {
					delete(user, passwordColumnName)
					dam.Responder.Success(user, w)
					return
				}
				dam.Responder.Error(record.AUTHENTICATION_FAILED, username, w, "")
				return
			}
			if path == "login" {
				users := dam.db.SelectAll(table, columnNames, condition, columnOrdering, 0, 1)
				for _, user := range users {
					if err := bcrypt.CompareHashAndPassword([]byte(fmt.Sprint(user[passwordColumnName])), []byte(password)); err == nil {
						session := utils.GetSession(w, r)
						delete(user, passwordColumnName)
						session.Values["user"] = user
						if err := session.Save(r, w); err == nil {
							dam.Responder.Success(user, w)
							return
						}
					}
				}
				dam.Responder.Error(record.AUTHENTICATION_FAILED, username, w, "")
				return
			}
			if path == "password" {
				session := utils.GetSession(w, r)
				var sessMapUser map[string]interface{}
				if sessUser, exists := session.Values["user"]; exists {
					sessMapUser, _ = sessUser.(map[string]interface{})
				}
				if sessMapUser == nil {
					dam.Responder.Error(record.AUTHENTICATION_FAILED, username, w, "")
					return
				}
				if val, exists := sessMapUser[usernameColumnName]; !exists || val != username {
					dam.Responder.Error(record.AUTHENTICATION_FAILED, username, w, "")
					return
				}
				userColumns := columnNames
				found := false
				for _, col := range columnNames {
					if col == pkName {
						found = true
					}
				}
				if !found {
					userColumns = append(userColumns, pkName)
				}
				users := dam.db.SelectAll(table, userColumns, condition, columnOrdering, 0, 1)
				for _, user := range users {
					if err := bcrypt.CompareHashAndPassword([]byte(fmt.Sprint(user[passwordColumnName])), []byte(password)); err == nil {
						data := map[string]interface{}{}
						hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
						if err == nil {
							data[passwordColumnName] = string(hash)
						} else {
							dam.Responder.Error(record.HTTP_MESSAGE_NOT_READABLE, "", w, "")
							return
						}
						if _, err := dam.db.UpdateSingle(nil, table, data, fmt.Sprint(user[pkName])); err != nil {
							dam.Responder.Error(record.INTERNAL_SERVER_ERROR, "", w, "")
							return
						}
						delete(user, passwordColumnName)
						if !found {
							delete(user, pkName)
						}
						session.Values["user"] = user
						if err := session.Save(r, w); err == nil {
							dam.Responder.Success(user, w)
							return
						}
					}
				}
				dam.Responder.Error(record.AUTHENTICATION_FAILED, username, w, "")
				return
			}
		}
		if method == http.MethodPost && path == "logout" {
			session := utils.GetSession(w, r)
			if user, exists := session.Values["user"]; exists {
				delete(session.Values, "user")
				session.Options.MaxAge = -1
				if err := session.Save(r, w); err == nil {
					dam.Responder.Success(user, w)
					return
				}
			}
			dam.Responder.Error(record.AUTHENTICATION_REQUIRED, "", w, "")
			return
		}
		if method == http.MethodGet && path == "me" {
			session := utils.GetSession(w, r)
			if user, exists := session.Values["user"]; exists {
				dam.Responder.Success(user, w)
				return
			}
			dam.Responder.Error(record.AUTHENTICATION_REQUIRED, "", w, "")
			return
		}
		session := utils.GetSession(w, r)
		if user, exists := session.Values["user"]; !exists || user == nil {
			if authenticationMode := dam.getProperty("mode", "required"); authenticationMode == "required" {
				dam.Responder.Error(record.AUTHENTICATION_REQUIRED, "", w, "")
				return
			}
		}
		next.ServeHTTP(w, r)

	})
}
