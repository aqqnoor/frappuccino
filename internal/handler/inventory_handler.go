package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/sunzhqr/frappuccino/internal/models"
	"github.com/sunzhqr/frappuccino/internal/service"
	"github.com/sunzhqr/frappuccino/pkg/response"
)

type InventoryHandler struct {
	inventoryService service.InventoryServiceInterface
	logger           *slog.Logger
}

func NewInventoryHandler(inventoryService service.InventoryServiceInterface, logger *slog.Logger) *InventoryHandler {
	return &InventoryHandler{inventoryService: inventoryService, logger: logger}
}

func decodeJSON(w http.ResponseWriter, r *http.Request, target interface{}) error {
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(target)
	if err != nil {
		response.SendError(w, "Could not decode request json data", http.StatusBadRequest)
		return err
	}
	return nil
}

func validateInventoryItem(item models.InventoryItem) error {
	if item.Name == "" || item.Unit == "" || item.Quantity <= 0 {
		return fmt.Errorf("some fields are empty or invalid")
	}
	return nil
}

func (h *InventoryHandler) handleError(w http.ResponseWriter, err error, message string, statusCode int) {
	h.logger.Error(message, "error", err)
	response.SendError(w, message, statusCode)
}

func (h *InventoryHandler) PostInventoryItem(w http.ResponseWriter, r *http.Request) {
	var newItem models.InventoryItem
	if err := decodeJSON(w, r, &newItem); err != nil {
		return
	}

	if err := validateInventoryItem(newItem); err != nil {
		h.handleError(w, err, "Some fields are empty or invalid", http.StatusBadRequest)
		return
	}

	err := h.inventoryService.AddInventoryItem(newItem)
	if err != nil {
		h.handleError(w, err, "Could not add new inventory item", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	response.SendSuccess(w, nil, "Inventory item created successfully", http.StatusCreated)
}

func (h *InventoryHandler) GetInventoryItems(w http.ResponseWriter, r *http.Request) {
	inventoryItems, err := h.inventoryService.GetAllInventoryItems()
	if err != nil {
		h.handleError(w, err, "Could not get inventory items", http.StatusInternalServerError)
		return
	}

	response.SendSuccess(w, inventoryItems, "Inventory items fetched successfully", http.StatusOK)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *InventoryHandler) GetInventoryItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.handleError(w, err, fmt.Sprint("Inventory id must be integer "+idStr), http.StatusBadRequest)
		return
	}

	if !h.inventoryService.Exists(id) {
		h.handleError(w, fmt.Errorf("inventory item does not exist"), "Inventory item does not exist", http.StatusNotFound)
		return
	}

	inventoryItem, err := h.inventoryService.GetItem(id)
	if err != nil {
		h.handleError(w, err, "Could not get inventory item", http.StatusInternalServerError)
		return
	}

	response.SendSuccess(w, inventoryItem, "Inventory item fetched successfully", http.StatusOK)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *InventoryHandler) PutInventoryItem(w http.ResponseWriter, r *http.Request) {
	var newItem models.InventoryItem
	if err := decodeJSON(w, r, &newItem); err != nil {
		return
	}

	if err := validateInventoryItem(newItem); err != nil {
		h.handleError(w, err, "Some fields are empty or invalid", http.StatusBadRequest)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.handleError(w, err, "Inventory id must be integer", http.StatusBadRequest)
		return
	}

	if !h.inventoryService.Exists(id) {
		h.handleError(w, fmt.Errorf("inventory item does not exist"), "Inventory item does not exist", http.StatusNotFound)
		return
	}

	err = h.inventoryService.UpdateItem(id, newItem)
	if err != nil {
		h.handleError(w, err, "Error updating inventory item", http.StatusInternalServerError)
		return
	}

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	response.SendSuccess(w, nil, "Inventory item updated successfully", http.StatusOK)
}

func (h *InventoryHandler) DeleteInventoryItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.handleError(w, err, "Inventory id must be integer", http.StatusBadRequest)
		return
	}

	if !h.inventoryService.Exists(id) {
		h.handleError(w, fmt.Errorf("inventory item does not exist"), "Inventory item does not exist", http.StatusNotFound)
		return
	}

	err = h.inventoryService.DeleteItem(id)
	if err != nil {
		h.handleError(w, err, "Could not delete inventory item", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *InventoryHandler) GetLeftOvers(w http.ResponseWriter, r *http.Request) {
	ParamSortBy := r.URL.Query().Get("sortBy")
	ParamPage := r.URL.Query().Get("page")
	ParamPageSize := r.URL.Query().Get("pageSize")

	resp, err := h.inventoryService.GetLeftOvers(ParamSortBy, ParamPage, ParamPageSize)
	if err != nil {
		h.handleError(w, err, fmt.Sprintf("Error %v", err), http.StatusBadRequest)
		return
	}

	response.SendSuccess(w, resp, "Leftovers fetched successfully", http.StatusOK)
}
