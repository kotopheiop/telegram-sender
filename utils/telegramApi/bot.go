package telegramApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

type Bot struct {
	ID        int    `json:"id"`
	IsBot     bool   `json:"is_bot"`
	FirstName string `json:"first_name"`
	Username  string `json:"username"`
	token     string
	client    *http.Client //будем переиспользовать клиент чтобы не создавать новый
}

type SendMessageParams struct {
	ChatID          int64  `json:"chat_id"`
	MessageThreadID *int   `json:"message_thread_id,omitempty"`
	Text            string `json:"text"`
}

func NewBot(botToken string) (*Bot, error) {

	client := &http.Client{
		Timeout: time.Second * 15,
	}

	resp, err := http.Get("https://api.telegram.org/bot" + botToken + "/getMe")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result struct {
		Ok     bool `json:"ok"`
		Result Bot  `json:"result"`
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}

	if !result.Ok {
		return nil, fmt.Errorf("Не удалось получить информацию о боте")
	}

	result.Result.token = botToken
	result.Result.client = client

	return &result.Result, nil
}

func (b *Bot) SendMessage(params SendMessageParams) error {
	body, err := json.Marshal(params)
	if err != nil {
		return err
	}

	resp, err := b.client.Post("https://api.telegram.org/bot"+b.token+"/sendMessage", "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	var result struct {
		Ok          bool   `json:"ok"`
		Description string `json:"description"`
	}
	err = json.Unmarshal(respBody, &result)
	if err != nil {
		return err
	}

	if !result.Ok {
		return fmt.Errorf("Не удалось отправить сообщение: %s", result.Description)
	}

	return nil
}
