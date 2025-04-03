package llm

import (
	"echosphere/apis"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type LargeLanguageModelAPI interface {
	GenerateResponse(review apis.ReviewCondensed, context string) (string, error)
	GetMood(review apis.ReviewCondensed, context string) (string, error)
	FindKeywords(review apis.ReviewCondensed, context string) (string, error)
}

type LlmRequestHeaders struct {
	ResponseHeader string `yaml:"response_header"`
	MoodHeader     string `yaml:"mood_header"`
	KeyWordsHeader string `yaml:"keywords_header"`
}

func GetRequestHeader(filePath string) (LlmRequestHeaders, error) {
	requestHeaders := LlmRequestHeaders{}
	yamlFile, err := os.ReadFile(filePath)
	if err != nil {
		return requestHeaders, fmt.Errorf("Could not get request header: %v ", err)
	}
	err = yaml.Unmarshal(yamlFile, &requestHeaders)
	if err != nil {
		return requestHeaders, fmt.Errorf("Could not parse request header: %v ", err)
	}

	return requestHeaders, nil
}
