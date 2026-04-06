package handlers

import (
	"ZVideo/internal/domain/video/usecase"
	"encoding/json"
	"net/http"
)

type VideoHandler struct {
	generateURLUseCase *usecase.GenerateUploadURLUseCase
	createVideoUseCase *usecase.CreateVideoUseCase
}

func NewVideoHandler(generateURLUseCase *usecase.GenerateUploadURLUseCase, createVideoUseCase *usecase.CreateVideoUseCase) *VideoHandler {
	return &VideoHandler{
		generateURLUseCase: generateURLUseCase,
		createVideoUseCase: createVideoUseCase,
	}
}

func (h *VideoHandler) GetUploadURL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Filename string `json:"filename"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input JSON", http.StatusBadRequest)
		return
	}

	uploadURL, filePath, err := h.generateURLUseCase.Execute(r.Context(), req.Filename)
	if err != nil {
		http.Error(w, "Error generating URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"upload_url": uploadURL,
		"filepath":   filePath,
	})
}

func (h *VideoHandler) CreateVideo(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ChannelID   int    `json:"channel_id"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Filepath    string `json:"filepath"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input JSON", http.StatusBadRequest)
		return
	}

	userID := 1

	ctx := r.Context()
	video, err := h.createVideoUseCase.Execute(
		ctx,
		req.ChannelID,
		userID,
		req.Title,
		req.Description,
		req.Filepath,
	)
	if err != nil {
		http.Error(w, "Error creating video: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(video)
}
