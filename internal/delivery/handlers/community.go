package handlers

import (
	"ZVideo/internal/delivery/openapi"
	"ZVideo/internal/delivery/response"
	"ZVideo/internal/domain"
	"net/http"
)

func (h *Handler) GetChannelCommunity(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId, params openapi.GetChannelCommunityParams) {
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	community, err := h.community.GetChannelCommunity(r.Context(), int(channelID), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toCommunity(community))
}

func (h *Handler) GetMyCommunity(w http.ResponseWriter, r *http.Request, params openapi.GetMyCommunityParams) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	community, err := h.community.GetMyCommunity(r.Context(), userID, limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toCommunity(community))
}

func (h *Handler) CreateCommunityPost(w http.ResponseWriter, r *http.Request, channelID openapi.ChannelId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(channelID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.CreateCommunityPostJSONRequestBody](w, r)
	if !ok {
		return
	}
	post, err := h.community.CreatePost(r.Context(), int(channelID), userID, body.Content)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, toCommunityPost(post))
}

func (h *Handler) UpdateCommunityPost(w http.ResponseWriter, r *http.Request, postID openapi.PostId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(postID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdateCommunityPostJSONRequestBody](w, r)
	if !ok {
		return
	}
	post, err := h.community.UpdatePost(r.Context(), int(postID), userID, body.Content)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toCommunityPost(post))
}

func (h *Handler) DeleteCommunityPost(w http.ResponseWriter, r *http.Request, postID openapi.PostId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(postID)) {
		badID(w)
		return
	}
	if err := h.community.DeletePost(r.Context(), int(postID), userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func (h *Handler) ListCommunityPostComments(w http.ResponseWriter, r *http.Request, postID openapi.PostId, params openapi.ListCommunityPostCommentsParams) {
	if !validID(int(postID)) {
		badID(w)
		return
	}
	limit, offset, err := pageParams(params.Limit, params.Offset)
	if err != nil {
		badRequest(w, "INVALID_PAGINATION", err.Error())
		return
	}
	page, err := h.community.GetPostComments(r.Context(), int(postID), limit, offset)
	if err != nil {
		handleError(w, err)
		return
	}
	items := make([]openapi.CommunityComment, 0, len(page.Items))
	for _, comment := range page.Items {
		if comment != nil {
			items = append(items, toCommunityComment(comment))
		}
	}
	response.RespondWithJSON(w, http.StatusOK, openapi.CommunityCommentPage{
		Items: items, Limit: page.Limit, Offset: page.Offset, TotalCount: page.TotalCount,
	})
}

func (h *Handler) CreateCommunityComment(w http.ResponseWriter, r *http.Request, postID openapi.PostId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(postID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.CreateCommunityCommentJSONRequestBody](w, r)
	if !ok {
		return
	}
	comment, err := h.community.CreateComment(r.Context(), int(postID), userID, body.Content)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusCreated, toCommunityComment(comment))
}

func (h *Handler) UpdateCommunityComment(w http.ResponseWriter, r *http.Request, commentID openapi.CommunityCommentId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(commentID)) {
		badID(w)
		return
	}
	body, ok := decodeJSON[openapi.UpdateCommunityCommentJSONRequestBody](w, r)
	if !ok {
		return
	}
	comment, err := h.community.UpdateComment(r.Context(), int(commentID), userID, body.Content)
	if err != nil {
		handleError(w, err)
		return
	}
	response.RespondWithJSON(w, http.StatusOK, toCommunityComment(comment))
}

func (h *Handler) DeleteCommunityComment(w http.ResponseWriter, r *http.Request, commentID openapi.CommunityCommentId) {
	userID, _, ok := requireCurrentUser(w, r)
	if !ok {
		return
	}
	if !validID(int(commentID)) {
		badID(w)
		return
	}
	if err := h.community.DeleteComment(r.Context(), int(commentID), userID); err != nil {
		handleError(w, err)
		return
	}
	writeNoContent(w)
}

func toCommunity(community *domain.Community) openapi.Community {
	posts := make([]openapi.CommunityPost, 0, len(community.Posts))
	for _, post := range community.Posts {
		if post != nil {
			posts = append(posts, toCommunityPost(post))
		}
	}
	return openapi.Community{
		Channel:    toChannel(community.Channel),
		Posts:      posts,
		TotalCount: community.TotalCount,
		Limit:      community.Limit,
		Offset:     community.Offset,
	}
}
