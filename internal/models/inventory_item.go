package models

type InventoryItem struct {
	IngredientID int     `json:"ingredient_id"`
	Name        string  `json:"name"`
	Quantity    float64 `json:"quantity"`
	Unit        string  `json:"unit"`
}
