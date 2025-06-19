package dialog

import (
	"context"
	"encoding/json"
	api "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	"log"
	"net/http"
	"strconv"
	"time"
)

func GetUserDialogsHandler(dialogsClient api.DialogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()

		// user_id обязателен
		userIDStr := query.Get("user_id")
		if userIDStr == "" {
			http.Error(w, `{"error":"user_id обязателен"}`, http.StatusBadRequest)
			return
		}

		userID, err := strconv.Atoi(userIDStr)
		if err != nil || userID <= 0 {
			http.Error(w, `{"error":"user_id должен быть положительным числом"}`, http.StatusBadRequest)
			return
		}

		// limit — опциональный
		var limitPtr *int32
		if limitStr := query.Get("limit"); limitStr != "" {
			limitVal, err := strconv.Atoi(limitStr)
			if err != nil || limitVal < 0 {
				http.Error(w, `{"error":"limit должен быть положительным числом"}`, http.StatusBadRequest)
				return
			}
			limit := int32(limitVal)
			limitPtr = &limit
		}

		// offset — опциональный
		var offsetPtr *int32
		if offsetStr := query.Get("offset"); offsetStr != "" {
			offsetVal, err := strconv.Atoi(offsetStr)
			if err != nil || offsetVal < 0 {
				http.Error(w, `{"error":"offset должен быть положительным числом"}`, http.StatusBadRequest)
				return
			}
			offset := int32(offsetVal)
			offsetPtr = &offset
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		grpcReq := &api.GetUserDialogsRequest{
			UserId: int32(userID),
			Limit:  limitPtr,
			Offset: offsetPtr,
		}

		resp, err := dialogsClient.GetUserDialogs(ctx, grpcReq)
		if err != nil {
			log.Printf("GetUserDialogs error: %v", err)
			http.Error(w, `{"error":"не удалось получить список диалогов"}`, http.StatusInternalServerError)
			return
		}

		// Формируем JSON-ответ
		dialogs := make([]map[string]interface{}, 0, len(resp.Dialogs))
		for _, d := range resp.Dialogs {
			dialogs = append(dialogs, map[string]interface{}{
				"dialog_id":    d.DialogId,
				"peer_id":      d.PeerId,
				"peer_login":   d.PeerLogin,
				"last_message": d.LastMessage,
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"dialogs": dialogs,
		})
	}
}
