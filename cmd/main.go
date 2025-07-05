package main

import (
	"context"
	dapi "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	uapi "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"messenger_frontend/internal/handlers"
	"messenger_frontend/internal/middleware"
	"messenger_frontend/internal/storage"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()

	storage.InitRedis()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Подключение к dialog-сервису
	dialogsConn, err := grpc.DialContext(ctx, "dialog_service:9001", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к dialogs gRPC: %v", err)
	}
	defer dialogsConn.Close()
	dialogsClient := dapi.NewDialogServiceClient(dialogsConn)

	// Подключение к users-сервису
	usersConn, err := grpc.DialContext(ctx, "users_service:9000", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к users gRPC: %v", err)
	}
	defer usersConn.Close()
	usersClient := uapi.NewUserServiceClient(usersConn)

	mux := http.NewServeMux()

	dialogHandler := handlers.NewDialogHandlerService(dialogsClient)
	dialogHandler.RegisterHandlers(mux)

	userHandler := handlers.NewUserHandlerService(usersClient, storage.Rdb)
	userHandler.RegisterHandlers(mux)

	notificationHandler := handlers.NewNotificationHandler("http://notifications:8082/notifications")
	notificationHandler.RegisterHandlers(mux)

	protectedMux := middleware.JWTAuthMiddleware(mux)

	// Запуск HTTP-сервера
	srv := &http.Server{
		Addr:         ":8080",
		Handler:      protectedMux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 35 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("HTTP сервер запущен на :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ошибка запуска HTTP-сервера: %v", err)
	}
}
