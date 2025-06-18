package handlers

import (
	"encoding/json"
	"github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"golang.org/x/net/context"
	"messenger_frontend/internal/grpc_clients/users"
	"net/http"
)

type CreateUserRequest struct {
	Login     string `json:"login"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
}

func CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "неправильный JSON", http.StatusBadRequest)
		return
	}

	grpcReq := &messenger_users_api.CreateRequest{
		Login:     req.Login,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
		Phone:     req.Phone,
		Password:  req.Password,
	}

	resp, err := client.UsersClient.(context.Background(), grpcReq)
	if err != nil {
		http.Error(w, "ошибка gRPC: "+err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": resp.Success,
	})
}
