package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/GalahadKingsman/messenger_dialog/pkg/messenger_dialog_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
	"messenger_frontend/internal/middleware"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockDialogServiceClient struct {
	mock.Mock
}

func (m *mockDialogServiceClient) CreateDialog(ctx context.Context, req *messenger_dialog_api.CreateDialogRequest, _ ...grpc.CallOption) (*messenger_dialog_api.CreateDialogResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*messenger_dialog_api.CreateDialogResponse), args.Error(1)
}

func (m *mockDialogServiceClient) SendMessage(ctx context.Context, req *messenger_dialog_api.SendMessageRequest, _ ...grpc.CallOption) (*messenger_dialog_api.SendMessageResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*messenger_dialog_api.SendMessageResponse), args.Error(1)
}

func (m *mockDialogServiceClient) GetUserDialogs(ctx context.Context, req *messenger_dialog_api.GetUserDialogsRequest, _ ...grpc.CallOption) (*messenger_dialog_api.GetUserDialogsResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*messenger_dialog_api.GetUserDialogsResponse), args.Error(1)
}

func (m *mockDialogServiceClient) GetDialogMessages(ctx context.Context, req *messenger_dialog_api.GetDialogMessagesRequest, _ ...grpc.CallOption) (*messenger_dialog_api.GetDialogMessagesResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*messenger_dialog_api.GetDialogMessagesResponse), args.Error(1)
}

func withUserContext(r *http.Request, userID string) *http.Request {
	ctx := context.WithValue(r.Context(), middleware.UserIDKey, userID)
	return r.WithContext(ctx)
}

func TestCreateDialogHandler_Success(t *testing.T) {
	mockClient := new(mockDialogServiceClient)
	handler := NewDialogHandlerService(mockClient)

	payload := map[string]interface{}{
		"peer_id":     2,
		"dialog_name": "TestDialog",
	}
	body, _ := json.Marshal(payload)

	mockClient.On("CreateDialog", mock.Anything, &messenger_dialog_api.CreateDialogRequest{
		UserId:     1,
		PeerId:     2,
		DialogName: "TestDialog",
	}).Return(&messenger_dialog_api.CreateDialogResponse{
		DialogId:   10,
		DialogName: "TestDialog",
		Success:    true,
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/dialog/create", bytes.NewBuffer(body))
	req = withUserContext(req, "1")
	w := httptest.NewRecorder()
	handler.CreateDialogHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "TestDialog")
}

func TestSendMessageHandler_Success(t *testing.T) {
	mockClient := new(mockDialogServiceClient)
	handler := NewDialogHandlerService(mockClient)

	payload := map[string]interface{}{
		"dialog_id": 10,
		"text":      "Hello world",
	}
	body, _ := json.Marshal(payload)

	mockClient.On("SendMessage", mock.Anything, &messenger_dialog_api.SendMessageRequest{
		DialogId: 10,
		UserId:   1,
		Text:     "Hello world",
	}).Return(&messenger_dialog_api.SendMessageResponse{
		MessageId: 99,
		Timestamp: timestamppb.Now(),
	}, nil)

	req := httptest.NewRequest(http.MethodPost, "/dialog/send", bytes.NewBuffer(body))
	req = withUserContext(req, "1")
	w := httptest.NewRecorder()
	handler.SendMessageHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "message_id")
}

func TestGetUserDialogsHandler_Success(t *testing.T) {
	mockClient := new(mockDialogServiceClient)
	handler := NewDialogHandlerService(mockClient)

	mockClient.On("GetUserDialogs", mock.Anything, &messenger_dialog_api.GetUserDialogsRequest{
		UserId: 1,
	}).Return(&messenger_dialog_api.GetUserDialogsResponse{
		Dialogs: []*messenger_dialog_api.DialogInfo{
			{
				DialogId:    1,
				PeerId:      2,
				PeerLogin:   "peeruser",
				LastMessage: "last text",
			},
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/dialog/user", nil)
	req = withUserContext(req, "1")
	w := httptest.NewRecorder()
	handler.GetUserDialogsHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "peeruser")
}

func TestGetDialogMessagesHandler_Success(t *testing.T) {
	mockClient := new(mockDialogServiceClient)
	handler := NewDialogHandlerService(mockClient)

	mockClient.On("GetDialogMessages", mock.Anything, &messenger_dialog_api.GetDialogMessagesRequest{
		DialogId: 10,
	}).Return(&messenger_dialog_api.GetDialogMessagesResponse{
		Messages: []*messenger_dialog_api.Message{
			{
				Id:        1,
				UserId:    2,
				Text:      "hi!",
				Timestamp: timestamppb.Now(),
			},
		},
	}, nil)

	req := httptest.NewRequest(http.MethodGet, "/dialog/messages?dialog_id=10", nil)
	w := httptest.NewRecorder()
	handler.GetDialogMessagesHandler().ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "hi!")
}
