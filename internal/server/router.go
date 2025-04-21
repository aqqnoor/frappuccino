package server

import (
	"database/sql"
	"log/slog"
	"net/http"

	"github.com/sunzhqr/frappuccino/internal/handler"
	"github.com/sunzhqr/frappuccino/internal/repository"
	"github.com/sunzhqr/frappuccino/internal/service"
)

func setupRoutes(router *http.ServeMux, db *sql.DB, logger *slog.Logger) {
	// Inventory
	inventoryRepo := repository.NewInventoryRepository(db)
	inventoryService := service.NewInventoryService(inventoryRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryService, logger)

	// Menu
	menuRepo := repository.NewMenuRepository(db)
	menuService := service.NewMenuService(menuRepo, inventoryRepo)
	menuHandler := handler.NewMenuHandler(menuService, logger)

	// Order
	orderRepo := repository.NewOrderRepository(db)
	orderService := service.NewOrderService(orderRepo, menuRepo, inventoryRepo)
	orderHandler := handler.NewOrderHandler(orderService, menuService, logger)

	// Aggregation
	aggregationRepo := repository.NewReportRespository(db)
	aggregationService := service.NewAggregationService(aggregationRepo)
	aggregationHandler := handler.NewAggregationHandler(orderService, aggregationService, logger)

	// Inventory Routes
	router.HandleFunc("POST /inventory", inventoryHandler.PostInventoryItem)
	router.HandleFunc("GET /inventory", inventoryHandler.GetInventoryItems)
	router.HandleFunc("GET /inventory/{id}", inventoryHandler.GetInventoryItem)
	router.HandleFunc("PUT /inventory/{id}", inventoryHandler.PutInventoryItem)
	router.HandleFunc("DELETE /inventory/{id}", inventoryHandler.DeleteInventoryItem)
	router.HandleFunc("GET /inventory/getLeftOvers", inventoryHandler.GetLeftOvers)

	// Menu Routes
	router.HandleFunc("POST /menu", menuHandler.PostMenuItem)
	router.HandleFunc("GET /menu", menuHandler.GetMenuItems)
	router.HandleFunc("GET /menu/{id}", menuHandler.GetMenuItem)
	router.HandleFunc("PUT /menu/{id}", menuHandler.PutMenuItem)
	router.HandleFunc("DELETE /menu/{id}", menuHandler.DeleteMenuItem)

	// Order routes
	router.HandleFunc("POST /orders", orderHandler.PostOrder)
	router.HandleFunc("GET /orders", orderHandler.GetOrders)
	router.HandleFunc("GET /orders/{id}", orderHandler.GetOrder)
	router.HandleFunc("PUT /orders/{id}", orderHandler.PutOrder)
	router.HandleFunc("DELETE /orders/{id}", orderHandler.DeleteOrder)
	router.HandleFunc("POST /orders/{id}/close", orderHandler.CloseOrder)
	router.HandleFunc("GET /orders/numberOfOrderedItems", orderHandler.GetNumberOfOrdered)
	router.HandleFunc("POST /orders/batch-process", orderHandler.BatchOrders)

	// Report routes
	router.HandleFunc("GET /reports/total-sales", aggregationHandler.TotalSalesHandler)
	router.HandleFunc("GET /reports/popular-items", aggregationHandler.PopularItemsHandler)
	router.HandleFunc("GET /reports/orderedItemsByPeriod", aggregationHandler.OrderByPeriod)
	router.HandleFunc("GET /reports/search", aggregationHandler.SearchHandler)
}
