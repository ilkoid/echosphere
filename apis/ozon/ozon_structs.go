package apis

import (
	"time"
)

// ------------------- DATA STRUCT FOR UNMARHSALING FROM JSON ----------------------

// POST /v1/review/list
type ReviewResponse struct {
	HasNext bool     `json:"has_next"`
	LastID  string   `json:"last_id"`
	Reviews []Review `json:"reviews"`
}

type Review struct {
	CommentsAmount      int32     `json:"comments_amount"`
	ID                  string    `json:"id"`
	IsRatingParticipant bool      `json:"is_rating_participant"`
	OrderStatus         string    `json:"order_status"`
	PhotosAmount        int32     `json:"photos_amount"`
	PublishedAt         time.Time `json:"published_at"`
	Rating              int32     `json:"rating"`
	SKU                 int64     `json:"sku"`
	Status              string    `json:"status"` // UNPROCESSED or PROCESSED
	Text                string    `json:"text"`
	VideosAmount        int32     `json:"videos_amount"`
}

// POST /v1/review/comment/list
type CommentResponse struct {
	Comments []Comment `json:"comments"`
	Offset   int       `json:"offset"`
}

type Comment struct {
	ID              string    `json:"id"`
	IsOfficial      bool      `json:"is_official"`
	IsOwner         bool      `json:"is_owner"`
	ParentCommentID string    `json:"parent_comment_id"`
	PublishedAt     time.Time `json:"published_at"`
	Text            string    `json:"text"`
}

// POST /v3/product/info/list
type ItemResponse struct {
	Items []Item `json:"items"`
}

type Item struct {
	Barcodes              []string          `json:"barcodes"`
	ColorImage            []string          `json:"color_image"`
	Commissions           []Commission      `json:"commissions"`
	CreatedAt             time.Time         `json:"created_at"`
	CurrencyCode          string            `json:"currency_code"`
	DescriptionCategoryID int               `json:"description_category_id"`
	DiscountedFBOStocks   int               `json:"discounted_fbo_stocks"`
	Errors                []Error           `json:"errors"`
	HasDiscountedFBOItem  bool              `json:"has_discounted_fbo_item"`
	ID                    int               `json:"id"`
	Images                []string          `json:"images"`
	Images360             []string          `json:"images360"`
	IsArchived            bool              `json:"is_archived"`
	IsAutoarchived        bool              `json:"is_autoarchived"`
	IsDiscounted          bool              `json:"is_discounted"`
	IsKGT                 bool              `json:"is_kgt"`
	IsPrepaymentAllowed   bool              `json:"is_prepayment_allowed"`
	IsSuper               bool              `json:"is_super"`
	MarketingPrice        string            `json:"marketing_price"`
	MinPrice              string            `json:"min_price"`
	ModelInfo             ModelInfo         `json:"model_info"`
	Name                  string            `json:"name"`
	OfferID               string            `json:"offer_id"`
	OldPrice              string            `json:"old_price"`
	Price                 string            `json:"price"`
	PriceIndexes          PriceIndexes      `json:"price_indexes"`
	PrimaryImage          []string          `json:"primary_image"`
	Sources               []Source          `json:"sources"`
	Statuses              Statuses          `json:"statuses"`
	Stocks                Stocks            `json:"stocks"`
	TypeID                int               `json:"type_id"`
	UpdatedAt             time.Time         `json:"updated_at"`
	VAT                   string            `json:"vat"`
	VisibilityDetails     VisibilityDetails `json:"visibility_details"`
	VolumeWeight          float64           `json:"volume_weight"`
}

type Commission struct {
	DeliveryAmount int    `json:"delivery_amount"`
	Percent        int    `json:"percent"`
	ReturnAmount   int    `json:"return_amount"`
	SaleSchema     string `json:"sale_schema"`
	Value          int    `json:"value"`
}

type Error struct {
	AttributeID int       `json:"attribute_id"`
	Code        string    `json:"code"`
	Field       string    `json:"field"`
	Level       string    `json:"level"`
	State       string    `json:"state"`
	Texts       ErrorText `json:"texts"`
}

type ErrorText struct {
	AttributeName    string  `json:"attribute_name"`
	Description      string  `json:"description"`
	HintCode         string  `json:"hint_code"`
	Message          string  `json:"message"`
	Params           []Param `json:"params"`
	ShortDescription string  `json:"short_description"`
}

type Param struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ModelInfo struct {
	Count   int `json:"count"`
	ModelID int `json:"model_id"`
}

type PriceIndexes struct {
	ColorIndex                string    `json:"color_index"`
	ExternalIndexData         IndexData `json:"external_index_data"`
	OzonIndexData             IndexData `json:"ozon_index_data"`
	SelfMarketplacesIndexData IndexData `json:"self_marketplaces_index_data"`
}

type IndexData struct {
	MinimalPrice         string `json:"minimal_price"`
	MinimalPriceCurrency string `json:"minimal_price_currency"`
	PriceIndexValue      int    `json:"price_index_value"`
}

type Source struct {
	CreatedAt    time.Time `json:"created_at"`
	QuantCode    string    `json:"quant_code"`
	ShipmentType string    `json:"shipment_type"`
	SKU          int       `json:"sku"`
	Source       string    `json:"source"`
}

type Statuses struct {
	IsCreated         bool      `json:"is_created"`
	ModerateStatus    string    `json:"moderate_status"`
	Status            string    `json:"status"`
	StatusDescription string    `json:"status_description"`
	StatusFailed      string    `json:"status_failed"`
	StatusName        string    `json:"status_name"`
	StatusTooltip     string    `json:"status_tooltip"`
	StatusUpdatedAt   time.Time `json:"status_updated_at"`
	ValidationStatus  string    `json:"validation_status"`
}

type Stocks struct {
	HasStock bool    `json:"has_stock"`
	Stocks   []Stock `json:"stocks"`
}

type Stock struct {
	Present  int    `json:"present"`
	Reserved int    `json:"reserved"`
	SKU      int    `json:"sku"`
	Source   string `json:"source"`
}

type VisibilityDetails struct {
	HasPrice bool `json:"has_price"`
	HasStock bool `json:"has_stock"`
}

// ------------------- DATA STRUCT FOR HUMAN USAGE ----------------------

// TODO: ADD BARCODE
type OzonReviewCondensed struct {
	ID          string
	PublishedAt time.Time
	Rating      int32
	Status      string
	Text        string
	Comments    []Comment
	Product     OzonItemCondensed
}

type OzonItemCondensed struct {
	ID                    int // nmId
	TypeID                int // ImtId
	DescriptionCategoryID int
	Name                  string
	Barcodes              []string
	Images                []string
	Images360             []string
	MarketingPrice        string
	MinPrice              string
	ModelInfo             ModelInfo
	OfferID               string
	OldPrice              string
	Price                 string
	PriceIndexes          PriceIndexes
	PrimaryImage          []string
	Sources               []Source
	CreatedAt             time.Time
	UpdatedAt             time.Time
	VolumeWeight          float64
}
