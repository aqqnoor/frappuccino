package repository

import (
	"database/sql"

	"github.com/sunzhqr/frappuccino/internal/models"
)

type MenuRepositoryInterface interface {
	GetAll() ([]models.MenuItem, error)
	Exists(itemID int) bool
	DeleteMenuItemRepo(MenuItemID int) error
	UpdateMenuItemRepo(menuItem models.MenuItem) error
	AddMenuItemRepo(menuItem models.MenuItem) error
	MenuCheckByIDRepo(ID int) bool
}

// MenuRepository implements MenuRepository using JSON files
type MenuRepository struct {
	db *sql.DB
}

// NewMenuRepository creates a new FileMenuRepository
func NewMenuRepository(db *sql.DB) *MenuRepository {
	return &MenuRepository{db: db}
}

func (repo *MenuRepository) GetAll() ([]models.MenuItem, error) {
	queryMenuItems := `
	select ID, Name, Description, Price from menu_items
	`
	rows, err := repo.db.Query(queryMenuItems)
	if err != nil {
		return []models.MenuItem{}, err
	}
	var MenuItems []models.MenuItem
	for rows.Next() {
		var MenuItem models.MenuItem
		rows.Scan(&MenuItem.ID, &MenuItem.Name, &MenuItem.Description, &MenuItem.Price)
		var MenuItemIngredients []models.MenuItemIngredient
		queryMenuItemIngredients := `
	        select IngredientID, Quantity from menu_item_ingredients where MenuID = $1
	    `
		rows1, err := repo.db.Query(queryMenuItemIngredients, MenuItem.ID)
		if err != nil {
			return []models.MenuItem{}, err
		}
		for rows1.Next() {
			var MenuItemIngredient models.MenuItemIngredient
			rows1.Scan(&MenuItemIngredient.IngredientID, &MenuItemIngredient.Quantity)
			MenuItemIngredients = append(MenuItemIngredients, MenuItemIngredient)
		}
		MenuItem.Ingredients = MenuItemIngredients
		MenuItems = append(MenuItems, MenuItem)
	}
	return MenuItems, nil
}

func (repo *MenuRepository) Exists(itemID int) bool {
	items, _ := repo.GetAll()
	for _, item := range items {
		if item.ID == itemID {
			return true
		}
	}
	return false
}

func (repo *MenuRepository) DeleteMenuItemRepo(MenuItemID int) error {
	queryDeleteMenuItem := `
	delete from menu_items
	where ID = $1
	`
	_, err := repo.db.Exec(queryDeleteMenuItem, MenuItemID)
	if err != nil {
		return err
	}
	return nil
}

func (repo *MenuRepository) UpdateMenuItemRepo(menuItem models.MenuItem) error {
	queryUpdateMenu := `
	update menu_items
	set Name = $1, Description = $2, Price = $3
	where ID = $4
	`
	_, err := repo.db.Exec(queryUpdateMenu, menuItem.Name, menuItem.Description, menuItem.Price, menuItem.ID)
	if err != nil {
		return err
	}
	queryUpdateMenuIngredients1 := `
			delete from menu_item_ingredients 
			where MenuID = $1
		`
	// Execute the update query
	_, err = repo.db.Exec(queryUpdateMenuIngredients1, menuItem.ID)
	if err != nil {
		return err
	}
	for _, v := range menuItem.Ingredients {

		queryUpdateMenuIngredients2 := `
			insert into menu_item_ingredients (MenuID, IngredientID, Quantity) values
			($1, $2, $3)
		`
		_, err = repo.db.Exec(queryUpdateMenuIngredients2, menuItem.ID, v.IngredientID, v.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *MenuRepository) AddMenuItemRepo(menuItem models.MenuItem) error {
	queryAddItem := `
	Insert into menu_items (Name, Description, Price) values
    ($1, $2, $3)
	RETURNING id
	`
	var menuID int

	err := repo.db.QueryRow(queryAddItem, menuItem.Name, menuItem.Description, menuItem.Price).Scan(&menuID)
	if err != nil {
		return err
	}
	for _, v := range menuItem.Ingredients {
		queryAddItemIngredients := `
		insert into menu_item_ingredients (MenuID, IngredientID, Quantity) values
		($1, $2, $3)
	    `
		_, err = repo.db.Exec(queryAddItemIngredients, menuID, v.IngredientID, v.Quantity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (repo *MenuRepository) MenuCheckByIDRepo(ID int) bool {
	queryIfExists := `
	select ID from menu_items where ID = $1
	`
	rows, err := repo.db.Query(queryIfExists, ID)
	if err != nil {
		return false
	}
	return rows.Next()
}
