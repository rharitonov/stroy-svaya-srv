package webservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"stroy-svaya/internal/model"
)

type WebService struct {
	BaseUrl string
}

func NewWebService(baseUrl string) *WebService {
	if baseUrl == "" {
		baseUrl = "http://localhost:8080"
	}
	return &WebService{BaseUrl: baseUrl}
}

func (w *WebService) GetPiles(filter model.PileFilter, mode string) ([]string, error) {
	projectId := 0
	url := fmt.Sprintf("%s/getpilestodriving?project_id=%d", w.BaseUrl, projectId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var piles []string
	if err := json.Unmarshal(body, &piles); err != nil {
		return nil, err
	}
	return piles, err
}

func (w *WebService) SendExcel(projectId int) error {
	url := fmt.Sprintf("%s/sendpdrlog?project_id=%d", w.BaseUrl, projectId)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (w *WebService) GetUserFullName(chatId int64, defaultUserName string) (string, error) {
	result := ""
	url := fmt.Sprintf("%s/getuserfullname?tg_chat_id=%d", w.BaseUrl, chatId)
	resp, err := http.Get(url)
	if err != nil {
		return result, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return result, err
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}
	if result == "" {
		result = defaultUserName
	}
	return result, nil
}

func (w *WebService) SendData(rec *model.PileDrivingRecordLine) error {
	jsonData, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/insertpdrline", w.BaseUrl)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		return errors.New(resp.Status)
	}
	return nil
}
