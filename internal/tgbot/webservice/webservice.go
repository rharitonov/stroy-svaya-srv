package webservice

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

func (w *WebService) GetPiles(filter model.PileFilter) ([]string, error) {
	values, err := w.getUrlValues(filter)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/getpiles?%s", w.BaseUrl, values)
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
	return piles, nil
}

func (w *WebService) GetPile(filter model.PileFilter) (*model.PileDrivingRecordLine, error) {
	values, err := w.getUrlValues(filter)
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("%s/getpile?%s", w.BaseUrl, values)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var pile model.PileDrivingRecordLine
	if err := json.Unmarshal(body, &pile); err != nil {
		return nil, err
	}
	return &pile, nil
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

func (w *WebService) GetUserSetup(chatId int64) (*model.User, error) {
	url := fmt.Sprintf("%s/getusersetup?tg_chat_id=%d", w.BaseUrl, chatId)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var user *model.User = new(model.User)
	if err := json.Unmarshal(body, user); err != nil {
		return nil, err
	}
	return user, nil
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

func (w *WebService) InsertOrUpdatePdrLine(rec *model.PileDrivingRecordLine) error {
	jsonData, err := json.Marshal(rec)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/insertorupdatepdrline", w.BaseUrl)
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

func (w *WebService) SendPdrLog(projectId int) error {
	if projectId == 0 {
		projectId = 1
	}
	url := fmt.Sprintf("%s/sendpdrlog?project_id=%d", w.BaseUrl, projectId)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func (w *WebService) getUrlValues(filter model.PileFilter) (string, error) {
	jsonStr, err := json.Marshal(filter)
	if err != nil {
		return "", err
	}
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return "", nil
	}
	values := url.Values{}
	for key, val := range data {
		switch v := val.(type) {
		case string:
			values.Add(key, v)
		case bool:
			values.Add(key, fmt.Sprintf("%t", v))
		default:
			values.Add(key, fmt.Sprintf("%v", v))
		}
	}
	return values.Encode(), nil
}
