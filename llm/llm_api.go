package llm

import (
	"echosphere/apis"
	"fmt"
	"gopkg.in/yaml.v3"
	"os"
)

type LargeLanguageModelAPI interface {
	MakeResponse(reviews []apis.ReviewCondensed) ([]apis.ReviewCondensed, error)
	// GetFeelings(reviews []apis.ReviewCondensed) []apis.ReviewCondensed // to be added later
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
