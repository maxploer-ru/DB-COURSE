package handlers

import (
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"net/http"
)

func (h *Handler) ListVideoComments(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId, params openapi.ListVideoCommentsParams) {
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.comment.ListByVideo(r.Context(), int(videoID), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	items := make([]openapi.Comment, 0, len(page.Items))
	for _, comment := range page.Items {
		if comment != nil {
			items = append(items, toComment(comment))
		}
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.CommentPage{
		Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount,
	})
}

func (h *Handler) CreateVideoComment(w http.ResponseWriter, r *http.Request, videoID openapi.VideoId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(videoID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.CreateVideoCommentJSONRequestBody](w, r)
	if !ok {
		return
	}
	comment, err := h.comment.Create(r.Context(), userID, int(videoID), body.Content)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, toComment(comment))
}

func (h *Handler) GetComment(w http.ResponseWriter, r *http.Request, commentID openapi.CommentId) {
	if !validID(int(commentID)) {
		badID(w)
		return
	}
	comment, err := h.comment.GetByID(r.Context(), int(commentID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toComment(comment))
}

func (h *Handler) UpdateComment(w http.ResponseWriter, r *http.Request, commentID openapi.CommentId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(commentID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdateCommentJSONRequestBody](w, r)
	if !ok {
		return
	}
	comment, err := h.comment.Update(r.Context(), int(commentID), userID, body.Content)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toComment(comment))
}

func (h *Handler) DeleteComment(w http.ResponseWriter, r *http.Request, commentID openapi.CommentId) {
	userID, role, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(commentID)) {
		badID(w)
		return
	}
	if err := h.comment.Delete(r.Context(), int(commentID), userID, role); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) RateComment(w http.ResponseWriter, r *http.Request, commentID openapi.CommentId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(commentID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.RateCommentJSONRequestBody](w, r)
	if !ok {
		return
	}
	if err := h.commentInteraction.Rate(r.Context(), userID, int(commentID), domain.RatingAction(body.Action)); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) GetCommentStats(w http.ResponseWriter, r *http.Request, commentID openapi.CommentId) {
	if !validID(int(commentID)) {
		badID(w)
		return
	}
	if _, err := h.comment.GetByID(r.Context(), int(commentID)); err != nil {
		handleError(w, err)
		return
	}
	likes, dislikes, err := h.commentInteraction.GetStats(r.Context(), int(commentID))
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.RatingStats{Likes: int(likes), Dislikes: int(dislikes)})
}
