package apis

import (
	"echosphere/apis"
	"encoding/json"
	"fmt"
	"image/png"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"

	"golang.org/x/image/webp"
)

type AdditionalProductData struct {
	Cards []Card `json:"cards"`
}

type Card struct {
	VendorCode  string  `json:"vendorCode"`
	Description string  `json:"description"`
	Photos      []Photo `json:"photos"`
}

type Photo struct {
	Big      string `json:"big"`
	C246x328 string `json:"c246x328"`
	C516x688 string `json:"c516x688"`
	Square   string `json:"square"`
	Tm       string `json:"tm"`
}

type AdditionalProductDataRequest struct {
	Settings Settings `json:"settings"`
}

type Settings struct {
	Filter Filter `json:"filter"`
}

type Filter struct {
	TextSearch string `json:"textSearch"`
}

func RequestProductAdditionalData(wbKey string, nmID int) (AdditionalProductData, error) {
	url := "https://content-api.wildberries.ru/content/v2/get/cards/list?locale=ru"

	requestData := AdditionalProductDataRequest{
		Settings{
			Filter{
				TextSearch: strconv.Itoa(nmID),
			},
		},
	}

	requestJson, err := json.Marshal(requestData)
	if err != nil {
		return AdditionalProductData{}, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(requestJson)))
	if err != nil {
		return AdditionalProductData{}, err
	}

	req.Header.Set("Authorization", "Bearer "+wbKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return AdditionalProductData{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return AdditionalProductData{}, fmt.Errorf("Response status code is not OK: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return AdditionalProductData{}, err
	}

	var additionalData AdditionalProductData
	if err := json.Unmarshal(body, &additionalData); err != nil {
		fmt.Println(err)
		return AdditionalProductData{}, err
	}

	return additionalData, nil
}

func ConvertWebPIntoPNG(photoURL string) (string, error) {
	response, err := http.Get(photoURL)
	if err != nil {
		return "", err
	}
	defer response.Body.Close()

	img, err := webp.Decode(response.Body)
	if err != nil {
		return "", err
	}

	pngFile, err := os.Create("tmp/tmp.png")
	if err != nil {
		return "", fmt.Errorf("Could not create tmp.png: tmp dir does not exist")
	}
	defer func() {
		pngFile.Close()
		os.Remove("tmp/tmp.png")
	}()

	err = png.Encode(pngFile, img)
	if err != nil {
		return "", err
	}

	if _, err = pngFile.Seek(0, io.SeekStart); err != nil {
		return "", err
	}

	byteSlice, err := io.ReadAll(pngFile)
	if err != nil {
		return "", err
	}

	return string(byteSlice), nil
}

func ConvertWBReview(wbKey string, wbReview Feedback) apis.ReviewCondensed {
	additionalData, err := RequestProductAdditionalData(wbKey, wbReview.ProductDetails.NmId)
	if err != nil {
		fmt.Println(err)
	}

	var photoByteSlices []string
	for _, photoURL := range additionalData.Cards[0].Photos {
		byteSlice, err := ConvertWebPIntoPNG(photoURL.Big)
		if err != nil {
			fmt.Println(err)
		}
		photoByteSlices = append(photoByteSlices, byteSlice)
	}

	return apis.ReviewCondensed{
		ID:          wbReview.ID,
		PublishedAt: wbReview.CreatedDate,
		Status:      wbReview.State,
		Rating:      wbReview.ProductValuation,
		Text:        wbReview.Text,
		Product: apis.Product{
			ID:          wbReview.ProductDetails.NmId,  // NmId как ID товара
			TypeID:      wbReview.ProductDetails.ImtId, // ImtId как TypeID
			Name:        wbReview.ProductDetails.ProductName,
			Description: additionalData.Cards[0].Description,
			VendorCode:  additionalData.Cards[0].VendorCode,
			Photos:      photoByteSlices,
		},
	}
}

func ConvertWbIntoCondensed(wbKey string, f FeedbackResponse) []apis.ReviewCondensed {
	res := []apis.ReviewCondensed{}
	for _, feedback := range f.Data.Feedbacks {
		res = append(res, ConvertWBReview(wbKey, feedback))
	}
	return res
}

func RequestFeedback(wbKey string, config apis.FeedbackRequestConfig) (FeedbackResponse, error) {
	url := fmt.Sprintf("https://feedbacks-api.wildberries.ru/api/v1/feedbacks?isAnswered=%t&take=%d&skip=%d", config.IsAnswered, config.Take, config.Skip)

	if config.DateFrom != nil {
		url = url + fmt.Sprintf("&dateFrom=%d", config.DateFrom.Unix())
	}
	if config.DateTo != nil {
		url = url + fmt.Sprintf("&dateTo=%d", config.DateTo.Unix())
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return FeedbackResponse{}, err
	}

	req.Header.Set("Authorization", "Bearer "+wbKey)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return FeedbackResponse{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return FeedbackResponse{}, fmt.Errorf("Response status code is not OK: %v", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return FeedbackResponse{}, err
	}

	var feedback FeedbackResponse
	if err := json.Unmarshal(body, &feedback); err != nil {
		return FeedbackResponse{}, err
	}

	return feedback, nil
}

func (w *WBAPI) GetFeedback(config apis.FeedbackRequestConfig) ([]apis.ReviewCondensed, error) {
	wbKey := os.Getenv("WB_API_SAFE")

	feedback, err := RequestFeedback(wbKey, config)
	if err != nil {
		return []apis.ReviewCondensed{}, err
	}

	wbAccessKey := os.Getenv("WB_API_CONTENT")
	return ConvertWbIntoCondensed(wbAccessKey, feedback), nil
}
