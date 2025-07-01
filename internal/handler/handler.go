package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"stroy-svaya/internal/model"
	"stroy-svaya/internal/service"
	"time"
)

type Handler struct {
	srv *service.Service
}

func NewHandler(s *service.Service) *Handler {
	return &Handler{srv: s}
}

func (h *Handler) InsertPdrLine(w http.ResponseWriter, r *http.Request) {
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

func (h *Handler) InsertOrUpdatePdrLine(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var rec model.PileDrivingRecordLine
	if err := json.NewDecoder(r.Body).Decode(&rec); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := h.srv.InsertOrUpdatePdrPile(&rec); err != nil {
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

func (h *Handler) GetPiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	filter, err := h.decodeUrlQueryToFilter(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	piles, err := h.srv.GetPiles(*filter)
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

func (h *Handler) GetPile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	query := r.URL.Query()
	filter, err := h.decodeUrlQueryToFilter(&query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	pile, err := h.srv.GetPile(*filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(pile); err != nil {
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

func (h *Handler) decodeUrlQueryToFilter(val *url.Values) (*model.PileFilter, error) {
	filter := model.PileFilter{}
	var err error
	for key, values := range *val {
		if len(values) == 0 {
			continue
		}
		value := values[0]
		switch key {
		case "project_id":
			filter.ProjectId, err = strconv.Atoi(value)
		case "pile_number":
			filter.PileNumber = &value
		case "pile_field_id":
			filter.PileFieldId, err = strconv.Atoi(value)
		case "start_date":
			dd, err0 := time.Parse(time.RFC3339, value)
			err = err0
			if err == nil {
				filter.StartDate = &dd
			}
		case "fact_pile_head":
			num, err0 := strconv.Atoi(value)
			err = err0
			if err == nil {
				filter.FactPileHead = &num
			}
		case "recorded_by":
			filter.RecordedBy = &value
		case "status":
			filter.Status, err = strconv.Atoi(value)
		default:
			return nil, fmt.Errorf("unknown url query key %s", key)
		}
		if err != nil {
			return nil, fmt.Errorf("decode PileFilter.%s: %s", key, err)
		}
	}
	return &filter, nil
}
