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

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
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

	w.WriteHeader(http.StatusOK)
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

	w.WriteHeader(http.StatusOK)
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
	w.WriteHeader(http.StatusOK)
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
		Reminders:      []db.DocumentReminder{},
		CreatedAt:      newDoc.CreatedAt,
		UpdatedAt:      newDoc.UpdatedAt,
	}

	resp := map[string]interface{}{
		"message":  "Document created successfully",
		"document": doc,
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "application/json")
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

	docResp := &DocumentResponse{
		ID:             doc.ID.String(),
		UserID:         doc.UserID.String(),
		Name:           doc.Name,
		Description:    doc.Description,
		Identifier:     doc.Identifier,
		ExpirationDate: doc.ExpirationDate.Format("Mon, 2 Jan, 2006"),
		Timezone:       doc.Timezone,
		AttachmentURL:  doc.AttachmentURL,
		Reminders:      []db.DocumentReminder{},
		CreatedAt:      doc.CreatedAt,
		UpdatedAt:      doc.UpdatedAt,
	}

	resp := map[string]interface{}{
		"message":  "Document fetched successfully",
		"document": docResp,
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) UpdateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Update a specific document by ID
	w.Write([]byte("Document Updated"))
}

func (h *Handler) DeleteDocumentHandler(w http.ResponseWriter, r *http.Request) {
	// Delete a specific document by ID
	w.Write([]byte("Document Deleted"))
}
