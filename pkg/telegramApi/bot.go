package telegramApi

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

type TelegramError struct {
	Code        int
	Description string
}

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

const telegramAPI = "https://api.telegram.org/bot"

func InitBot(botToken string) (*Bot, error) {
	client := &http.Client{
		Timeout: time.Second * 15,
	}

	resp, err := http.Get(telegramAPI + botToken + "/getMe")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, &TelegramError{
			Code:        resp.StatusCode,
			Description: "не удалось получить информацию о боте",
		}
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать тело ответа: %w", err)
	}

	var result struct {
		Ok     bool `json:"ok"`
		Result Bot  `json:"result"`
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, fmt.Errorf("не удалось разобрать JSON: %w", err)
	}

	if !result.Ok {
		return nil, &TelegramError{
			Code:        resp.StatusCode,
			Description: "не удалось получить информацию о боте",
		}
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

	resp, err := b.client.Post(telegramAPI+b.token+"/sendMessage", "application/json", bytes.NewBuffer(body))
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
		return &TelegramError{
			Code:        resp.StatusCode,
			Description: result.Description,
		}
	}

	return nil
}

func ParseRetryAfter(err *TelegramError) (int, error) {
	if err.Code == 429 {
		matches := regexp.MustCompile(`retry after (\d+)`).FindStringSubmatch(err.Description)
		retryAfter, err := strconv.Atoi(matches[1])
		if err != nil {
			return 0, fmt.Errorf("failed to convert retry after to int: %w", err)
		}
		return retryAfter, nil
	}

	return 0, nil
}

func Pluralize(n int, singular, plural1, plural2 string) string {
	n = n % 100
	if n > 10 && n < 20 {
		return plural2
	}

	n = n % 10
	if n == 1 {
		return singular
	}

	if n > 1 && n < 5 {
		return plural1
	}

	return plural2
}

func (e *TelegramError) Error() string {
	return fmt.Sprintf("API Telegram вернуло ошибку %d: %s", e.Code, e.Description)
}
