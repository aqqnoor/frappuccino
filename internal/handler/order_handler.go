package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/sunzhqr/frappuccino/internal/models"
	"github.com/sunzhqr/frappuccino/internal/service"
	"github.com/sunzhqr/frappuccino/pkg/response"
)

type OrderHandler struct {
	orderService service.OrderServiceInterface
	menuService  service.MenuServiceInterface
	logger       *slog.Logger
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(orderService service.OrderServiceInterface, menuService service.MenuServiceInterface, logger *slog.Logger) *OrderHandler {
	return &OrderHandler{orderService: orderService, menuService: menuService, logger: logger}
}

// PostOrder creates new Order
func (h *OrderHandler) PostOrder(w http.ResponseWriter, r *http.Request) {
	var NewOrder models.Order
	err := json.NewDecoder(r.Body).Decode(&NewOrder)
	if err != nil {
		h.logger.Error("Could not decode request json data", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not decode request json adata", http.StatusBadRequest)
		return
	}

	for _, OrderItem := range NewOrder.Items {
		if err = h.menuService.MenuCheckByID(OrderItem.ProductID, true); err != nil {
			h.logger.Error("Requested order item does not exist in menu", "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, "Requested order item does not exist in menu", http.StatusBadRequest)
			return
		}
		if err = h.menuService.IngredientsCheckByID(OrderItem.ProductID, OrderItem.Quantity); err != nil {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}

	_, _, err = h.orderService.AddOrder(NewOrder)
	if err != nil {
		if err.Error() == "something wrong with your requested order" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, "Something wrong when adding new order", http.StatusInternalServerError)
			return
		}
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.WriteHeader(http.StatusCreated)
}

func (h *OrderHandler) GetOrders(w http.ResponseWriter, r *http.Request) {
	Orders, err := h.orderService.GetAllOrders()
	if err != nil {
		h.logger.Error("Can not read order data from server", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Can not read order data from server", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.MarshalIndent(Orders, "", "    ")
	if err != nil {
		h.logger.Error("Can not convert order data to json", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Can not convert order data to json", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonData)
}

func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.SendError(w, "The id should be positive integer", http.StatusBadRequest)
		h.logger.Error("The id should be positive integer", "method", r.Method, "url", r.URL)
		return
	}
	RequestedOrder, err := h.orderService.GetOrder(ID)
	if err != nil {
		if err.Error() == "the order with given ID soes not exist" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusNotFound)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	jsonData, err := json.MarshalIndent(RequestedOrder, "", "    ")
	if err != nil {
		h.logger.Error("Can not convert order data to json", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Can not convert order data to json", http.StatusInternalServerError)
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(200)
	w.Write(jsonData)
}

func (h *OrderHandler) PutOrder(w http.ResponseWriter, r *http.Request) {
	var RequestedOrder models.Order
	err := json.NewDecoder(r.Body).Decode(&RequestedOrder)
	if err != nil {
		h.logger.Error("Could not decode request json data", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not decode request json data", http.StatusBadRequest)
		return
	}

	for _, OrderItem := range RequestedOrder.Items {
		if err = h.menuService.MenuCheckByID(OrderItem.ProductID, true); err != nil {
			h.logger.Error("Updated order item does not exist in menu", "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, "Updated order item does not exist in menu", http.StatusBadRequest)
			return
		}
		if err = h.menuService.IngredientsCheckByID(OrderItem.ProductID, OrderItem.Quantity); err != nil {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		}
	}
	err = h.orderService.UpdateOrder(RequestedOrder, r.PathValue("id"))
	if err != nil {
		if err.Error() == "could not update the order because it is already closed" || err.Error() == "something wrong with your updated order" || err.Error() == "the requested order does not exist" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.WriteHeader(200)
}

func (h *OrderHandler) DeleteOrder(w http.ResponseWriter, r *http.Request) {
	ID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		response.SendError(w, "The id should be positive integer", http.StatusBadRequest)
		h.logger.Error("The id should be positive integer", "method", r.Method, "url", r.URL)
		return
	}
	err = h.orderService.DeleteOrderByID(ID)
	if err != nil {
		if err.Error() == "the order with given ID does not exist" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, "Error updating orders database", http.StatusInternalServerError)
			return
		}
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.WriteHeader(204)
}

func (h *OrderHandler) CloseOrder(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	ID, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Order id must be integer", "method", r.Method, "url", r.URL)
		response.SendError(w, "Order id must be integer", http.StatusBadRequest)
		return
	}

	err = h.orderService.CloseOrder(ID)
	if err != nil {
		h.logger.Error("Error closing order", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *OrderHandler) GetNumberOfOrdered(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("startDate")
	endDate := r.URL.Query().Get("endDate")
	if startDate == "" {
		startDate = "1970-01-01"
	}
	if endDate == "" {
		endDate = time.Now().Format("2006-01-02")
	}
	items, err := h.orderService.GetNumberOfItems(startDate, endDate)
	if err != nil {
		h.logger.Error(err.Error(), "query", r.URL.Query, "error", err)
		response.SendError(w, fmt.Sprintf("Error getting number of ordered items. Error:%v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(items); err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not encode request json data", http.StatusInternalServerError)
	}
}

/*
POST /orders/batch-process:
Process multiple orders simultaneously while ensuring inventory consistency.
This endpoint must handle concurrent orders and maintain data integrity using transactions.
*/
func (h *OrderHandler) BatchOrders(w http.ResponseWriter, r *http.Request) {
	request := struct {
		Orders []models.Order `json:"orders"`
	}{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		h.logger.Error("Could not decode request json data", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not decode request json data", http.StatusBadRequest)
		return
	}

	ordersReport, err := h.orderService.BulkOrders(request.Orders)
	if err != nil {
		h.logger.Error("Error proccing orders", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Error proccing orders. "+err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(ordersReport); err != nil {
		h.logger.Error("Error encoding responce", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Error encoding responce", http.StatusInternalServerError)
	}
}
