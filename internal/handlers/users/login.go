package users

import (
	"context"
	"encoding/json"
	ap "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"log"
	"net/http"
	"time"
)

func LoginHandler(usersClient ap.UserServiceClient) http.HandlerFunc {
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

		req := &ap.LoginRequest{
			Login:    body.Login,
			Password: body.Password,
		}

		resp, err := usersClient.Login(ctx, req)
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
