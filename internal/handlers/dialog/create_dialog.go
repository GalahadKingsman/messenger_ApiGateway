package dialog

import (
	"context"
	"encoding/json"
	api "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	"log"
	"net/http"
	"time"
)

func CreateDialogHandler(dialogsClient api.DialogServiceClient) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		type RequestBody struct {
			UserID     int32  `json:"user_id"`
			PeerID     int32  `json:"peer_id"`
			DialogName string `json:"dialog_name"` // Добавим поддержку имени
		}

		w.Header().Set("Content-Type", "application/json")

		var reqBody RequestBody
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			http.Error(w, `{"error":"неправильный формат запроса"}`, http.StatusBadRequest)
			return
		}

		grpcReq := &api.CreateDialogRequest{
			UserId:     reqBody.UserID,
			PeerId:     reqBody.PeerID,
			DialogName: reqBody.DialogName,
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		resp, err := dialogsClient.CreateDialog(ctx, grpcReq)
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
