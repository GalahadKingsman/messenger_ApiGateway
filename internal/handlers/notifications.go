package handlers

import (
	"io"
	"log"
	"messenger_frontend/internal/middleware"
	"net/http"
	"net/url"
)

type NotificationHandler struct {
	BaseURL string
}

type Notification struct {
	From    string `json:"from"`
	Message string `json:"message"`
}

func NewNotificationHandler(baseURL string) *NotificationHandler {
	return &NotificationHandler{BaseURL: baseURL}
}

func (h *NotificationHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/notifications", h.proxy(""))
	mux.HandleFunc("/notifications/clear", h.proxy("/clear"))
	mux.HandleFunc("/notifications/longpoll", h.proxy("/longpoll"))
}

func (h *NotificationHandler) proxy(endpoint string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("[Gateway] proxy %s %s → %s%s?%s",
			r.Method, r.URL.Path,
			h.BaseURL, endpoint, r.URL.RawQuery,
		)
		userID := r.Context().Value(middleware.UserIDKey)
		userIDStr, ok := userID.(string)
		if !ok {
			http.Error(w, "user ID missing", http.StatusUnauthorized)
			return
		}

		proxyURL, _ := url.Parse(h.BaseURL + endpoint)
		query := r.URL.Query()
		query.Set("userID", userIDStr)
		proxyURL.RawQuery = query.Encode()

		proxyReq, _ := http.NewRequest(r.Method, proxyURL.String(), r.Body)
		proxyReq.Header = r.Header.Clone()

		resp, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
			http.Error(w, "proxy error", http.StatusBadGateway)
			return
		}
		defer resp.Body.Close()

		log.Printf("[Gateway] upstream returned %d %s", resp.StatusCode, resp.Status)

		for k, vv := range resp.Header {
			for _, v := range vv {
				w.Header().Add(k, v)
			}
		}
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
	}
}

func (h *NotificationHandler) RegisterHandlersAndGet(endpoint string) http.HandlerFunc {
	return h.proxy(endpoint)
}
