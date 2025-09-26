package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"xpired/internal/auth"
	"xpired/internal/db"
)

type Handler struct {
	repo db.Repository
}

func NewHandler(repo db.Repository) *Handler {
	return &Handler{
		repo: repo,
	}
}

func (h *Handler) HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":    "ok",
		"service":   "xpired-api",
		"timestamp": time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) RegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req UserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	if err := h.repo.CheckUserExistsByEmail(r.Context(), req.Email); err == nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	newUser := &db.User{
		ID:          uuid.New(),
		Email:       req.Email,
		Password:    string(hashedPassword),
		Name:        req.Name,
		PhoneNumber: req.PhoneNumber,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	if err := h.repo.CreateUser(r.Context(), newUser); err != nil {
		http.Error(w, "Failed to create user", http.StatusInternalServerError)
		return
	}

	token, err := auth.GenerateToken(newUser.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // TODO: change to true in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	userResp := &UserResponse{
		ID:          newUser.ID.String(),
		Email:       newUser.Email,
		Name:        newUser.Name,
		PhoneNumber: newUser.PhoneNumber,
	}

	resp := map[string]interface{}{
		"message": "User registered successfully",
		"user":    userResp,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    token,
		HttpOnly: true,
		Path:     "/",
		Secure:   false, // TODO: change to true in production
		SameSite: http.SameSiteStrictMode,
		MaxAge:   86400,
	})

	userResp := &UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Name:        user.Name,
		PhoneNumber: user.PhoneNumber,
	}

	resp := map[string]interface{}{
		"message": "User login successful",
		"user":    userResp,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	userResp := &UserResponse{
		ID:          user.ID.String(),
		Email:       user.Email,
		Name:        user.Name,
		PhoneNumber: user.PhoneNumber,
	}

	resp := map[string]interface{}{
		"message": "User Profile",
		"user":    userResp,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     "auth",
		Value:    "",
		HttpOnly: true,
		Secure:   false, // TODO: change to true in production
		SameSite: http.SameSiteNoneMode,
		MaxAge:   0,
		Path:     "/",
	})

	resp := map[string]interface{}{
		"message": "Logout Successful",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	documents, err := h.repo.ListDocumentsByUserID(r.Context(), userID)
	if err != nil {
		http.Error(w, "Failed to fetch documents", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"message":   "List of Documents",
		"documents": documents,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var req DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	if req.Name == "" || req.ExpirationDate.IsZero() || req.Timezone == "" {
		http.Error(w, "Missing required fields", http.StatusBadRequest)
		return
	}

	newDoc := &db.Document{
		ID:             uuid.New(),
		UserID:         uuid.MustParse(userID),
		Name:           req.Name,
		Description:    req.Description,
		Identifier:     req.Identifier,
		ExpirationDate: req.ExpirationDate,
		Timezone:       req.Timezone,
		AttachmentURL:  req.AttachmentURL,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	err = h.repo.CreateDocument(r.Context(), newDoc)
	if err != nil {
		http.Error(w, "Failed to create document", http.StatusInternalServerError)
		return
	}

	reminderIntervals, err := h.repo.GetReminderIntervalsFromIdLabels(r.Context(), req.Reminders)
	if err != nil {
		http.Error(w, "Failed to fetch reminder intervals", http.StatusInternalServerError)
		return
	}

	var reminders []ReminderIntervalResponse
	for _, interval := range reminderIntervals {
		rmInterval := ReminderIntervalResponse{
			ID:    interval.IdLabel,
			Label: interval.Label,
		}
		reminders = append(reminders, rmInterval)
		docReminder := &db.DocumentReminder{
			ID:                 uuid.New(),
			DocumentID:         newDoc.ID.String(),
			ReminderIntervalID: interval.ID,
			Enabled:            true,
			SentAt:             nil,
		}
		err = h.repo.SetDocumentReminders(r.Context(), newDoc.ID.String(), docReminder)
		if err != nil {
			http.Error(w, "Failed to set document reminders", http.StatusInternalServerError)
			return
		}
	}

	expiryDateTime := newDoc.ExpirationDate.In(time.FixedZone(newDoc.Timezone, 0))
	expiryDate := time.Date(expiryDateTime.Year(), expiryDateTime.Month(), expiryDateTime.Day(), 0, 0, 0, 0, expiryDateTime.Location())

	doc := &DocumentResponse{
		ID:             newDoc.ID.String(),
		UserID:         newDoc.UserID.String(),
		Name:           newDoc.Name,
		Description:    newDoc.Description,
		Identifier:     newDoc.Identifier,
		ExpirationDate: expiryDate.Format("Mon, 2 Jan, 2006"),
		Timezone:       newDoc.Timezone,
		AttachmentURL:  newDoc.AttachmentURL,
		Reminders:      reminders,
		CreatedAt:      newDoc.CreatedAt,
		UpdatedAt:      newDoc.UpdatedAt,
	}

	resp := map[string]interface{}{
		"message":  "Document created successfully",
		"document": doc,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) GetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if doc.UserID.String() != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	reminders, err := h.repo.GetDocumentRemindersByDocumentID(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Failed to fetch document reminders", http.StatusInternalServerError)
		return
	}

	var rems []ReminderIntervalResponse
	for _, reminder := range reminders {
		interval, err := h.repo.GetReminderIntervalByID(r.Context(), reminder.ReminderIntervalID)
		if err == nil {
			remInterval := ReminderIntervalResponse{
				ID:    interval.IdLabel,
				Label: interval.Label,
			}
			rems = append(rems, remInterval)
		}
	}

	docResp := &DocumentResponse{
		ID:             doc.ID.String(),
		UserID:         doc.UserID.String(),
		Name:           doc.Name,
		Description:    doc.Description,
		Identifier:     doc.Identifier,
		ExpirationDate: doc.ExpirationDate.Format("Mon, 2 Jan, 2006"),
		Timezone:       doc.Timezone,
		AttachmentURL:  doc.AttachmentURL,
		Reminders:      rems,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
	}

	resp := map[string]interface{}{
		"message":  "Document fetched successfully",
		"document": docResp,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if doc.UserID.String() != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	var req DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Name != "" {
		doc.Name = req.Name
	}
	if req.Description != nil {
		doc.Description = req.Description
	}
	if req.Identifier != nil {
		doc.Identifier = req.Identifier
	}
	if !req.ExpirationDate.IsZero() {
		doc.ExpirationDate = req.ExpirationDate
	}
	if req.Timezone != "" {
		doc.Timezone = req.Timezone
	}
	if req.AttachmentURL != nil {
		doc.AttachmentURL = req.AttachmentURL
	}
	doc.UpdatedAt = time.Now()

	err = h.repo.UpdateDocument(r.Context(), doc)
	if err != nil {
		http.Error(w, "Failed to update document", http.StatusInternalServerError)
		return
	}

	reminderIntervals, err := h.repo.GetReminderIntervalsFromIdLabels(r.Context(), req.Reminders)
	if err != nil {
		http.Error(w, "Failed to fetch reminder intervals", http.StatusInternalServerError)
		return
	}

	var reminders []ReminderIntervalResponse
	for _, interval := range reminderIntervals {
		reminderInterval := ReminderIntervalResponse{
			ID:    interval.IdLabel,
			Label: interval.Label,
		}
		reminders = append(reminders, reminderInterval)
		docReminder := &db.DocumentReminder{
			ID:                 uuid.New(),
			DocumentID:         doc.ID.String(),
			ReminderIntervalID: interval.ID,
			Enabled:            true,
			SentAt:             nil,
		}
		err = h.repo.SetDocumentReminders(r.Context(), doc.ID.String(), docReminder)
		if err != nil {
			http.Error(w, "Failed to set document reminders", http.StatusInternalServerError)
			return
		}
	}

	updatedDoc := &DocumentResponse{
		ID:             doc.ID.String(),
		UserID:         doc.UserID.String(),
		Name:           doc.Name,
		Description:    doc.Description,
		Identifier:     doc.Identifier,
		ExpirationDate: doc.ExpirationDate.Format("Mon, 2 Jan, 2006"),
		Timezone:       doc.Timezone,
		AttachmentURL:  doc.AttachmentURL,
		Reminders:      reminders,
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
	}

	resp := map[string]interface{}{
		"message":  "Document updated successfully",
		"document": updatedDoc,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if doc.UserID.String() != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}
	err = h.repo.DeleteDocument(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Failed to delete document", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetReminderIntervalsHandler(w http.ResponseWriter, r *http.Request) {
	intervals, err := h.repo.GetAllReminderIntervals(r.Context())
	if err != nil {
		http.Error(w, "Failed to fetch reminder intervals", http.StatusInternalServerError)
		return
	}

	var respIntervals []ReminderIntervalResponse
	for _, interval := range intervals {
		respInterval := ReminderIntervalResponse{
			ID:    interval.IdLabel,
			Label: interval.Label,
		}
		respIntervals = append(respIntervals, respInterval)
	}

	resp := map[string]interface{}{
		"message":           "List of Reminder Intervals",
		"reminderIntervals": respIntervals,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) GetDocumentRemindersHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if doc.UserID.String() != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	reminders, err := h.repo.GetDocumentRemindersByDocumentID(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Failed to fetch document reminders", http.StatusInternalServerError)
		return
	}

	var rems []DocumentReminderIntervalResponse
	for _, reminder := range reminders {
		interval, err := h.repo.GetReminderIntervalByID(r.Context(), reminder.ReminderIntervalID)
		if err == nil {
			remInterval := DocumentReminderIntervalResponse{
				ID:      interval.IdLabel,
				Label:   interval.Label,
				Enabled: reminder.Enabled,
			}
			rems = append(rems, remInterval)
		}
	}

	docResp := map[string]interface{}{
		"id":          doc.ID.String(),
		"name":        doc.Name,
		"description": doc.Description,
		"reminders":   rems,
	}

	resp := map[string]interface{}{
		"message": "Document Reminders fetched successfully",
		"data":    docResp,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) ToggleDocumentReminderHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	if doc.UserID.String() != userID {
		http.Error(w, "Forbidden", http.StatusForbidden)
		return
	}

	var req ToggleDocumentReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	reminderIntervals, err := h.repo.GetReminderIntervalsFromIdLabels(r.Context(), []string{req.ReminderIntervalID})
	if err != nil || len(reminderIntervals) == 0 {
		http.Error(w, "Reminder interval not found", http.StatusNotFound)
		return
	}
	reminderInterval := reminderIntervals[0]
	err = h.repo.ToggleDocumentReminder(r.Context(), doc.ID.String(), reminderInterval.ID, req.Enabled)
	if err != nil {
		http.Error(w, "Failed to update document reminder", http.StatusInternalServerError)
		return
	}

	resp := map[string]interface{}{
		"message": "Document reminder updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
