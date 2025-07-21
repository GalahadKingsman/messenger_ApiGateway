package handlers

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"messenger_frontend/internal/middleware"
)

func TestNotificationHandler_proxy_Success(t *testing.T) {
	// Создаём тестовый upstream-сервер, который эмулирует микросервис уведомлений
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем, что userID добавился в query
		userID := r.URL.Query().Get("userID")
		if userID != "12345" {
			t.Errorf("expected userID to be 12345, got %s", userID)
		}
		w.Header().Set("X-Test", "ok")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("proxied response"))
	}))
	defer upstream.Close()

	handler := NewNotificationHandler(upstream.URL)

	// Создаём проксируемый запрос
	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "12345"))
	w := httptest.NewRecorder()

	// Выполняем handler.proxy("")
	handlerFunc := handler.RegisterHandlersAndGet("/notifications")
	handlerFunc.ServeHTTP(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status 200 OK, got %d", resp.StatusCode)
	}
	if string(body) != "proxied response" {
		t.Errorf("unexpected body: %s", string(body))
	}
	if resp.Header.Get("X-Test") != "ok" {
		t.Errorf("missing expected header X-Test")
	}
}

func TestNotificationHandler_proxy_Unauthorized(t *testing.T) {
	handler := NewNotificationHandler("http://example.com")

	req := httptest.NewRequest(http.MethodGet, "/notifications", nil)
	// Контекст без userID
	w := httptest.NewRecorder()

	handlerFunc := handler.RegisterHandlersAndGet("/notifications")
	handlerFunc.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", resp.StatusCode)
	}
}

func TestNotificationHandler_proxy_UpstreamError(t *testing.T) {
	// Некорректный адрес upstream
	handler := NewNotificationHandler("http://invalidhost")

	req := httptest.NewRequest(http.MethodGet, "/notifications", strings.NewReader(""))
	req = req.WithContext(context.WithValue(req.Context(), middleware.UserIDKey, "12345"))
	w := httptest.NewRecorder()

	handlerFunc := handler.RegisterHandlersAndGet("/notifications")
	handlerFunc.ServeHTTP(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadGateway {
		t.Errorf("expected status 502 Bad Gateway, got %d", resp.StatusCode)
	}
}
