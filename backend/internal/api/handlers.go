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
	worker "xpired/internal/worker"
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
		errResp := BadRequestError("Invalid request body")
		WriteErrorResponse(w, errResp)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		errResp := InternalServerError("Failed to hash password")
		WriteErrorResponse(w, errResp)
		return
	}

	if req.Email == "" || req.Password == "" || req.Name == "" {
		errResp := BadRequestError("Missing required fields")
		WriteErrorResponse(w, errResp)
		return
	}

	if err := h.repo.CheckUserExistsByEmail(r.Context(), req.Email); err == nil {
		errResp := ConflictError("User already exists")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to create user")
		WriteErrorResponse(w, errResp)
		return
	}

	token, err := auth.GenerateToken(newUser.ID)
	if err != nil {
		errResp := InternalServerError("Failed to generate token")
		WriteErrorResponse(w, errResp)
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
		"token":   token,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp := BadRequestError("Invalid request body")
		WriteErrorResponse(w, errResp)
		return
	}

	user, err := h.repo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		errResp := UnauthorizedError("Invalid email or password")
		WriteErrorResponse(w, errResp)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		errResp := UnauthorizedError("Invalid email or password")
		WriteErrorResponse(w, errResp)
		return
	}

	token, err := auth.GenerateToken(user.ID)
	if err != nil {
		errResp := InternalServerError("Failed to generate token")
		WriteErrorResponse(w, errResp)
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
		"token":   token,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}

	user, err := h.repo.GetUserByID(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}

func (h *Handler) ListDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
		return
	}

	documents, err := h.repo.ListDocumentsByUserID(r.Context(), userID)
	if err != nil {
		errResp := InternalServerError("Failed to fetch documents")
		WriteErrorResponse(w, errResp)
		return
	}

	resp := map[string]interface{}{
		"message":   "List of Documents",
		"documents": documents,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}

func (h *Handler) CreateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	var req DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp := BadRequestError("Invalid request body")
		WriteErrorResponse(w, errResp)
		return
	}

	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}
	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
		return
	}

	if req.Name == "" || req.ExpirationDate.IsZero() || req.Timezone == "" {
		errResp := BadRequestError("Missing required fields")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to create document")
		WriteErrorResponse(w, errResp)
		return
	}

	reminderIntervals, err := h.repo.GetReminderIntervalsFromIdLabels(r.Context(), req.Reminders)
	if err != nil {
		errResp := InternalServerError("Failed to fetch reminder intervals")
		WriteErrorResponse(w, errResp)
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
			errResp := InternalServerError("Failed to set document reminders")
			WriteErrorResponse(w, errResp)
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

	var reminderValues []db.ReminderInterval
	for _, interval := range reminderIntervals {
		reminderValues = append(reminderValues, *interval)
	}
	worker.ScheduleReminders(*newDoc, uuid.MustParse(userID), reminderValues)

	resp := map[string]interface{}{
		"message":  "Document created successfully",
		"document": doc,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}

func (h *Handler) GetDocumentHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" || documentId == "undefined" {
		errResp := BadRequestError("Document ID is required")
		WriteErrorResponse(w, errResp)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		errResp := NotFoundError("Document not found")
		WriteErrorResponse(w, errResp)
		return
	}

	if doc.UserID.String() != userID {
		errResp := ForbiddenError("Forbidden")
		WriteErrorResponse(w, errResp)
		return
	}

	reminders, err := h.repo.GetDocumentRemindersByDocumentID(r.Context(), documentId)
	if err != nil {
		errResp := InternalServerError("Failed to fetch document reminders")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}

func (h *Handler) UpdateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" || documentId == "undefined" {
		errResp := BadRequestError("Document ID is required")
		WriteErrorResponse(w, errResp)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		errResp := NotFoundError("Document not found")
		WriteErrorResponse(w, errResp)
		return
	}

	if doc.UserID.String() != userID {
		errResp := ForbiddenError("Forbidden")
		WriteErrorResponse(w, errResp)
		return
	}
	var req DocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp := BadRequestError("Invalid request body")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to update document")
		WriteErrorResponse(w, errResp)
		return
	}

	reminderIntervals, err := h.repo.GetReminderIntervalsFromIdLabels(r.Context(), req.Reminders)
	if err != nil {
		errResp := InternalServerError("Failed to fetch reminder intervals")
		WriteErrorResponse(w, errResp)
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
			errResp := InternalServerError("Failed to set document reminders")
			WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}

func (h *Handler) DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" || documentId == "undefined" {
		errResp := BadRequestError("Document ID is required")
		WriteErrorResponse(w, errResp)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		errResp := NotFoundError("Document not found")
		WriteErrorResponse(w, errResp)
		return
	}

	if doc.UserID.String() != userID {
		errResp := ForbiddenError("Forbidden")
		WriteErrorResponse(w, errResp)
		return
	}
	err = h.repo.DeleteDocument(r.Context(), documentId)
	if err != nil {
		errResp := InternalServerError("Failed to delete document")
		WriteErrorResponse(w, errResp)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) GetReminderIntervalsHandler(w http.ResponseWriter, r *http.Request) {
	intervals, err := h.repo.GetAllReminderIntervals(r.Context())
	if err != nil {
		errResp := InternalServerError("Failed to fetch reminder intervals")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}

func (h *Handler) GetDocumentRemindersHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" || documentId == "undefined" {
		errResp := BadRequestError("Document ID is required")
		WriteErrorResponse(w, errResp)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		errResp := NotFoundError("Document not found")
		WriteErrorResponse(w, errResp)
		return
	}

	if doc.UserID.String() != userID {
		errResp := ForbiddenError("Forbidden")
		WriteErrorResponse(w, errResp)
		return
	}

	reminders, err := h.repo.GetDocumentRemindersByDocumentID(r.Context(), documentId)
	if err != nil {
		errResp := InternalServerError("Failed to fetch document reminders")
		WriteErrorResponse(w, errResp)
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
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}

func (h *Handler) ToggleDocumentReminderHandler(w http.ResponseWriter, r *http.Request) {
	documentId := chi.URLParam(r, "id")
	if documentId == "" || documentId == "undefined" {
		errResp := BadRequestError("Document ID is required")
		WriteErrorResponse(w, errResp)
		return
	}
	userID, err := auth.GetUserIDFromContext(r)
	if err != nil {
		errResp := UnauthorizedError("Unauthorized")
		WriteErrorResponse(w, errResp)
		return
	}

	err = h.repo.CheckUserExistsById(r.Context(), userID)
	if err != nil {
		errResp := NotFoundError("User not found")
		WriteErrorResponse(w, errResp)
		return
	}

	doc, err := h.repo.GetDocumentByID(r.Context(), documentId)
	if err != nil {
		errResp := NotFoundError("Document not found")
		WriteErrorResponse(w, errResp)
		return
	}

	if doc.UserID.String() != userID {
		errResp := ForbiddenError("Forbidden")
		WriteErrorResponse(w, errResp)
		return
	}

	var req ToggleDocumentReminderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errResp := BadRequestError("Invalid request body")
		WriteErrorResponse(w, errResp)
		return
	}

	reminderIntervals, err := h.repo.GetReminderIntervalsFromIdLabels(r.Context(), []string{req.ReminderIntervalID})
	if err != nil || len(reminderIntervals) == 0 {
		errResp := NotFoundError("Reminder interval not found")
		WriteErrorResponse(w, errResp)
		return
	}
	reminderInterval := reminderIntervals[0]
	err = h.repo.ToggleDocumentReminder(r.Context(), doc.ID.String(), reminderInterval.ID, req.Enabled)
	if err != nil {
		errResp := InternalServerError("Failed to toggle document reminder")
		WriteErrorResponse(w, errResp)
		return
	}

	resp := map[string]interface{}{
		"message": "Document reminder updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		errResp := InternalServerError("Failed to encode response")
		WriteErrorResponse(w, errResp)
	}
}
