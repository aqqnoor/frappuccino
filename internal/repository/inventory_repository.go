package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"github.com/sunzhqr/frappuccino/internal/models"
)

type InventoryRepositoryInterface interface {
	GetAll() ([]models.InventoryItem, error)
	Exists(ID int) bool
	SubtractIngredients(ingredients map[int]float64) error
	AddInventoryItemRepo(item models.InventoryItem) error
	UpdateItemRepo(id int, newItem models.InventoryItem) error
	DeleteItemRepo(id int) error
	GetLeftOvers(sortBy, page, pageSize string) (map[string]any, error)
}

type InventoryRepository struct {
	db *sql.DB
}

func NewInventoryRepository(db *sql.DB) *InventoryRepository {
	return &InventoryRepository{db: db}
}

func (repo *InventoryRepository) GetAll() ([]models.InventoryItem, error) {
	queryGetIngredients := `
	select IngredientID, Name, Quantity, Unit from inventory
	`
	rows, err := repo.db.Query(queryGetIngredients)
	if err != nil {
		return nil, err
	}
	var inventoryItems []models.InventoryItem

	for rows.Next() {
		var inventoryItem models.InventoryItem
		err = rows.Scan(&inventoryItem.IngredientID, &inventoryItem.Name, &inventoryItem.Quantity, &inventoryItem.Unit)
		if err != nil {
			return nil, err
		}
		inventoryItems = append(inventoryItems, inventoryItem)
	}
	return inventoryItems, nil
}

func (repo *InventoryRepository) Exists(ID int) bool {
	queryIfExists := `
	select IngredientID from inventory where IngredientID = $1
	`
	rows, err := repo.db.Query(queryIfExists, ID)
	if err != nil {
		return false
	}
	return rows.Next()
}

func (repo *InventoryRepository) SubtractIngredients(ingredients map[int]float64) error {
	for key, value := range ingredients {
		queryToSubtract := `
	        update inventory
	        set Quantity  = Quantity - $1
	        where IngredientID = $2
	    `
		_, err := repo.db.Exec(queryToSubtract, value, key)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *InventoryRepository) AddInventoryItemRepo(item models.InventoryItem) error {
	queryToAddInventory := `
	insert into inventory (Name, Quantity, Unit) values
	($1, $2, $3)
	`
	_, err := repo.db.Exec(queryToAddInventory, item.Name, item.Quantity, item.Unit)
	if err != nil {
		return err
	}
	return nil
}

func (repo *InventoryRepository) UpdateItemRepo(id int, newItem models.InventoryItem) error {
	queryToUpdate := `
	update inventory
	set Quantity = $1, Name = $2, Unit = $3
	where IngredientID = $4
	`
	_, err := repo.db.Exec(queryToUpdate, newItem.Quantity, newItem.Name, newItem.Unit, id)
	if err != nil {
		return err
	}
	return nil
}

func (repo *InventoryRepository) DeleteItemRepo(id int) error {
	queryToDelete := `
	delete from inventory
	where IngredientID = $1
	`
	_, err := repo.db.Exec(queryToDelete, id)
	if err != nil {
		return err
	}
	return nil
}

func (repo *InventoryRepository) GetLeftOvers(sortBy, page, pageSize string) (map[string]any, error) {
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum <= 0 {
		pageNum = 1
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum <= 0 {
		pageSizeNum = 10
	}

	offset := (pageNum - 1) * pageSizeNum

	query := `
        SELECT i.IngredientID, i.Name, i.Quantity, i.Unit
        FROM inventory i
    `
	switch sortBy {
	case "price":
		query += " ORDER BY i.Price"
	case "quantity":
		query += " ORDER BY i.Quantity"
	default:
		return nil, errors.New("invalid sortBy value, must be 'price' or 'quantity'")
	}

	query += fmt.Sprintf(" LIMIT %d OFFSET %d", pageSizeNum, offset)

	var leftovers []map[string]any
	rows, err := repo.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var ingredientID int
		var name string
		var quantity int
		var unit string
		if err := rows.Scan(&ingredientID, &name, &quantity, &unit); err != nil {
			return nil, fmt.Errorf("failed to scan row: %v", err)
		}

		leftovers = append(leftovers, map[string]any{
			"ingredientID": ingredientID,
			"name":         name,
			"quantity":     quantity,
			"unit":         unit,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows iteration error: %v", err)
	}

	var totalItems int
	err = repo.db.QueryRow("SELECT COUNT(*) FROM inventory").Scan(&totalItems)
	if err != nil {
		return nil, fmt.Errorf("failed to count total items: %v", err)
	}

	totalPages := (totalItems + pageSizeNum - 1) / pageSizeNum
	hasNextPage := pageNum < totalPages

	response := map[string]any{
		"currentPage": pageNum,
		"hasNextPage": hasNextPage,
		"pageSize":    pageSizeNum,
		"totalPages":  totalPages,
		"data":        leftovers,
	}

	return response, nil
}
