package models

type TotalSales struct {
	TotalSales int `json:"total_sales"`
}

type PopularItems struct {
	Items []PopularItem `json:"popular_items"`
}

type PopularItem struct {
	ProductID   int    `json:"product_id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Quantity    int    `json:"quantity"`
}

type SearchResult struct {
	MenuItems    []SearchMenuItem    `json:"menu_items"`
	Orders       []SearchOrderResult `json:"orders"`
	TotalMatches int                 `json:"total_matches"`
}

type SearchMenuItem struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Relevance   float64 `json:"relavance"`
}

type SearchOrderResult struct {
	ID           int      `json:"id"`
	CustomerName string   `json:"customer_name"`
	Items        []string `json:"items"`
	Total        float64  `json:"total"`
	Relevance    float64  `json:"relavance"`
}
