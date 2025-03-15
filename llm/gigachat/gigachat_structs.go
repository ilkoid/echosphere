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

type GigachatAPI struct{}
