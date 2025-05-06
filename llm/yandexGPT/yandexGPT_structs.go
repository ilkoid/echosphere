package yandexgpt

type Request struct {
	ModelURI          string            `json:"modelUri"`
	CompletionOptions CompletionOptions `json:"completionOptions"`
	Messages          []RequestMessage         `json:"messages"`
}

type CompletionOptions struct {
	Stream           bool             `json:"stream"`
	Temperature      float64          `json:"temperature"`
}

type RequestMessage struct {
	Role string `json:"role"`
	Text string `json:"text"`
}

type YandexApi struct {
	CompletionURL string
	IamToken string
}

type Response struct {
	Result Result `json:"result"`
}

type Result struct {
	Alternatives []Alternative `json:"alternatives"`
}

type Alternative struct {
	Message ResponseMessage `json:"message"`
	Status  string  `json:"status"`
}

type ResponseMessage struct {
	Text string `json:"text"`
}
