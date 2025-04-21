package service

import (
	"errors"
	"strings"

	"github.com/sunzhqr/frappuccino/internal/models"
	"github.com/sunzhqr/frappuccino/internal/repository"
)

type MenuServiceInterface interface {
	AddMenuItem(menuItem models.MenuItem) error
	GetMenuItem(MenuItemID int) (models.MenuItem, error)
	GetMenuItems() ([]models.MenuItem, error)
	CheckNewMenu(MenuItem models.MenuItem) error
	DeleteMenuItem(MenuItemID int) error
	UpdateMenuItem(menuItem models.MenuItem) error
	MenuCheckByID(MenuItemID int, isDelete bool) error
	IngredientsCheckByID(menuItemID int, quantity int) error
	IngredientsCheckForNewItem(menuItem models.MenuItem) error
	SubtractIngredientsByID(OrderID int, quantity int) error
}

type MenuService struct {
	menuRepo      repository.MenuRepositoryInterface
	inventoryRepo repository.InventoryRepositoryInterface
}

func NewMenuService(menuRepo repository.MenuRepositoryInterface, inventoryRepo repository.InventoryRepositoryInterface) *MenuService {
	return &MenuService{menuRepo: menuRepo, inventoryRepo: inventoryRepo}
}

func (s *MenuService) DeleteMenuItem(MenuItemID int) error {
	return s.menuRepo.DeleteMenuItemRepo(MenuItemID)
}

func (s *MenuService) UpdateMenuItem(menuItem models.MenuItem) error {
	return s.menuRepo.UpdateMenuItemRepo(menuItem)
}

func (s *MenuService) MenuCheckByID(MenuItemID int, isDelete bool) error {
	if isDelete {
		flag := false
		if s.menuRepo.MenuCheckByIDRepo(MenuItemID) {
			flag = true
		}
		if flag {
			return nil
		} else {
			return errors.New("the requested menu item does not exist in menu")
		}
	}
	if s.menuRepo.MenuCheckByIDRepo(MenuItemID) {
		return errors.New("the requested menu item to add already exists in menu")
	}
	return nil
}

func (s *MenuService) IngredientsCheckByID(menuItemID int, quantity int) error {
	menuItems, _ := s.menuRepo.GetAll()
	ingredientsNeeded := make(map[int]float64)
	flag := false
	for _, item := range menuItems {
		if item.ID == menuItemID {
			for _, ingr := range item.Ingredients {
				ingredientsNeeded[ingr.IngredientID] += float64(ingr.Quantity) * float64(quantity)
			}
		}
	}

	inventoryItems, _ := s.inventoryRepo.GetAll()

	for _, inventoryItem := range inventoryItems {
		if value, exists := ingredientsNeeded[inventoryItem.IngredientID]; exists {
			flag = true
			if value > inventoryItem.Quantity {
				return errors.New("not enough ingredients for item")
			}
		}
	}
	if flag {
		return nil
	}
	return errors.New("no ingredients for item in inventory")
}

func (s *MenuService) IngredientsCheckForNewItem(menuItem models.MenuItem) error {
	inventoryItems, _ := s.inventoryRepo.GetAll()
	count := 0
	for _, inventoryItem := range inventoryItems {
		for _, ingredients := range menuItem.Ingredients {
			if ingredients.IngredientID == inventoryItem.IngredientID {
				count++
				if ingredients.Quantity > inventoryItem.Quantity {
					return errors.New("not enough ingredients for item")
				}
			}
		}
	}
	if count != len(menuItem.Ingredients) {
		return errors.New("no ingredients for item in inventory")
	}
	return nil
}

func (s *MenuService) SubtractIngredientsByID(OrderID int, quantity int) error {
	if err := s.IngredientsCheckByID(OrderID, quantity); err != nil {
		return errors.New("not enough ingredients or needed ingredients do not exist")
	}

	ingredients := make(map[int]float64)
	menuItems, _ := s.menuRepo.GetAll()

	for _, item := range menuItems {
		if item.ID == OrderID {
			for _, ingr := range item.Ingredients {
				ingredients[ingr.IngredientID] += float64(ingr.Quantity) * float64(quantity)
			}
		}
	}

	return s.inventoryRepo.SubtractIngredients(ingredients)
}

func (s *MenuService) AddMenuItem(menuItem models.MenuItem) error {
	return s.menuRepo.AddMenuItemRepo(menuItem)
}

func (s *MenuService) GetMenuItem(MenuItemID int) (models.MenuItem, error) {
	MenuItems, err := s.menuRepo.GetAll()
	if err != nil {
		return models.MenuItem{}, err
	}
	for i, MenuItem := range MenuItems {
		if MenuItem.ID == MenuItemID {
			return MenuItems[i], nil
		}
	}
	return models.MenuItem{}, errors.New("could not find menu item by the given id")
}

func (s *MenuService) GetMenuItems() ([]models.MenuItem, error) {
	MenuItems, err := s.menuRepo.GetAll()
	if err != nil {
		return []models.MenuItem{}, err
	}
	return MenuItems, err
}

func (s *MenuService) CheckNewMenu(MenuItem models.MenuItem) error {
	// if strings.TrimSpace(MenuItem.ID) == "" {
	// 	return errors.New("new menu item's ID is empty")
	// }
	if strings.TrimSpace(MenuItem.Name) == "" {
		return errors.New("new menu item's Name is empty")
	}
	if strings.TrimSpace(MenuItem.Description) == "" {
		return errors.New("new menu item's Description is empty")
	}
	if MenuItem.Price < 0 {
		return errors.New("new menu item's Price is awkward")
	}
	for _, ingredient := range MenuItem.Ingredients {
		// if strings.TrimSpace(ingredient.IngredientID) == "" {
		// 	return errors.New("new menu item's ingredient is empty")
		// }
		if ingredient.Quantity < 0 {
			return errors.New("new menu item's quantity is awkward")
		}
	}
	return nil
}
