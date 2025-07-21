package handlers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/GalahadKingsman/messenger_users/pkg/messenger_users_api"
	"github.com/go-redis/redismock/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"messenger_frontend/internal/handlers"
	"net/http"
	"net/http/httptest"
	"testing"
)

type mockUserServiceClient struct {
	mock.Mock
}

func (m *mockUserServiceClient) GetUser(ctx context.Context, req *messenger_users_api.GetUserRequest, _ ...grpc.CallOption) (*messenger_users_api.GetUserResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*messenger_users_api.GetUserResponse), args.Error(1)
}

func (m *mockUserServiceClient) Login(ctx context.Context, req *messenger_users_api.LoginRequest, _ ...grpc.CallOption) (*messenger_users_api.LoginResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*messenger_users_api.LoginResponse), args.Error(1)
}

func (m *mockUserServiceClient) CreateUser(ctx context.Context, req *messenger_users_api.CreateRequest, _ ...grpc.CallOption) (*messenger_users_api.CreateResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*messenger_users_api.CreateResponse), args.Error(1)
}

func ptr[T any](v T) *T { return &v }

// --------------------------- ТЕСТЫ ----------------------------

func TestCreateUserHandler_Success(t *testing.T) {
	mockClient := new(mockUserServiceClient)
	mockRedis, _ := redismock.NewClientMock()
	handler := handlers.NewUserHandlerService(mockClient, mockRedis)

	createReq := &messenger_users_api.CreateRequest{
		Login:     "user1",
		Password:  "pass123",
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john@example.com",
		Phone:     "1234567890",
	}
	mockClient.On("CreateUser", mock.Anything, createReq).
		Return(&messenger_users_api.CreateResponse{Success: "99"}, nil)

	body, _ := json.Marshal(createReq)
	r := httptest.NewRequest(http.MethodPost, "/users/create", bytes.NewBuffer(body))
	w := httptest.NewRecorder()

	handler.CreateUserHandler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"success":"99"`)
}

func TestLoginHandler_Success(t *testing.T) {
	mockClient := new(mockUserServiceClient)
	mockRedis, redisMock := redismock.NewClientMock()
	handler := handlers.NewUserHandlerService(mockClient, mockRedis)

	reqBody := map[string]string{"login": "user1", "password": "pass123"}
	body, _ := json.Marshal(reqBody)

	mockClient.On("Login", mock.Anything, &messenger_users_api.LoginRequest{
		Login:    "user1",
		Password: "pass123",
	}).Return(&messenger_users_api.LoginResponse{
		UserId:  42,
		Token:   "token123",
		Message: "OK",
	}, nil)

	redisMock.ExpectSet("token:42", "token123", 0).SetVal("OK")

	r := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	handler.LoginHandler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "token123")
	assert.Contains(t, w.Body.String(), `"user_id":42`)
}

func TestGetUserHandler_ByLogin_Success(t *testing.T) {
	mockClient := new(mockUserServiceClient)
	handler := handlers.NewUserHandlerService(mockClient, nil)

	mockClient.On("GetUser", mock.Anything, &messenger_users_api.GetUserRequest{
		Login: ptr("user1"),
	}).Return(&messenger_users_api.GetUserResponse{
		Users: []*messenger_users_api.GetUserResponse_User{
			{
				Id:        1,
				Login:     "user1",
				FirstName: "John",
				LastName:  "Doe",
				Email:     "john@example.com",
				Phone:     "1234567890",
			},
		},
	}, nil)

	r := httptest.NewRequest(http.MethodGet, "/users/get?login=user1", nil)
	w := httptest.NewRecorder()
	handler.GetUserHandler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), `"login":"user1"`)
}

func TestGetUserHandler_NoParams(t *testing.T) {
	handler := handlers.NewUserHandlerService(nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/users/get", nil)
	w := httptest.NewRecorder()
	handler.GetUserHandler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "нужно указать хотя бы один параметр")
}

func TestLoginHandler_InvalidJSON(t *testing.T) {
	handler := handlers.NewUserHandlerService(nil, nil)

	r := httptest.NewRequest(http.MethodPost, "/users/login", bytes.NewBufferString("{invalid"))
	w := httptest.NewRecorder()
	handler.LoginHandler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "не удалось разобрать тело запроса")
}

func TestCreateUserHandler_InvalidMethod(t *testing.T) {
	handler := handlers.NewUserHandlerService(nil, nil)

	r := httptest.NewRequest(http.MethodGet, "/users/create", nil)
	w := httptest.NewRecorder()
	handler.CreateUserHandler().ServeHTTP(w, r)

	assert.Equal(t, http.StatusMethodNotAllowed, w.Code)
}
