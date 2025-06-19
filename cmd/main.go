package main

import (
	"context"
	api "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	ap "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"messenger_frontend/internal/handlers/dialog"
	"messenger_frontend/internal/handlers/users"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()

	// Общие gRPC опции
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	dialogsConn, err := grpc.DialContext(ctx, "localhost:9001", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к dialogs gRPC: %v", err)
	}
	defer dialogsConn.Close()
	dialogsClient := api.NewDialogServiceClient(dialogsConn)

	usersConn, err := grpc.DialContext(ctx, "localhost:9000", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к users gRPC: %v", err)
	}
	defer usersConn.Close()
	usersClient := ap.NewUserServiceClient(usersConn)

	http.HandleFunc("/dialog/create", dialog.CreateDialogHandler(dialogsClient))
	http.HandleFunc("/dialog/messages", dialog.GetDialogMessagesHandler(dialogsClient))
	http.HandleFunc("/dialog/send", dialog.SendMessageHandler(dialogsClient))
	http.HandleFunc("/dialogs/user", dialog.GetUserDialogsHandler(dialogsClient))

	http.HandleFunc("/users/create", users.CreateUserHandler(usersClient))
	http.HandleFunc("/users/get", users.GetUserHandler(usersClient))
	http.HandleFunc("/users/login", users.LoginHandler(usersClient))

	// Запуск HTTP-сервера
	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("HTTP сервер запущен на :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ошибка запуска HTTP-сервера: %v", err)
	}
}
