package service

import (
	"errors"
	"strconv"

	"github.com/sunzhqr/frappuccino/internal/models"
	"github.com/sunzhqr/frappuccino/internal/repository"
)

type InventoryServiceInterface interface {
	AddInventoryItem(item models.InventoryItem) error
	GetAllInventoryItems() ([]models.InventoryItem, error)
	GetItem(id int) (models.InventoryItem, error)
	UpdateItem(id int, newItem models.InventoryItem) error
	DeleteItem(id int) error
	Exists(id int) bool
	GetLeftOvers(sortBy, page, pageSize string) (map[string]any, error)
}

type InventoryService struct {
	inventoryRepository repository.InventoryRepositoryInterface
}

func NewInventoryService(inventoryRepository repository.InventoryRepositoryInterface) *InventoryService {
	return &InventoryService{inventoryRepository: inventoryRepository}
}

func (s *InventoryService) AddInventoryItem(item models.InventoryItem) error {
	return s.inventoryRepository.AddInventoryItemRepo(item)
}

func (s *InventoryService) GetAllInventoryItems() ([]models.InventoryItem, error) {
	items, err := s.inventoryRepository.GetAll()
	if err != nil {
		return nil, err
	}
	return items, nil
}

func (s *InventoryService) GetItem(id int) (models.InventoryItem, error) {
	inventoryItems, err := s.inventoryRepository.GetAll()
	if err != nil {
		return models.InventoryItem{}, err
	}

	for _, inventoryItem := range inventoryItems {
		if inventoryItem.IngredientID == id {
			return inventoryItem, nil
		}
	}
	return models.InventoryItem{}, errors.New("inventory item does not exist")
}

func (s *InventoryService) UpdateItem(id int, newItem models.InventoryItem) error {
	if !s.inventoryRepository.Exists(id) {
		return errors.New("inventory item does not exist")
	}
	return s.inventoryRepository.UpdateItemRepo(id, newItem)
}

func (s *InventoryService) DeleteItem(id int) error {
	if !s.inventoryRepository.Exists(id) {
		return errors.New("inventory item does not exist")
	}
	return s.inventoryRepository.DeleteItemRepo(id)
}

func (s *InventoryService) Exists(id int) bool {
	return s.inventoryRepository.Exists(id)
}

func (s *InventoryService) GetLeftOvers(sortBy, page, pageSize string) (map[string]any, error) {
	if sortBy == "" {
		sortBy = "price"
	}
	if sortBy != "price" && sortBy != "quantity" {
		return nil, errors.New("invalid sortBy value, must be 'price' or 'quantity'")
	}

	if page == "" {
		page = "1"
	}
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum <= 0 {
		return nil, errors.New("invalid page parameter. page must be a positive integer greater than 0")
	}

	if pageSize == "" {
		pageSize = "10"
	}
	pageSizeNum, err := strconv.Atoi(pageSize)
	if err != nil || pageSizeNum <= 0 {
		return nil, errors.New("invalid pageSize parameter. pageSize must be a positive integer greater than 0")
	}

	return s.inventoryRepository.GetLeftOvers(sortBy, page, pageSize)
}
