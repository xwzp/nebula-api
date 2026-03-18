package brave_search

type SearchRequest struct {
	Model        string `json:"model"`
	Query        string `json:"q"`
	Count        *int   `json:"count,omitempty"`
	Offset       *int   `json:"offset,omitempty"`
	Country      string `json:"country,omitempty"`
	SearchLang   string `json:"search_lang,omitempty"`
	Freshness    string `json:"freshness,omitempty"`
	TextFormat   string `json:"text_format,omitempty"`
	ResultFilter string `json:"result_filter,omitempty"`
}
