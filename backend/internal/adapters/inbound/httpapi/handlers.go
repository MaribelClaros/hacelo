package httpapi

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"hacelo/internal/application/port"
	"hacelo/internal/domain"
)

// Handler translates HTTP requests into calls to the inbound ports.
type Handler struct {
	tasks          port.TaskUseCases
	subscriptions  port.SubscriptionUseCases
	vapidPublicKey string
	logger         *slog.Logger
}

func NewHandler(
	tasks port.TaskUseCases,
	subscriptions port.SubscriptionUseCases,
	vapidPublicKey string,
	logger *slog.Logger,
) *Handler {
	return &Handler{tasks: tasks, subscriptions: subscriptions, vapidPublicKey: vapidPublicKey, logger: logger}
}

// registerRoutes mounts the /api endpoints. It relies on the path
// parameters of the Go 1.22+ net/http router, so no external router is needed.
func (h *Handler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/health", h.health)

	mux.HandleFunc("GET /api/tasks", h.listTasks)
	mux.HandleFunc("POST /api/tasks", h.createTask)
	mux.HandleFunc("GET /api/tasks/{id}", h.getTask)
	mux.HandleFunc("PATCH /api/tasks/{id}", h.updateTask)
	mux.HandleFunc("DELETE /api/tasks/{id}", h.deleteTask)
	mux.HandleFunc("POST /api/tasks/{id}/actions/snooze", h.snoozeTask)
	mux.HandleFunc("POST /api/tasks/{id}/actions/done", h.completeTask)

	mux.HandleFunc("GET /api/push/vapid-public-key", h.getVapidPublicKey)
	mux.HandleFunc("POST /api/push/subscribe", h.subscribe)
	mux.HandleFunc("DELETE /api/push/subscribe", h.unsubscribe)
}

func (h *Handler) health(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (h *Handler) listTasks(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := port.TaskFilter{
		Date:   query.Get("date"),
		Status: domain.Status(query.Get("status")),
	}
	if filter.Date != "" {
		if _, err := time.Parse("2006-01-02", filter.Date); err != nil {
			h.writeError(w, fmt.Errorf("%w: date must be YYYY-MM-DD", domain.ErrInvalid))
			return
		}
	}

	tasks, err := h.tasks.ListTasks(r.Context(), filter)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponses(tasks))
}

func (h *Handler) createTask(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, fmt.Errorf("%w: malformed body", domain.ErrInvalid))
		return
	}
	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		h.writeError(w, fmt.Errorf("%w: scheduled_at must be RFC 3339", domain.ErrInvalid))
		return
	}

	task, err := h.tasks.CreateTask(r.Context(), port.CreateTaskCommand{
		Description:   req.Description,
		Urgency:       domain.Urgency(req.Urgency),
		ScheduledAt:   scheduledAt,
		SnoozeMinutes: req.SnoozeMinutes,
	})
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, toTaskResponse(task))
}

func (h *Handler) getTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	task, err := h.tasks.GetTask(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

func (h *Handler) updateTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	var req updateTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, fmt.Errorf("%w: malformed body", domain.ErrInvalid))
		return
	}

	cmd := port.UpdateTaskCommand{
		Description:   req.Description,
		SnoozeMinutes: req.SnoozeMinutes,
	}
	if req.Urgency != nil {
		urgency := domain.Urgency(*req.Urgency)
		cmd.Urgency = &urgency
	}
	if req.NextReminderAt != nil {
		at, err := time.Parse(time.RFC3339, *req.NextReminderAt)
		if err != nil {
			h.writeError(w, fmt.Errorf("%w: next_reminder_at must be RFC 3339", domain.ErrInvalid))
			return
		}
		cmd.NextReminderAt = &at
	}

	task, err := h.tasks.UpdateTask(r.Context(), id, cmd)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

func (h *Handler) deleteTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	if err := h.tasks.DeleteTask(r.Context(), id); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// snoozeTask serves the "I'm on it" notification button.
func (h *Handler) snoozeTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	task, err := h.tasks.SnoozeTask(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

// completeTask serves the "I did it" notification button.
func (h *Handler) completeTask(w http.ResponseWriter, r *http.Request) {
	id, ok := h.pathID(w, r)
	if !ok {
		return
	}
	task, err := h.tasks.CompleteTask(r.Context(), id)
	if err != nil {
		h.writeError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, toTaskResponse(task))
}

func (h *Handler) getVapidPublicKey(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"public_key": h.vapidPublicKey})
}

func (h *Handler) subscribe(w http.ResponseWriter, r *http.Request) {
	var req subscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, fmt.Errorf("%w: malformed body", domain.ErrInvalid))
		return
	}
	err := h.subscriptions.Subscribe(r.Context(), domain.PushSubscription{
		Endpoint:  req.Endpoint,
		P256dhKey: req.Keys.P256dh,
		AuthKey:   req.Keys.Auth,
	})
	if err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) unsubscribe(w http.ResponseWriter, r *http.Request) {
	var req unsubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, fmt.Errorf("%w: malformed body", domain.ErrInvalid))
		return
	}
	if err := h.subscriptions.Unsubscribe(r.Context(), req.Endpoint); err != nil {
		h.writeError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) pathID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		h.writeError(w, fmt.Errorf("%w: invalid id", domain.ErrInvalid))
		return 0, false
	}
	return id, true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// writeError maps domain errors to HTTP status codes. Unexpected errors are
// logged and reported as a generic 500 so internals are not leaked.
func (h *Handler) writeError(w http.ResponseWriter, err error) {
	status := http.StatusInternalServerError
	message := "internal error"

	switch {
	case errors.Is(err, domain.ErrNotFound):
		status, message = http.StatusNotFound, err.Error()
	case errors.Is(err, domain.ErrInvalid):
		status, message = http.StatusBadRequest, err.Error()
	case errors.Is(err, domain.ErrAlreadyDone):
		status, message = http.StatusConflict, err.Error()
	default:
		h.logger.Error("unhandled error", "error", err)
	}
	writeJSON(w, status, map[string]string{"error": message})
}
