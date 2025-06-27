package main

import (
	"context"
	dapi "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	uapi "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"messenger_frontend/internal/handlers"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Подключение к dialog-сервису
	dialogsConn, err := grpc.DialContext(ctx, "localhost:9001", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к dialogs gRPC: %v", err)
	}
	defer dialogsConn.Close()
	dialogsClient := dapi.NewDialogServiceClient(dialogsConn)

	// Подключение к users-сервису
	usersConn, err := grpc.DialContext(ctx, "localhost:9000", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к users gRPC: %v", err)
	}
	defer usersConn.Close()
	usersClient := uapi.NewUserServiceClient(usersConn)

	mux := http.NewServeMux()

	dialogHandler := handlers.NewDialogHandlerService(dialogsClient)
	dialogHandler.RegisterHandlers(mux)

	userHandler := handlers.NewUserHandlerService(usersClient)
	userHandler.RegisterHandlers(mux)

	notificationHandler := handlers.NewNotificationHandler("http://notifications:8082")
	notificationHandler.RegisterHandlers(mux)

	// Запуск HTTP-сервера
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("HTTP сервер запущен на :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ошибка запуска HTTP-сервера: %v", err)
	}
}
