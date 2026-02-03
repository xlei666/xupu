package models

// FanqieBook Represents a book item from Fanqie Novel API
type FanqieBook struct {
	BookID       string  `json:"bookId"`
	BookName     string  `json:"bookName"`
	Author       string  `json:"author"`
	ThumbURI     string  `json:"thumbUri"`
	Category     string  `json:"category"`
	Score        float64 `json:"score"`
	ReadCount    int     `json:"read_count"`
	Description  string  `json:"description"`
	CreationTime string  `json:"creationTime"`
}

// FanqieRankResponse Represents the API response structure
type FanqieRankResponse struct {
	Code int `json:"code"`
	Data struct {
		BookList []FanqieBook `json:"book_list"`
		HasMore  bool         `json:"has_more"`
	} `json:"data"`
	Message string `json:"message"`
}
