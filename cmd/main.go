package main

import (
	"context"
	api "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"log"
	"messenger_frontend/internal/handlers/dialog"
	"net/http"
	"time"
)

func main() {
	ctx := context.Background()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	// Подключение к gRPC-сервису
	conn, err := grpc.DialContext(ctx, "localhost:9001", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к gRPC 9001: %v", err)
	}
	connn, err := grpc.DialContext(ctx, "localhost:9000", opts...)
	if err != nil {
		log.Fatalf("не удалось подключиться к gRPC 9000: %v", err)
	}

	defer conn.Close()

	// Создание gRPC-клиента
	dialogsClient := api.NewDialogServiceClient(conn)

	// Настройка HTTP-обработчиков
	http.HandleFunc("/dialog/create", dialog(dialogsClient))

	// Запуск HTTP-сервера
	srv := &http.Server{
		Addr:         ":8080",
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Println("HTTP сервер запущен на :8080")
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("ошибка запуска сервера: %v", err)
	}
}
