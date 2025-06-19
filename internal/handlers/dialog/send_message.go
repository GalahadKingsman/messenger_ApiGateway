package dialog

import (
	"context"
	"encoding/json"
	api "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	"log"
	"net/http"
	"time"
)

func SendMessageHandler(dialogsClient api.DialogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type RequestBody struct {
			DialogID int32  `json:"dialog_id"`
			UserID   int32  `json:"user_id"`
			Text     string `json:"text"`
		}

		w.Header().Set("Content-Type", "application/json")

		// Декодируем JSON тело
		var reqBody RequestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, `{"error":"неправильный формат запроса"}`, http.StatusBadRequest)
			return
		}

		// Простая валидация
		if reqBody.DialogID == 0 || reqBody.UserID == 0 || reqBody.Text == "" {
			http.Error(w, `{"error":"dialog_id, user_id и text обязательны"}`, http.StatusBadRequest)
			return
		}

		// gRPC-запрос
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcReq := &api.SendMessageRequest{
			DialogId: reqBody.DialogID,
			UserId:   reqBody.UserID,
			Text:     reqBody.Text,
		}

		resp, err := dialogsClient.SendMessage(ctx, grpcReq)
		if err != nil {
			log.Printf("SendMessage error: %v", err)
			http.Error(w, `{"error":"не удалось отправить сообщение"}`, http.StatusInternalServerError)
			return
		}

		// Формируем и возвращаем JSON-ответ
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message_id": resp.MessageId,
			"timestamp":  resp.Timestamp.AsTime().Format(time.RFC3339),
		})
	}
}
