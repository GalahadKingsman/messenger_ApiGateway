package grpc_clients

import (
	api "github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	ap "github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
)

type Service struct {
	dialogServiceClient api.DialogServiceClient
	userServiceClient   ap.UserServiceClient
}
