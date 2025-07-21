package handler

import (
	"net/http"

	"github.com/naoyafurudono/dirty/example/crosspackage/service"
)

// APIHandler handles HTTP requests
type APIHandler struct {
	userService *service.UserService
}

// dirty: { select[users] | http_response }
func (h *APIHandler) GetUserHandler(w http.ResponseWriter, r *http.Request) {
	// Correctly includes the transitive effect from db package
	user, err := h.userService.GetUser(123)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Write response (simulated)
	w.Write([]byte(user.Name))
}

// This handler is missing required effects
// Should declare { select[users] | update[users] | http_response }
// dirty: { http_response }
func (h *APIHandler) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
	// ERROR: Missing { select[users] | update[users] } effects
	err := h.userService.UpdateUserEmailIfExists(123, "new@example.com")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
