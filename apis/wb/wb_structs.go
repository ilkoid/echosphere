package apis

import (
	"time"
)

// ------------------- DATA STRUCT FOR UNMARHSALING FROM JSON ----------------------

// GET feedbacks-api.wildberries.ru/api/v1/feedbacks
type FeedbackResponse struct {
	Data             FeedbackData `json:"data"`
	Error            bool         `json:"error"`
	ErrorText        string       `json:"errorText"`
	AdditionalErrors *string      `json:"additionalErrors"`
}

type FeedbackData struct {
	CountUnanswered int        `json:"countUnanswered"`
	CountArchive    int        `json:"countArchive"`
	Feedbacks       []Feedback `json:"feedbacks"`
}

type Feedback struct {
	ID                              string         `json:"id"`
	Text                            string         `json:"text"`
	Pros                            string         `json:"pros"`
	Cons                            string         `json:"cons"`
	ProductValuation                int            `json:"productValuation"`
	CreatedDate                     time.Time      `json:"createdDate"`
	Answer                          Answer         `json:"answer"`
	State                           string         `json:"state"`
	ProductDetails                  ProductDetails `json:"productDetails"`
	Video                           Video          `json:"video"`
	WasViewed                       bool           `json:"wasViewed"`
	PhotoLinks                      []PhotoLink    `json:"photoLinks"`
	UserName                        string         `json:"userName"`
	MatchingSize                    string         `json:"matchingSize"`
	IsAbleSupplierFeedbackValuation bool           `json:"isAbleSupplierFeedbackValuation"`
	SupplierFeedbackValuation       int            `json:"supplierFeedbackValuation"`
	IsAbleSupplierProductValuation  bool           `json:"isAbleSupplierProductValuation"`
	SupplierProductValuation        int            `json:"supplierProductValuation"`
	IsAbleReturnProductOrders       bool           `json:"isAbleReturnProductOrders"`
	ReturnProductOrdersDate         time.Time      `json:"returnProductOrdersDate"`
	Bables                          []string       `json:"bables"`
	LastOrderShkId                  int            `json:"lastOrderShkId"`
	LastOrderCreatedAt              time.Time      `json:"lastOrderCreatedAt"`
	Color                           string         `json:"color"`
	SubjectId                       int            `json:"subjectId"`
	SubjectName                     string         `json:"subjectName"`
	ParentFeedbackId                *string        `json:"parentFeedbackId"` // Pointer to handle null
	ChildFeedbackId                 string         `json:"childFeedbackId"`
}

type Answer struct {
	Text     string `json:"text"`
	State    string `json:"state"`
	Editable bool   `json:"editable"`
}

type ProductDetails struct {
	ImtId           int    `json:"imtId"`
	NmId            int    `json:"nmId"`
	ProductName     string `json:"productName"`
	SupplierArticle string `json:"supplierArticle"`
	SupplierName    string `json:"supplierName"`
	BrandName       string `json:"brandName"`
	Size            string `json:"size"`
}

type Video struct {
	PreviewImage string `json:"previewImage"`
	Link         string `json:"link"`
	DurationSec  int    `json:"durationSec"`
}

type PhotoLink struct {
	FullSize string `json:"fullSize"`
	MiniSize string `json:"miniSize"`
}

// ------------------- DATA STRUCT FOR HUMAN USAGE ----------------------

// TODO: ADD BARCODE
type WBReviewCondensed struct {
	ID               string
	UserName         string
	CreatedDate      time.Time
	ProductValuation int
	State            string
	Text             string
	Answer           Answer
	Product          WBProductDetails
}

type WBProductDetails struct {
	ProductDetails ProductDetails
	MatchingSize   string
	Color          string
	Pros           string
	Cons           string
	Video          Video
	PhotoLinks     []PhotoLink
}

func ConvertWbIntoCondensed(f FeedbackResponse) []WBReviewCondensed {
	res := []WBReviewCondensed{}
	for _, feedback := range f.Data.Feedbacks {
		cur := WBReviewCondensed{
			ID:               feedback.ID,
			UserName:         feedback.UserName,
			CreatedDate:      feedback.CreatedDate,
			ProductValuation: feedback.ProductValuation,
			State:            feedback.State,
			Text:             feedback.Text,
			Answer:           feedback.Answer,
			Product: WBProductDetails{
				ProductDetails: feedback.ProductDetails,
				MatchingSize:   feedback.MatchingSize,
				Color:          feedback.Color,
				Pros:           feedback.Pros,
				Cons:           feedback.Cons,
				Video:          feedback.Video,
				PhotoLinks:     feedback.PhotoLinks,
			},
		}
		res = append(res, cur)
	}
	return res
}
