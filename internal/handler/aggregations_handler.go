package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/sunzhqr/frappuccino/internal/service"
	"github.com/sunzhqr/frappuccino/pkg/response"
)

type AggregationHandler struct {
	orderService       service.OrderServiceInterface
	aggregationService service.AggregationServiceInterface
	logger             *slog.Logger
}

func NewAggregationHandler(orderService service.OrderServiceInterface, aggregationService service.AggregationServiceInterface, logger *slog.Logger) *AggregationHandler {
	return &AggregationHandler{orderService: orderService, aggregationService: aggregationService, logger: logger}
}

// Return all saled items as key and quantity as value in JSON
func (h *AggregationHandler) TotalSalesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.SendError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	totalSales, err := h.orderService.GetTotalSales()
	if err != nil {
		h.logger.Error("Error getting data", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Error getting data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(totalSales)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

// Returns Each item as key and quatity as value
func (h *AggregationHandler) PopularItemsHandler(w http.ResponseWriter, r *http.Request) {
	popularItems, err := h.aggregationService.GetPopularMenuItems()
	if err != nil {
		h.logger.Error("Error getting popular items", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Error getting popular items", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(popularItems)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *AggregationHandler) SearchHandler(w http.ResponseWriter, r *http.Request) {
	searchQuery := r.URL.Query().Get("q")
	filter := r.URL.Query().Get("filter")
	minPrice := r.URL.Query().Get("minPrice")
	maxPrice := r.URL.Query().Get("maxPrice")

	var MinPrice int
	var err error
	if minPrice != "" {
		MinPrice, err = strconv.Atoi(minPrice)
		if err != nil {
			h.logger.Error("Min Price should be number", "method", r.Method, "url", r.URL)
			response.SendError(w, "Min Price should be number", http.StatusBadRequest)
			return
		}
		if MinPrice < 0 {
			h.logger.Error("Min Price should be positive", "method", r.Method, "url", r.URL)
			response.SendError(w, service.ErrPriceNotPositive.Error(), http.StatusBadRequest)
			return
		}
	} else {
		MinPrice = -1
	}

	var MaxPrice int
	if maxPrice != "" {
		MaxPrice, err = strconv.Atoi(maxPrice)
		if err != nil {
			h.logger.Error("Max Price should be number", "method", r.Method, "url", r.URL)
			response.SendError(w, "Max Price should be number", http.StatusBadRequest)
			return
		}

		if MaxPrice < 0 {
			h.logger.Error("Max Price should be positive", "method", r.Method, "url", r.URL)
			response.SendError(w, service.ErrPriceNotPositive.Error(), http.StatusBadRequest)
			return
		}
	} else {
		MaxPrice = -1
	}

	searchResult, err := h.aggregationService.Search(searchQuery, MinPrice, MaxPrice, filter)
	if err != nil {
		h.logger.Error("Error searching", "method", r.Method, "url", r.URL, "err", err.Error())
		if err == service.ErrSearchRequired || err == service.ErrWrongFilterOptions || err == service.ErrPriceNotPositive {
			response.SendError(w, err.Error(), http.StatusBadRequest)
		} else {
			response.SendError(w, fmt.Sprintf("Error searching %v string", searchQuery), http.StatusInternalServerError)
			return
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(w).Encode(searchResult); err != nil {
		h.logger.Error("Could not encode json data", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not encode request json data", http.StatusInternalServerError)
	}
}

/*
	3
	GET /reports/orderedItemsByPeriod?period={day|month}&month={month}: Returns the number of orders for the specified period, grouped by day within a month or by month within a year. The period parameter can take the value day or month. The month parameter is optional and used only when period=day.

##### Parameters:

	period (required):
	    day: Groups data by day within the specified month.
	    month: Groups data by month within the specified year.
	month (optional): Specifies the month (e.g., october). Used only if period=day.
	year (optional): Specifies the year. Used only if period=month.
*/
func (h *AggregationHandler) OrderByPeriod(w http.ResponseWriter, r *http.Request) {
	period := r.URL.Query().Get("period")
	month := r.URL.Query().Get("month")
	year := r.URL.Query().Get("year")

	orders, err := h.orderService.GetOrderedItemsByPeriod(period, month, year)
	if err != nil {
		h.logger.Error(err.Error(), "msg", "Error getting orders by time period", "url", r.URL)
		response.SendError(w, fmt.Sprintf("Error getting orders by time period. %v", err), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(orders)
	if err != nil {
		h.logger.Error(err.Error(), "msg", "Failed to encode orders", "url", r.URL)
		response.SendError(w, "Failed to encode orders", http.StatusInternalServerError)
		return
	}
}
