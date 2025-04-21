package service

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/sunzhqr/frappuccino/internal/models"
	"github.com/sunzhqr/frappuccino/internal/repository"
)

type OrderServiceInterface interface {
	AddOrder(order models.Order) (models.BatchOrderInfo, []models.BatchOrderInventoryUpdate, error)
	BulkOrders(orders []models.Order) (models.BatchOrdersResponce, error)
	GetAllOrders() ([]models.Order, error)
	GetOrder(OrderID int) (models.Order, error)
	UpdateOrder(updatedOrder models.Order, OrderID string) error
	GetTotalSales() (models.TotalSales, error)
	DeleteOrderByID(OrderID int) error
	CloseOrder(OrderID int) error
	GetNumberOfItems(startDate, endDate string) (map[string]int, error)
	GetOrderedItemsByPeriod(period, month, year string) (map[string]interface{}, error)
}

type OrderService struct {
	orderRepo     repository.OrderRepositoryInterface
	menuRepo      repository.MenuRepositoryInterface
	inventoryRepo repository.InventoryRepositoryInterface
}

func NewOrderService(orderRepo repository.OrderRepositoryInterface, menuRepo repository.MenuRepositoryInterface, inventoryRepo repository.InventoryRepositoryInterface) *OrderService {
	return &OrderService{
		orderRepo:     orderRepo,
		menuRepo:      menuRepo,
		inventoryRepo: inventoryRepo,
	}
}

// AddOrder adds a new order to the repository
func (s *OrderService) AddOrder(order models.Order) (models.BatchOrderInfo, []models.BatchOrderInventoryUpdate, error) {
	if err := validateOrder(order); err != nil {
		return models.BatchOrderInfo{
			OrderID:      order.ID,
			CustomerName: order.CustomerName,
			Status:       models.StatusOrderRejected,
			Reason:       err.Error(),
			Total:        0,
		}, []models.BatchOrderInventoryUpdate{}, err
	}

	return s.orderRepo.Add(order)
}

func (s *OrderService) BulkOrders(orders []models.Order) (models.BatchOrdersResponce, error) {
	proccesedOrdersInfo := []models.BatchOrderInfo{}
	summary := models.BatchOrderSummary{
		TotalOrders: len(orders),
	}

	invCheckMap := make(map[int]models.BatchOrderInventoryUpdate)
	for _, order := range orders {
		orderInfo, inventoryInfo, err := s.AddOrder(order)
		if err != nil {
			log.Printf("Error: %v", err)
		}

		// Counting accepted and rejected orders
		if orderInfo.Status == models.StatusOrderAccepted {
			summary.Accepted++
		} else {
			summary.Rejected++
		}
		summary.TotalRevenue += orderInfo.Total // Total revenue
		proccesedOrdersInfo = append(proccesedOrdersInfo, orderInfo)

		// summary.InventoryUpdates = append(summary.InventoryUpdates, inventoryInfo...)
		for _, v := range inventoryInfo {
			if value, ok := invCheckMap[v.IngredientID]; ok {
				v.Quantity_used += value.Quantity_used
				invCheckMap[v.IngredientID] = v
			} else {
				invCheckMap[v.IngredientID] = v
			}

		}

		// Closing the Order
		err = s.orderRepo.CloseOrderRepo(orderInfo.OrderID)
		if err != nil && err != models.ErrOrderNotFound {
			return models.BatchOrdersResponce{}, err
		}
	}

	// Filling InventoryUpdates by map
	for _, val := range invCheckMap {
		summary.InventoryUpdates = append(summary.InventoryUpdates, val)
	}

	result := models.BatchOrdersResponce{
		Processed_orders: proccesedOrdersInfo,
		Summary:          summary,
	}
	return result, nil
}

// GetAllOrders retrieves all orders from the repository
func (s *OrderService) GetAllOrders() ([]models.Order, error) {
	return s.orderRepo.GetAll()
}

func (s *OrderService) GetOrder(OrderID int) (models.Order, error) {
	return s.orderRepo.GetOrderByID(OrderID)
}

// UpdateOrder updates an existing order
func (s *OrderService) UpdateOrder(updatedOrder models.Order, OrderID string) error {
	if err := validateOrder(updatedOrder); err != nil {
		return err
	}
	return s.orderRepo.SaveUpdatedOrder(updatedOrder, OrderID)
}

func (s *OrderService) GetTotalSales() (models.TotalSales, error) {
	existingOrders, err := s.orderRepo.GetAll()
	if err != nil {
		return models.TotalSales{}, err
	}

	// Counting sales amount
	totalSales := models.TotalSales{}

	for _, order := range existingOrders {
		for _, item := range order.Items {
			totalSales.TotalSales += item.Quantity
		}
	}
	return totalSales, nil
}

func (s *OrderService) DeleteOrderByID(OrderID int) error {
	return s.orderRepo.DeleteOrder(OrderID)
}

func (s *OrderService) CloseOrder(OrderID int) error {
	return s.orderRepo.CloseOrderRepo(OrderID)
}

func (s *OrderService) GetNumberOfItems(startDate, endDate string) (map[string]int, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, fmt.Errorf("invalid time format of startDate")
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, fmt.Errorf("invalid time format of endDate")
	}

	return s.orderRepo.GetNumberOfItems(start, end)
}

func (s *OrderService) GetOrderedItemsByPeriod(period, month, year string) (map[string]interface{}, error) {
	if period == "" {
		return nil, fmt.Errorf("period is required")
	}

	if period == "day" {
		if month == "" {
			return nil, fmt.Errorf("period equal to 'day', but month not provided")
		}
	} else if period == "month" {
		if year == "" {
			return nil, fmt.Errorf("period equal to 'month', but year not provided")
		}
	} else {
		return nil, fmt.Errorf("invalid period value, must be 'day' or 'month'")
	}

	if period == "day" {
		monthInt := getMonthNumber(strings.ToLower(month))
		if monthInt == -1 {
			return nil, fmt.Errorf("%s, month does not exist", month)
		}

		var yearInt int
		if year == "" {
			yearInt = -1
		} else {
			n, err := strconv.Atoi(year)
			if err != nil {
				return nil, fmt.Errorf("year must be a number")
			}
			yearInt = n
		}

		return s.orderRepo.OrderedItemsByDay(monthInt, yearInt)
	} else if period == "month" {

		yearInt, err := strconv.Atoi(year)
		if err != nil {
			return nil, fmt.Errorf("year should be number")
		}
		return s.orderRepo.OrderedItemsByMonth(yearInt)
	}

	return nil, fmt.Errorf("invalid inputs. Period: %v, Month: %s, Year: %s", period, month, year)
}

func getMonthNumber(month string) int {
	months := map[string]int{
		"january":   1,
		"february":  2,
		"march":     3,
		"april":     4,
		"may":       5,
		"june":      6,
		"july":      7,
		"august":    8,
		"september": 9,
		"october":   10,
		"november":  11,
		"december":  12,
	}

	v, ok := months[strings.ToLower(month)]
	if !ok {
		return -1
	}
	return v
}

func validateOrder(order models.Order) error {
	if order.Items == nil {
		return errors.New("no items provided. Array of items it required")
	}

	if strings.TrimSpace(order.CustomerName) == "" {
		return errors.New("customer name is required")
	}
	for _, order := range order.Items {
		if order.Quantity < 1 {
			return errors.New("quantity a product must be greater than zero")
		}
	}

	return nil
}
