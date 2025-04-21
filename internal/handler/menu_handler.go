package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/sunzhqr/frappuccino/internal/models"
	"github.com/sunzhqr/frappuccino/internal/service"
	"github.com/sunzhqr/frappuccino/pkg/response"
)

type MenuHandler struct {
	menuService service.MenuServiceInterface
	logger      *slog.Logger
}

func NewMenuHandler(menuService service.MenuServiceInterface, logger *slog.Logger) *MenuHandler {
	return &MenuHandler{menuService: menuService, logger: logger}
}

func (h *MenuHandler) PostMenuItem(w http.ResponseWriter, r *http.Request) {
	var newItem models.MenuItem
	err := json.NewDecoder(r.Body).Decode(&newItem)
	if err != nil {
		h.logger.Error("Could not decode request json data", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not decode request json data", http.StatusBadRequest)
		return
	}
	if err = h.menuService.CheckNewMenu(newItem); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Use the service to check if the item already exists
	if err = h.menuService.MenuCheckByID(newItem.ID, false); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = h.menuService.IngredientsCheckForNewItem(newItem); err != nil {
		h.logger.Error(err.Error(), "method", r.Method, "url", r.URL, "error", err)
		response.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Add the new menu item using the service
	if err = h.menuService.AddMenuItem(newItem); err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not add menu item", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *MenuHandler) GetMenuItems(w http.ResponseWriter, r *http.Request) {
	MenuItems, err := h.menuService.GetMenuItems()
	if err != nil {
		h.logger.Error("Could not read menu database", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not read menu database", http.StatusInternalServerError)
		return
	}
	jsonData, err := json.MarshalIndent(MenuItems, "", "    ")
	if err != nil {
		h.logger.Error("Could not read menu database", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not read menu items", http.StatusInternalServerError)
		return
	}
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func (h *MenuHandler) GetMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		response.SendError(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	MenuItem, err := h.menuService.GetMenuItem(id)
	if err != nil {
		if err.Error() == "could not find menu item by the given id" {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, err.Error(), http.StatusBadRequest)
			return
		} else {
			h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
			response.SendError(w, "Could not read menu database", http.StatusInternalServerError)
			return
		}
	}
	jsonData, err := json.MarshalIndent(MenuItem, "", "    ")
	if err != nil {
		h.logger.Error("Could not convert Menu Items to jsondata", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not send menu item", http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)

	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *MenuHandler) PutMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		response.SendError(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	err = h.menuService.MenuCheckByID(id, true)
	if err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var RequestedMenuItem models.MenuItem
	err = json.NewDecoder(r.Body).Decode(&RequestedMenuItem)
	if err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not read requested menu item", http.StatusInternalServerError)
		return
	}
	RequestedMenuItem.ID = id

	err = h.menuService.CheckNewMenu(RequestedMenuItem)
	if err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err = h.menuService.IngredientsCheckForNewItem(RequestedMenuItem); err != nil {
		h.logger.Error(err.Error(), "method", r.Method, "url", r.URL)
		response.SendError(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = h.menuService.UpdateMenuItem(RequestedMenuItem)
	if err != nil {
		h.logger.Error(err.Error(), "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not update menu database", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(201)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}

func (h *MenuHandler) DeleteMenuItem(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		h.logger.Error("Menu id must be integer", "method", r.Method, "url", r.URL)
		response.SendError(w, "Menu id must be integer", http.StatusBadRequest)
		return
	}

	err = h.menuService.DeleteMenuItem(id)
	if err != nil {
		h.logger.Error("Could not delete menu item", "error", err, "method", r.Method, "url", r.URL)
		response.SendError(w, "Could not delete menu item", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(204)
	h.logger.Info("Request handled successfully.", "method", r.Method, "url", r.URL)
}
