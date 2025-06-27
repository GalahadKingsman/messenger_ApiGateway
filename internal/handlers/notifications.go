package handlers

import (
	"io"
	"net/http"
)

type NotificationHandler struct {
	TargetURL string
}

func NewNotificationHandler(target string) *NotificationHandler {
	return &NotificationHandler{TargetURL: target}
}

func (h *NotificationHandler) RegisterHandlers(mux *http.ServeMux) {
	mux.HandleFunc("/notifications/", h.proxy)
}

func (h *NotificationHandler) proxy(w http.ResponseWriter, r *http.Request) {
	client := &http.Client{}
	url := h.TargetURL + r.URL.Path
	req, err := http.NewRequest(r.Method, url, r.Body)
	if err != nil {
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}

	req.Header = r.Header

	resp, err := client.Do(req)
	if err != nil {
		http.Error(w, "notifications service unavailable", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	for k, v := range resp.Header {
		w.Header()[k] = v
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
