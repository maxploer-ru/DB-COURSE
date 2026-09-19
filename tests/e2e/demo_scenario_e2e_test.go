package e2e_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

type authResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
}

type userResponse struct {
	ID       int    `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

type channelResponse struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func TestDemoScenario_E2E_StateTransition_EquivalencePartitioning(t *testing.T) {
	baseURL := os.Getenv("E2E_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	client := &http.Client{Timeout: 10 * time.Second}
	waitForAPI(t, client, baseURL)

	suffix := strconv.FormatInt(time.Now().UnixNano(), 10)
	username := "e2e_user_" + suffix
	email := "e2e_" + suffix + "@example.com"
	password := "StrongPass123!"

	registerStatus, registerBody := requestJSON(t, client, http.MethodPost, baseURL+"/api/v1/auth/register", "", map[string]any{
		"username": username,
		"email":    email,
		"password": password,
	})
	require.Equal(t, http.StatusCreated, registerStatus, string(registerBody))

	loginStatus, loginBody := requestJSON(t, client, http.MethodPost, baseURL+"/api/v1/auth/login", "", map[string]any{
		"email":    email,
		"password": password,
	})
	require.Equal(t, http.StatusOK, loginStatus, string(loginBody))
	var auth authResponse
	require.NoError(t, json.Unmarshal(loginBody, &auth))
	require.NotEmpty(t, auth.AccessToken)
	require.NotEmpty(t, auth.RefreshToken)

	t.Cleanup(func() {
		_, _, _ = doJSON(context.Background(), client, http.MethodPost, baseURL+"/api/v1/auth/logout", auth.AccessToken, map[string]any{
			"refreshToken": auth.RefreshToken,
		})
		_, _, _ = doJSON(context.Background(), client, http.MethodDelete, baseURL+"/api/v1/users/me", auth.AccessToken, nil)
	})

	profileStatus, profileBody := requestJSON(t, client, http.MethodGet, baseURL+"/api/v1/users/me", auth.AccessToken, nil)
	require.Equal(t, http.StatusOK, profileStatus, string(profileBody))
	var profile userResponse
	require.NoError(t, json.Unmarshal(profileBody, &profile))
	require.Equal(t, username, profile.Username)
	require.Equal(t, email, profile.Email)

	createChannelStatus, createChannelBody := requestJSON(t, client, http.MethodPost, baseURL+"/api/v1/channels", auth.AccessToken, map[string]any{
		"name":        "e2e_channel_" + suffix,
		"description": "E2E demonstration channel",
	})
	require.Equal(t, http.StatusCreated, createChannelStatus, string(createChannelBody))
	var channel channelResponse
	require.NoError(t, json.Unmarshal(createChannelBody, &channel))
	require.NotZero(t, channel.ID)

	publicStatus, publicBody := requestJSON(t, client, http.MethodGet, baseURL+"/api/v1/channels/"+strconv.Itoa(channel.ID), "", nil)
	require.Equal(t, http.StatusOK, publicStatus, string(publicBody))
	var publicChannel channelResponse
	require.NoError(t, json.Unmarshal(publicBody, &publicChannel))
	require.Equal(t, channel.Name, publicChannel.Name)
}

func waitForAPI(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()
	deadline := time.Now().Add(30 * time.Second)
	for time.Now().Before(deadline) {
		response, err := client.Get(baseURL + "/api/v1/channels?limit=1")
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode == http.StatusOK {
				return
			}
		}
		time.Sleep(500 * time.Millisecond)
	}
	t.Fatalf("API did not become ready at %s", baseURL)
}

func requestJSON(t *testing.T, client *http.Client, method, url, accessToken string, payload any) (int, []byte) {
	t.Helper()
	status, responseBody, err := doJSON(t.Context(), client, method, url, accessToken, payload)
	require.NoError(t, err)
	return status, responseBody
}

func doJSON(ctx context.Context, client *http.Client, method, url, accessToken string, payload any) (int, []byte, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return 0, nil, err
		}
		body = bytes.NewReader(encoded)
	}
	request, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return 0, nil, err
	}
	request.Header.Set("Accept", "application/json")
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	response, err := client.Do(request)
	if err != nil {
		return 0, nil, err
	}
	defer response.Body.Close()
	responseBody, err := io.ReadAll(response.Body)
	if err != nil {
		return 0, nil, err
	}
	return response.StatusCode, responseBody, nil
}
