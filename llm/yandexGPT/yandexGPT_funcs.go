package yandexgpt

import (
	"bytes"
	"echosphere/apis"
	"echosphere/llm"
	"encoding/json"
	"io"
	"net/http"
	"os/exec"
	"strings"

	"fmt"
)

func New() *YandexApi {
	return &YandexApi{
		CompletionURL: "https://llm.api.cloud.yandex.net/foundationModels/v1/completion",
		IamToken:      "",
	}
}

func (y *YandexApi) GenerateResponse(review apis.ReviewCondensed, context string) (string, error) {
	requestHeaders, err := llm.GetRequestHeader("config.yaml")
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("Исходи что товар - %s, а его оценка, которую ему присвоили - %d. Ответь на слудющий отзыв: %s. ", review.Product.Name, review.Rating, review.Text)
	if context != "" {
		content = content + context
	}

	requestData := Request{
		ModelURI: "gpt://b1gu9lktlgs9o8824hq5/yandexgpt",
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: 0.5,
		},
		Messages: []RequestMessage{
			{
				Role: "system",
				Text: requestHeaders.ResponseHeader,
			},
			{
				Role: "user",
				Text: content,
			},
		},
	}

	requestJson, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", y.CompletionURL, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+y.IamToken)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Response status code is not OK: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Result.Alternatives[0].Message.Text, nil
}

func (y *YandexApi) GetMood(review apis.ReviewCondensed, context string) (string, error) {
	requestHeaders, err := llm.GetRequestHeader("config.yaml")
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("Исходи что товар - %s, а его оценка, которую ему присвоили - %d. Ответь на слудющий отзыв: %s. ", review.Product.Name, review.Rating, review.Text)
	if context != "" {
		content = content + context
	}

	requestData := Request{
		ModelURI: "gpt://b1gu9lktlgs9o8824hq5/yandexgpt",
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: 0.5,
		},
		Messages: []RequestMessage{
			{
				Role: "system",
				Text: requestHeaders.MoodHeader,
			},
			{
				Role: "user",
				Text: content,
			},
		},
	}

	requestJson, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", y.CompletionURL, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+y.IamToken)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Response status code is not OK: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Result.Alternatives[0].Message.Text, nil
}

func (y *YandexApi) FindKeywords(review apis.ReviewCondensed, context string) (string, error) {
	requestHeaders, err := llm.GetRequestHeader("config.yaml")
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("Отзыв: %s", review.Text)
	if context != "" {
		content = content + context
	}

	requestData := Request{
		ModelURI: "gpt://b1gu9lktlgs9o8824hq5/yandexgpt",
		CompletionOptions: CompletionOptions{
			Stream:      false,
			Temperature: 0.5,
		},
		Messages: []RequestMessage{
			{
				Role: "system",
				Text: requestHeaders.MoodHeader,
			},
			{
				Role: "user",
				Text: content,
			},
		},
	}

	requestJson, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", y.CompletionURL, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+y.IamToken)

	var client http.Client
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Response status code is not OK: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Result.Alternatives[0].Message.Text, nil
}

func (y *YandexApi) ValidateKey() error {
	cmd := exec.Command("yc", "iam", "create-token")

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("Command execution failed: %v\nError output: %s\n", err, stderr.String())
	}

	y.IamToken = strings.TrimSpace(out.String())
	return nil
}
