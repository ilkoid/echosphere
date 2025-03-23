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

		rawData := fmt.Sprintf(
			"{\n"+
				"  \"model\": \"GigaChat\",\n"+
				"  \"messages\": [\n"+
				"    {\n"+
				"      \"role\": \"system\",\n"+
				"      \"content\": \"Ты профессиональный маркетолог и выступаешь от лица компании Play Today. Не дублируй сообщение в отзыве. Не оставляй контактных данных(номера телефонов, электронных почт) и не пиши про возврат товара или его замену на другой\"\n"+
				"    },\n"+
				"    {\n"+
				"      \"role\": \"user\",\n"+
				"      \"content\": \"Исходи что товар - %s, а его оценка, которую ему присвоили - %d. Ответь на слудющий отзыв: %s\"\n"+
				"    }\n"+
				"  ],\n"+
				"  \"stream\": false,\n"+
				"  \"repetition_penalty\": 1\n"+
				"}\n",
			reviews[i].Product.Name, reviews[i].Rating, reviews[i].Text)
		req, err := http.NewRequest("POST", url, strings.NewReader(rawData))
		if err != nil {
			return []apis.ReviewCondensed{}, err
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
			return reviews, err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return reviews, err
		}
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return reviews, err
		}

		var response GigachatChatCompletion
		if err := json.Unmarshal(body, &response); err != nil {
			return []apis.ReviewCondensed{}, err
		}

		reviews[i].Response = response.Choices[0].Message.Content
	}

	return reviews, nil
}
