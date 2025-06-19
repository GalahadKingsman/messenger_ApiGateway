package users

import (
	"encoding/json"
	ap "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"io"
	"net/http"
)

func CreateUserHandler(usersClient ap.UserServiceClient) http.HandlerFunc {
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

		var req ap.CreateRequest
		if err := json.Unmarshal(body, &req); err != nil {
			http.Error(w, "Некорректный JSON", http.StatusBadRequest)
			return
		}

		// Вызов gRPC-метода
		resp, err := usersClient.CreateUser(r.Context(), &req)
		if err != nil {
			http.Error(w, "Ошибка при создании пользователя: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// Отправка успешного ответа
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}
