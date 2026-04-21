package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

var (
	apiBaseURL   = "http://localhost:8080"
	accessToken  string
	refreshToken string
	httpClient   = &http.Client{Timeout: 30 * time.Second}
)

func SetBaseURL(url string) { apiBaseURL = url }
func SetTokens(access, refresh string) {
	accessToken = access
	refreshToken = refresh
}
func ClearTokens() {
	accessToken = ""
	refreshToken = ""
}
func IsLoggedIn() bool { return accessToken != "" }

func doRequest(method, path string, body interface{}, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}
	req, err := http.NewRequest(method, apiBaseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if accessToken != "" {
		req.Header.Set("Authorization", "Bearer "+accessToken)
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errResp struct {
			Error struct {
				Code    string `json:"code"`
				Message string `json:"message"`
			} `json:"error"`
		}
		json.NewDecoder(resp.Body).Decode(&errResp)
		return fmt.Errorf("ошибка API (%d): %s - %s", resp.StatusCode, errResp.Error.Code, errResp.Error.Message)
	}
	if result != nil {
		return json.NewDecoder(resp.Body).Decode(result)
	}
	return nil
}

type AuthResponse struct {
	User         UserResponse `json:"user"`
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
}
type UserResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}
type ChannelResponse struct {
	ID               int    `json:"id"`
	UserID           int    `json:"user_id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	SubscribersCount int    `json:"subscribers_count"`
}
type VideoResponse struct {
	ID          int    `json:"id"`
	ChannelID   int    `json:"channel_id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Views       int    `json:"views"`
	Likes       int    `json:"likes"`
	Dislikes    int    `json:"dislikes"`
	Comments    int    `json:"comments"`
	CreatedAt   string `json:"created_at"`
}
type UploadURLResponse struct {
	URL     string `json:"url"`
	FileKey string `json:"file_key"`
}
type StreamURLResponse struct {
	URL string `json:"url"`
}

type CommentResponse struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	VideoID   int       `json:"video_id"`
	Content   string    `json:"content"`
	Likes     int       `json:"likes"`
	Dislikes  int       `json:"dislikes"`
	CreatedAt time.Time `json:"created_at"`
}

type CommentListResponse struct {
	Comments []CommentResponse `json:"comments"`
	Total    int64             `json:"total,omitempty"`
}

type SubscriptionChannelResponse struct {
	ID               int    `json:"id"`
	UserID           int    `json:"user_id"`
	Name             string `json:"name"`
	Description      string `json:"description"`
	SubscribersCount int    `json:"subscribers_count"`
	NewVideosCount   int    `json:"new_videos_count"`
	SubscribedAt     string `json:"subscribed_at"`
}

func banUser(userID int) error {
	return doRequest("POST", fmt.Sprintf("/admin/users/%d/ban", userID), nil, nil)
}

func unbanUser(userID int) error {
	return doRequest("POST", fmt.Sprintf("/admin/users/%d/unban", userID), nil, nil)
}
