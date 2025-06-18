package client

import (
	"log"

	pb "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api" // убедись, что путь корректный
	"google.golang.org/grpc"
)

var UsersClient pb.UserServiceClient

func InitClients() {
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("не удалось подключиться к users-сервису: %v", err)
	}
	UsersClient = pb.NewUserServiceClient(conn)
}
