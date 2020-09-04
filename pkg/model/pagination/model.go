package pagination

// Sort options
type Sort struct {
	Field   string `json:"f"`
	OrderBy string `json:"o"`
}

// Pagination configuration
type Pagination struct {
	Limit  int64  `json:"limit"`
	Offset int64  `json:"offset"`
	SortBy []Sort `json:"sortBy"`
}

// Filter used to limit the result
type Filter struct {
	Key      string      `json:"k"`
	Operator string      `json:"o"`
	Value    interface{} `json:"v"`
}

// Result returns a list of data
type Result struct {
	Count  int64       `json:"count"`
	Limit  int64       `json:"limit"`
	Offset int64       `json:"offset"`
	Data   interface{} `json:"data"`
}