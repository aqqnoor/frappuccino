package models

var (
	StatusOrderAccepted = "accepted"
	StatusOrderRejected = "rejected"
)

type Order struct {
	ID           int                    `json:"order_id"`
	CustomerName string                 `json:"customer_name"`
	Items        []OrderItem            `json:"items"`
	Status       string                 `json:"status"`
	Notes        map[string]interface{} `json:"notes"`
	CreatedAt    string                 `json:"created_at"`
}

type OrderItem struct {
	ProductID int `json:"product_id"`
	Quantity  int `json:"quantity"`
}

type BatchOrdersResponce struct {
	Processed_orders []BatchOrderInfo  `json:"processed_orders"`
	Summary          BatchOrderSummary `json:"summary"`
}

type BatchOrderInfo struct {
	OrderID      int     `json:"order_id"`
	CustomerName string  `json:"customer_name"`
	Status       string  `json:"status"`
	Reason       string  `json:"reason"`
	Total        float64 `json:"total"`
}

type BatchOrderSummary struct {
	TotalOrders      int                         `json:"total_orders"`
	Accepted         int                         `json:"accepted"`
	Rejected         int                         `json:"rejected"`
	TotalRevenue     float64                     `json:"total_revenue"`
	InventoryUpdates []BatchOrderInventoryUpdate `json:"inventory_updates"`
}

type BatchOrderInventoryUpdate struct {
	IngredientID  int    `json:"ingredient_id"`
	Name          string `json:"name"`
	Quantity_used int    `json:"quantity_used"`
	Remaining     int    `json:"remaining"`
}