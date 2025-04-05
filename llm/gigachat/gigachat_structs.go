package llm

type GigachatMessage struct {
	Content string `json:"content"`
	Role    string `json:"role"`
}

type GigachatChoice struct {
	FinishReason string          `json:"finish_reason"`
	Index        int             `json:"index"`
	Message      GigachatMessage `json:"message"`
}

type GigachatUsage struct {
	CompletionTokens int `json:"completion_tokens"`
	PromptTokens     int `json:"prompt_tokens"`
	SystemTokens     int `json:"system_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

type GigachatChatCompletion struct {
	Choices []GigachatChoice `json:"choices"`
	Created int64            `json:"created"`
	Model   string           `json:"model"`
	Usage   GigachatUsage    `json:"usage"`
}

type GigachatAPI struct {
	accessToken   string
	completionURL string
}

var gigachatInstance *GigachatAPI

func GetGigachatAPI() *GigachatAPI {
	if gigachatInstance == nil {
		accessToken, err := GetAccesToken()
		if err != nil {
			return nil
		}
		return &GigachatAPI{
			accessToken:   accessToken,
			completionURL: "https://gigachat.devices.sberbank.ru/api/v1/chat/completions",
		}
	}

	return gigachatInstance
}

type GigachatChatCompletionRequest struct {
	Model    string                   `json:"model"`
	Messages []GigachatMessageContent `json:"messages"`

	// отрезок [0, 1]. Параметр используется как альтернатива температуре (поле temperature).
	// Задает вероятностную массу токенов, которые должна учитывать модель.
	// Так, если передать значение 0.1, модель будет учитывать только токены, чья вероятностная масса входит в верхние 10%.
	TopP              float32 `json:"top_p"`
	RepetitionPenalty float32 `json:"repetiotion_penalty"`

	Stream bool `json:"stream"`
	// Параметр потокового режима ("stream": "true"). Задает минимальный интервал в секундах, который проходит между отправкой токенов.
	// Например, если указать 1, сообщения будут приходить каждую секунду, но размер каждого из них будет больше, так как за секунду накапливается много токенов.
	UpdateInterval int `json:"update_interval"`
}

type GigachatMessageContent struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}
