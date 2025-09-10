package handlers

import (
	"html/template"
	"net/http"

	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/internal/domain"
	"github.com/arajski/rate-limiting-with-istio/ratelimiting-service/pkg/service"
)

type HTTPHandlers struct {
	RateLimitService service.RateLimit
}

func (h *HTTPHandlers) Users(w http.ResponseWriter, r *http.Request) {
	tmpl := template.Must(template.ParseFiles("templates/users.html"))
	users, err := h.RateLimitService.UsersRepository.GetAllUsers()
	if err != nil {
		http.Error(w, "could not fetch users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl.Execute(w, users)
}

func (h *HTTPHandlers) RateLimits(w http.ResponseWriter, r *http.Request) {
	limits, err := h.RateLimitService.RateLimitsRepository.GetAllRateLimits()
	if err != nil {
		http.Error(w, "could not fetch rate limits: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/ratelimits.html"))
	tmpl.Execute(w, limits)
}

func (h *HTTPHandlers) NewRateLimit(w http.ResponseWriter, r *http.Request) {
	users, err := h.RateLimitService.UsersRepository.GetAllUsers()
	if err != nil {
		http.Error(w, "could not fetch users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := template.Must(template.ParseFiles("templates/ratelimits_new.html"))
	tmpl.Execute(w, users)
}

func (h *HTTPHandlers) CreateRateLimit(w http.ResponseWriter, r *http.Request) {
	userID := r.FormValue("user_id")
	endpoint := r.FormValue("endpoint")
	limitUnit := r.FormValue("limit_unit")
	limitInterval := r.FormValue("limit_interval")

	err := h.RateLimitService.RateLimitsRepository.CreateRateLimit(userID, endpoint, limitUnit, limitInterval)
	if err != nil {
		http.Error(w, "Failed to create rate limit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.RateLimitService.RefreshCache()
	if err != nil {
		http.Error(w, "could not refresh envoy cache"+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/ratelimits", http.StatusSeeOther)
}

func (h *HTTPHandlers) EditRateLimit(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	limit, err := h.RateLimitService.RateLimitsRepository.GetRateLimitByID(id)
	if err != nil {
		http.Error(w, "Rate limit not found: "+err.Error(), http.StatusNotFound)
		return
	}

	users, err := h.RateLimitService.UsersRepository.GetAllUsers()
	if err != nil {
		http.Error(w, "could not fetch users: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Users []domain.User
		Limit domain.RateLimit
	}{
		Users: users,
		Limit: *limit,
	}

	tmpl := template.Must(template.ParseFiles("templates/ratelimits_edit.html"))
	tmpl.Execute(w, data)
}

func (h *HTTPHandlers) UpdateRateLimit(w http.ResponseWriter, r *http.Request) {
	id := r.FormValue("id")
	userID := r.FormValue("user_id")
	endpoint := r.FormValue("endpoint")
	limitUnit := r.FormValue("limit_unit")
	limitInterval := r.FormValue("limit_interval")

	err := h.RateLimitService.RateLimitsRepository.UpdateRateLimit(id, userID, endpoint, limitUnit, limitInterval)
	if err != nil {
		http.Error(w, "failed to update rate limit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.RateLimitService.RefreshCache()
	if err != nil {
		http.Error(w, "could not refresh envoy cache"+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/ratelimits", http.StatusSeeOther)
}

func (h *HTTPHandlers) DeleteRateLimit(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	err := h.RateLimitService.RateLimitsRepository.DeleteRateLimit(id)
	if err != nil {
		http.Error(w, "failed to delete rate limit: "+err.Error(), http.StatusInternalServerError)
		return
	}

	err = h.RateLimitService.RefreshCache()
	if err != nil {
		http.Error(w, "could not refresh envoy cache"+err.Error(), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/ratelimits", http.StatusSeeOther)
}
