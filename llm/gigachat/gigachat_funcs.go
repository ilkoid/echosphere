package llm

import (
	"bufio"
	"crypto/tls"
	"echosphere/apis"
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

func GetReviewRespone(accessToken string, url string, productName string, rating int, reviewText string) (string, error) {
	requestData := GigachatChatCompletionRequest{
		Model: "GigaChat",
		Messages: []GigachatMessageContent{
			{
				Role:    "system",
				Content: "Ты профессиональный специалист клиентской службы Play Today. Не дублируй сообщение в отзыве. Не оставляй контактных данных(номера телефонов, электронных почт) и не пиши про возврат товара или его замену на другой",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Исходи что товар - %s, а его оценка, которую ему присвоили - %d. Ответь на слудющий отзыв: %s", productName, rating, reviewText),
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
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

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

func GetReviewMood(accessToken string, url string, productName string, rating int, reviewText string) (string, error) {
	requestData := GigachatChatCompletionRequest{
		Model: "GigaChat",
		Messages: []GigachatMessageContent{
			{
				Role:    "system",
				Content: "Ты профессиональный оценщик отзывов. Твоя задача оценить одним словом тональность (позитивный, нейтральный, негативный) отзыва.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Исходи что товар - %s, а его оценка, которую ему присвоили - %d. Ответь на слудющий отзыв: %s", productName, rating, reviewText),
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
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

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

func GetReviewKeyWord(accessToken string, url string, productName string, rating int, reviewText string) (string, error) {
	requestData := GigachatChatCompletionRequest{
		Model: "GigaChat",
		Messages: []GigachatMessageContent{
			{
				Role:    "system",
				Content: "Ты профессиональный оценщик отзывов. Твоя задача выписать 1-3 ключевых слова относящихся к товару.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("отзыв: %s", reviewText),
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
	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJson)))
	if err != nil {
		return "", err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

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

func (g *GigachatAPI) MakeResponse(reviews []apis.ReviewCondensed) ([]apis.ReviewCondensed, error) {
	accessToken, err := GetAccesToken()
	if err != nil {
		return []apis.ReviewCondensed{}, err
	}

	url := "https://gigachat.devices.sberbank.ru/api/v1/chat/completions"

	for i := range reviews {
		if len(reviews[i].Text) == 0 {
			// implement some templated responses without the use of llms...
			continue
		}

		reviewResponse, err := GetReviewRespone(accessToken, url, reviews[i].Product.Name, reviews[i].Rating, reviews[i].Text)
		if err != nil {
			return reviews, err
		}

		reviewMood, err := GetReviewMood(accessToken, url, reviews[i].Product.Name, reviews[i].Rating, reviews[i].Text)
		if err != nil {
			return reviews, err
		}

		reviewKeyWord, err := GetReviewKeyWord(accessToken, url, reviews[i].Product.Name, reviews[i].Rating, reviews[i].Text)
		if err != nil {
			return reviews, err
		}

		reviews[i].Response = reviewResponse
		reviews[i].Mood = reviewMood
		reviews[i].KeyWords = reviewKeyWord

	}

	return reviews, nil
}
