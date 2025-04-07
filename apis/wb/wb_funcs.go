package apis

import (
	"echosphere/apis"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func ConvertToByteSlice(photolink string) (string, error) {
	fmt.Print(photolink)
	response, err := http.Get(photolink)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		return "", err
	}
	return string(bodyBytes), nil
}

func ConvertWBReview(wbReview Feedback) apis.ReviewCondensed {
	fmt.Print(len(wbReview.PhotoLinks))
	for _, photolink := range wbReview.PhotoLinks {
		_, err := ConvertToByteSlice(photolink.FullSize)
		fmt.Print(err)
	}
	return apis.ReviewCondensed{
		ID:          wbReview.ID,
		PublishedAt: wbReview.CreatedDate,
		Status:      wbReview.State,
		Rating:      wbReview.ProductValuation,
		Text:        wbReview.Text,
		Product: apis.Product{
			ID:     wbReview.ProductDetails.NmId,  // NmId как ID товара
			TypeID: wbReview.ProductDetails.ImtId, // ImtId как TypeID
			Name:   wbReview.ProductDetails.ProductName,

		},
	}
}

func ConvertWbIntoCondensed(f FeedbackResponse) []apis.ReviewCondensed {
	res := []apis.ReviewCondensed{}
	for _, feedback := range f.Data.Feedbacks {
		res = append(res, ConvertWBReview(feedback))
	}
	return res
}

func (w *WBAPI) GetFeedback(config apis.FeedbackRequestConfig) ([]apis.ReviewCondensed, error) {
	wb_key := os.Getenv("WB_API_SAFE")

	url := fmt.Sprintf("https://feedbacks-api.wildberries.ru/api/v1/feedbacks?isAnswered=%t&take=%d&skip=%d", config.IsAnswered, config.Take, config.Skip)

	if config.DateFrom != nil {
		url = url + fmt.Sprintf("&dateFrom=%d", config.DateFrom.Unix())
	}
	if config.DateTo != nil {
		url = url + fmt.Sprintf("&dateTo=%d", config.DateTo.Unix())
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return []apis.ReviewCondensed{}, err
	}

	req.Header.Set("Authorization", "Bearer "+wb_key)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return []apis.ReviewCondensed{}, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return []apis.ReviewCondensed{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []apis.ReviewCondensed{}, err
	}

	var feedback FeedbackResponse
	if err := json.Unmarshal(body, &feedback); err != nil {
		return []apis.ReviewCondensed{}, err
	}

	return ConvertWbIntoCondensed(feedback), nil
}
