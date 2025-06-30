package handlers

import (
	"context"
	"encoding/json"
	dapi "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	"log"
	"messenger_frontend/internal/middleware"
	"net/http"
	"strconv"
	"time"
)

type DialogHandlerService struct {
	dialogServiceClient dapi.DialogServiceClient
}

func NewDialogHandlerService(client dapi.DialogServiceClient) *DialogHandlerService {
	return &DialogHandlerService{dialogServiceClient: client}
}

func (d *DialogHandlerService) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/dialog/create", d.CreateDialogHandler())
	mux.HandleFunc("/dialog/send", d.SendMessageHandler())
	mux.HandleFunc("/dialog/messages", d.GetDialogMessagesHandler())
	mux.HandleFunc("/dialog/user", d.GetUserDialogsHandler())
}

func (d *DialogHandlerService) CreateDialogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(middleware.UserIDKey)
		userIDStr, ok := userIDVal.(string)
		if !ok {
			http.Error(w, "user ID missing", http.StatusUnauthorized)
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusUnauthorized)
			return
		}

		type RequestBody struct {
			PeerID     int32  `json:"peer_id"`
			DialogName string `json:"dialog_name"`
		}

		var reqBody RequestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, `{"error":"неправильный формат запроса"}`, http.StatusBadRequest)
			return
		}

		grpcReq := &dapi.CreateDialogRequest{
			UserId:     int32(userID),
			PeerId:     reqBody.PeerID,
			DialogName: reqBody.DialogName,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		resp, err := d.dialogServiceClient.CreateDialog(ctx, grpcReq)
		if err != nil {
			log.Printf("CreateDialog error: %v", err)
			http.Error(w, `{"error":"ошибка при создании диалога"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"dialog_id":   resp.DialogId,
			"dialog_name": resp.DialogName,
			"success":     resp.Success,
		})
	}
}

func (d *DialogHandlerService) SendMessageHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(middleware.UserIDKey)
		userIDStr, ok := userIDVal.(string)
		if !ok {
			http.Error(w, "user ID missing", http.StatusUnauthorized)
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusUnauthorized)
			return
		}

		type RequestBody struct {
			DialogID int32  `json:"dialog_id"`
			Text     string `json:"text"`
		}

		var reqBody RequestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, `{"error":"неправильный формат запроса"}`, http.StatusBadRequest)
			return
		}

		if reqBody.DialogID == 0 || reqBody.Text == "" {
			http.Error(w, `{"error":"dialog_id и text обязательны"}`, http.StatusBadRequest)
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		grpcReq := &dapi.SendMessageRequest{
			DialogId: reqBody.DialogID,
			UserId:   int32(userID),
			Text:     reqBody.Text,
		}
		resp, err := d.dialogServiceClient.SendMessage(ctx, grpcReq)
		if err != nil {
			log.Printf("SendMessage error: %v", err)
			http.Error(w, `{"error":"не удалось отправить сообщение"}`, http.StatusInternalServerError)
			return
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"message_id": resp.MessageId,
			"timestamp":  resp.Timestamp.AsTime().Format(time.RFC3339),
		})
	}
}

func (d *DialogHandlerService) GetUserDialogsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userIDVal := r.Context().Value(middleware.UserIDKey)
		userIDStr, ok := userIDVal.(string)
		if !ok {
			http.Error(w, "user ID missing", http.StatusUnauthorized)
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			http.Error(w, "invalid user ID", http.StatusUnauthorized)
			return
		}

		query := r.URL.Query()
		var limitPtr, offsetPtr *int32

		if limitStr := query.Get("limit"); limitStr != "" {
			limitVal, err := strconv.Atoi(limitStr)
			if err == nil && limitVal > 0 {
				val := int32(limitVal)
				limitPtr = &val
			}
		}
		if offsetStr := query.Get("offset"); offsetStr != "" {
			offsetVal, err := strconv.Atoi(offsetStr)
			if err == nil && offsetVal >= 0 {
				val := int32(offsetVal)
				offsetPtr = &val
			}
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		grpcReq := &dapi.GetUserDialogsRequest{
			UserId: int32(userID),
			Limit:  limitPtr,
			Offset: offsetPtr,
		}
		resp, err := d.dialogServiceClient.GetUserDialogs(ctx, grpcReq)
		if err != nil {
			log.Printf("GetUserDialogs error: %v", err)
			http.Error(w, `{"error":"не удалось получить список диалогов"}`, http.StatusInternalServerError)
			return
		}

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

func (d *DialogHandlerService) GetDialogMessagesHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		var limitPtr, offsetPtr *int32
		if limitStr := query.Get("limit"); limitStr != "" {
			limitVal, err := strconv.Atoi(limitStr)
			if err == nil && limitVal > 0 {
				val := int32(limitVal)
				limitPtr = &val
			}
		}
		if offsetStr := query.Get("offset"); offsetStr != "" {
			offsetVal, err := strconv.Atoi(offsetStr)
			if err == nil && offsetVal >= 0 {
				val := int32(offsetVal)
				offsetPtr = &val
			}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		grpcReq := &dapi.GetDialogMessagesRequest{
			DialogId: int32(dialogID),
			Limit:    limitPtr,
			Offset:   offsetPtr,
		}
		resp, err := d.dialogServiceClient.GetDialogMessages(ctx, grpcReq)
		if err != nil {
			http.Error(w, `{"error":"не удалось получить сообщения"}`, http.StatusInternalServerError)
			return
		}
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
