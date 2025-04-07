package llm

import (
	"bufio"
	"crypto/tls"
	"echosphere/apis"
	"echosphere/llm"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

func GetAccesToken() (string, error) {
	tokenFilePath := "/etc/gigachat_credentials"

	file, err := os.Open(tokenFilePath)
	if err != nil {
		return "", fmt.Errorf("Could not read from token file: %v", err)
	}
	defer file.Close()

	var res string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		res = res + fmt.Sprint(scanner.Text())
	}

	return res, err
}

func (g *GigachatAPI) GenerateResponse(review apis.ReviewCondensed, context string) (string, error) {
	requestHeaders, err := llm.GetRequestHeader("config.yaml")
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("Исходи что товар - %s, а его оценка, которую ему присвоили - %d. Ответь на слудющий отзыв: %s. ", review.Product.Name, review.Rating, review.Text)
	if context != "" {
		content = content + context
	}

	requestData := GigachatChatCompletionRequest{
		Model: "GigaChat",
		Messages: []GigachatMessageContent{
			{
				Role:    "system",
				Content: requestHeaders.ResponseHeader,
			},

			{
				Role:    "user",
				Content: content,
			},
		},
		TopP:              0.5,
		RepetitionPenalty: 1,
		Stream:            false,
		UpdateInterval:    1,
	}

	requestJson, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", g.completionURL, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.accessToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
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

	var response GigachatChatCompletion
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}

func (g *GigachatAPI) GetMood(review apis.ReviewCondensed, context string) (string, error) {
	requestHeaders, err := llm.GetRequestHeader("config.yaml")
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("Исходи что товар - %s, а его оценка, которую ему присвоили - %d. Ответь на слудющий отзыв: %s. ", review.Product.Name, review.Rating, review.Text)
	if context != "" {
		content = content + context
	}

	requestData := GigachatChatCompletionRequest{
		Model: "GigaChat",
		Messages: []GigachatMessageContent{
			{
				Role:    "system",
				Content: requestHeaders.MoodHeader,
			},
			{
				Role:    "user",
				Content: content,
			},
		},
		TopP:              0.5,
		RepetitionPenalty: 1,
		Stream:            false,
		UpdateInterval:    1,
	}

	requestJson, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", g.completionURL, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.accessToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response GigachatChatCompletion
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}

func (g *GigachatAPI) FindKeywords(review apis.ReviewCondensed, context string) (string, error) {
	requestHeaders, err := llm.GetRequestHeader("config.yaml")
	if err != nil {
		return "", err
	}

	content := fmt.Sprintf("Отзыв: %s", review.Text)
	if context != "" {
		content = content + context
	}

	requestData := GigachatChatCompletionRequest{
		Model: "GigaChat",
		Messages: []GigachatMessageContent{
			{
				Role:    "system",
				Content: requestHeaders.KeyWordsHeader,
			},
			{
				Role:    "user",
				Content: content,
			},
		},
		TopP:              0.5,
		RepetitionPenalty: 1,
		Stream:            false,
		UpdateInterval:    1,
	}

	requestJson, err := json.Marshal(requestData)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", g.completionURL, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+g.accessToken)

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	client := &http.Client{Transport: tr}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var response GigachatChatCompletion
	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.Choices[0].Message.Content, nil
}
