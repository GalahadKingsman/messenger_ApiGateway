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

func GetDialogMessagesHandler(dialogsClient api.DialogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		query := r.URL.Query()

		dialogIDStr := query.Get("dialog_id")
		if dialogIDStr == "" {
			http.Error(w, `{"error":"dialog_id обязателен"}`, http.StatusBadRequest)
			return
		}

		dialogID, err := strconv.Atoi(dialogIDStr)
		if err != nil {
			http.Error(w, `{"error":"dialog_id должен быть числом"}`, http.StatusBadRequest)
			return
		}

		// limit и offset — опциональные
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

		grpcReq := &api.GetDialogMessagesRequest{
			DialogId: int32(dialogID),
			Limit:    limitPtr,
			Offset:   offsetPtr,
		}

		resp, err := dialogsClient.GetDialogMessages(ctx, grpcReq)
		if err != nil {
			log.Printf("GetDialogMessages error: %v", err)
			http.Error(w, `{"error":"не удалось получить сообщения"}`, http.StatusInternalServerError)
			return
		}

		// Формируем JSON-ответ
		messages := make([]map[string]interface{}, 0, len(resp.Messages))
		for _, msg := range resp.Messages {
			messages = append(messages, map[string]interface{}{
				"id":        msg.Id,
				"user_id":   msg.UserId,
				"text":      msg.Text,
				"timestamp": msg.Timestamp.AsTime().Format(time.RFC3339),
			})
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"messages": messages,
		})
	}
}
