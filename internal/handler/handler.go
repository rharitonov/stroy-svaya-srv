package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"stroy-svaya/internal/model"
	"stroy-svaya/internal/service"
)

type Handler struct {
	srv *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{srv: s}
}

func (h *Handler) InsertPileDrivingRecordLine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var rec model.PileDrivingRecordLine
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := h.srv.InsertPileDrivingRecordLine(&rec); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func (h *Handler) GetPileDrivingRecord(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	projectIdTxt := query.Get("project_id")
	if projectIdTxt == "" {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}
	projectId, err := strconv.Atoi(projectIdTxt)
	if err != nil {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}
	lines, err := h.srv.GetPileDrivingRecord(projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(lines); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *Handler) SendPileDrivingRecordLog(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	projectIdTxt := query.Get("project_id")
	if projectIdTxt == "" {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}
	projectId, err := strconv.Atoi(projectIdTxt)
	if err != nil {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}
	if err := h.srv.SendPileDrivingRecordLog(projectId); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})

}

func (h *Handler) GetPilesToDriving(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	projectIdTxt := query.Get("project_id")
	if projectIdTxt == "" {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}
	projectId, err := strconv.Atoi(projectIdTxt)
	if err != nil {
		http.Error(w, "Missing project id", http.StatusBadRequest)
		return
	}
	piles, err := h.srv.GetPilesToDriving(projectId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(piles); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func (h *Handler) GetUserFullNameInitialFormat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	tgChatIdTxt := query.Get("tg_chat_id")
	if tgChatIdTxt == "" {
		http.Error(w, "Missing tg chat id", http.StatusBadRequest)
		return
	}
	tgChatId, err := strconv.ParseInt(tgChatIdTxt, 10, 64)
	if err != nil {
		http.Error(w, "Incorrect tg chat id value", http.StatusBadRequest)
		return
	}

	userName, err := h.srv.GetUserFullNameInitialFormat(tgChatId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(userName); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}
