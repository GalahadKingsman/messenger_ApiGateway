package handlers

import (
	"context"
	"encoding/json"
	uapi "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

type UserHandlerService struct {
	UserServiceClient uapi.UserServiceClient
}

func NewUserHandlerService(client uapi.UserServiceClient) *UserHandlerService {
	return &UserHandlerService{UserServiceClient: client}
}

func (u *UserHandlerService) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/users/create", u.CreateUserHandler())
	mux.HandleFunc("/users/get", u.GetUserHandler())
	mux.HandleFunc("/users/login", u.LoginHandler())
}

func (u *UserHandlerService) GetUserHandler() http.HandlerFunc {
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

		req := &uapi.GetUserRequest{
			Id:        idPtr,
			Login:     loginPtr,
			FirstName: firstNamePtr,
			LastName:  lastNamePtr,
			Email:     emailPtr,
			Phone:     phonePtr,
		}

		resp, err := u.UserServiceClient.GetUser(ctx, req)
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

func (u *UserHandlerService) LoginHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Только POST разрешён
		if r.Method != http.MethodPost {
			http.Error(w, `{"error":"только POST запрос разрешен"}`, http.StatusMethodNotAllowed)
			return
		}

		var body struct {
			Login    string `json:"login"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			http.Error(w, `{"error":"не удалось разобрать тело запроса"}`, http.StatusBadRequest)
			return
		}

		if body.Login == "" || body.Password == "" {
			http.Error(w, `{"error":"login и password обязательны"}`, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		req := &uapi.LoginRequest{
			Login:    body.Login,
			Password: body.Password,
		}

		resp, err := u.UserServiceClient.Login(ctx, req)
		if err != nil {
			log.Printf("Login error: %v", err)
			http.Error(w, `{"error":"ошибка сервера при входе"}`, http.StatusInternalServerError)
			return
		}

		// Формируем JSON-ответ
		response := map[string]interface{}{
			"message": resp.Message,
		}
		if resp.UserId != 0 {
			response["user_id"] = resp.UserId
		}

		json.NewEncoder(w).Encode(response)
	}
}

func (u *UserHandlerService) CreateUserHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Метод не поддерживается", http.StatusMethodNotAllowed)
			return
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Не удалось прочитать тело запроса", http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		var req uapi.CreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}

		// Вызов gRPC-метода
		resp, err := u.UserServiceClient.CreateUser(r.Context(), &req)
		if err != nil {
			http.Error(w, "Ошибка при создании пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправка успешного ответа
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
