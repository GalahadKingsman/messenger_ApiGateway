package users

import (
	"context"
	"encoding/json"
	ap "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetUserHandler(usersClient ap.UserServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()

		var (
			idPtr        *int64
			loginPtr     *string
			firstNamePtr *string
			lastNamePtr  *string
			emailPtr     *string
			phonePtr     *string
		)

		if idStr := query.Get("id"); idStr != "" {
			idVal, err := strconv.ParseInt(idStr, 10, 64)
			if err != nil || idVal <= 0 {
				http.Error(w, `{"error":"id должен быть положительным числом"}`, http.StatusBadRequest)
				return
			}
			idPtr = &idVal
		}
		if login := query.Get("login"); login != "" {
			loginPtr = &login
		}
		if first := query.Get("first_name"); first != "" {
			firstNamePtr = &first
		}
		if last := query.Get("last_name"); last != "" {
			lastNamePtr = &last
		}
		if email := query.Get("email"); email != "" {
			emailPtr = &email
		}
		if phone := query.Get("phone"); phone != "" {
			phonePtr = &phone
		}

		// Проверка: должен быть хотя бы один параметр
		if idPtr == nil && loginPtr == nil && firstNamePtr == nil &&
			lastNamePtr == nil && emailPtr == nil && phonePtr == nil {
			http.Error(w, `{"error":"нужно указать хотя бы один параметр"}`, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := &ap.GetUserRequest{
			Id:        idPtr,
			Login:     loginPtr,
			FirstName: firstNamePtr,
			LastName:  lastNamePtr,
			Email:     emailPtr,
			Phone:     phonePtr,
		}

		resp, err := usersClient.GetUser(ctx, req)
		if err != nil {
			log.Printf("GetUser error: %v", err)
			http.Error(w, `{"error":"не удалось получить пользователя"}`, http.StatusInternalServerError)
			return
		}

		// Формируем JSON-ответ
		users := make([]map[string]interface{}, 0, len(resp.Users))
		for _, u := range resp.Users {
			users = append(users, map[string]interface{}{
				"id":         u.Id,
				"login":      u.Login,
				"first_name": u.FirstName,
				"last_name":  u.LastName,
				"email":      u.Email,
				"phone":      u.Phone,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"users": users,
		})
	}
}
