package middleware

import (
	"ZVideo/internal/domain"
	appLogger "ZVideo/internal/infrastructure/logger"
	"ZVideo/internal/testing/mocks"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestGetUserFromContext_Positive_StateTransition(t *testing.T) {
	// Arrange.
	want := &UserContext{UserID: 42, Role: domain.RoleModerator}
	ctx := context.WithValue(context.Background(), UserContextKey, want)

	// Act.
	got, err := GetUserFromContext(ctx)

	// Assert.
	require.NoError(t, err)
	require.Equal(t, want, got)
}

func TestGetUserFromContext_Negative_EquivalencePartitioning(t *testing.T) {
	// Arrange.
	ctx := context.Background()

	// Act.
	got, err := GetUserFromContext(ctx)

	// Assert.
	require.Error(t, err)
	require.Nil(t, got)
}

func TestAuth_Positive_StateTransition(t *testing.T) {
	// Arrange.
	authSvc := mocks.NewAuthService(t)
	authSvc.On("ValidateAccessToken", mock.MatchedBy(func(value context.Context) bool {
		return value != nil
	}), "access-token").Return(&domain.AccessTokenData{
		UserID:   42,
		UserName: "moderator",
		Role:     domain.RoleModerator,
	}, nil).Once()

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		user, err := GetUserFromContext(r.Context())
		require.NoError(t, err)
		require.Equal(t, &UserContext{UserID: 42, Role: domain.RoleModerator}, user)
		require.Equal(t, 42, r.Context().Value(domain.UserIDKey))
		w.WriteHeader(http.StatusNoContent)
	})
	handler := Auth(authSvc)(next)
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer access-token")
	recorder := httptest.NewRecorder()

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.True(t, nextCalled)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestAuth_Negative_EquivalencePartitioning(t *testing.T) {
	// Arrange.
	authSvc := mocks.NewAuthService(t)
	nextCalled := false
	next := http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		nextCalled = true
	})
	handler := Auth(authSvc)(next)
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Basic credentials")
	recorder := httptest.NewRecorder()

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.False(t, nextCalled)
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"UNAUTHORIZED"`)
}

func TestAuth_Negative_InvalidToken_Exception(t *testing.T) {
	// Arrange.
	authSvc := mocks.NewAuthService(t)
	authSvc.On("ValidateAccessToken", mock.MatchedBy(func(value context.Context) bool {
		return value != nil
	}), "expired-token").Return(nil, domain.ErrInvalidAccessToken).Once()
	handler := Auth(authSvc)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	}))
	req := httptest.NewRequest(http.MethodGet, "/private", nil)
	req.Header.Set("Authorization", "Bearer expired-token")
	recorder := httptest.NewRecorder()

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"INVALID_TOKEN"`)
}

func TestRequireRole_Positive_EquivalencePartitioning(t *testing.T) {
	// Arrange.
	ctx := context.WithValue(context.Background(), UserContextKey, &UserContext{UserID: 42, Role: domain.RoleAdmin})
	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	nextCalled := false
	handler := RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.True(t, nextCalled)
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestRequireRole_Negative_EquivalencePartitioning(t *testing.T) {
	// Arrange.
	req := httptest.NewRequest(http.MethodGet, "/admin", nil)
	handler := RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	}))
	recorder := httptest.NewRecorder()

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.Equal(t, http.StatusUnauthorized, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"UNAUTHORIZED"`)
}

func TestRequireRole_Negative_ForbiddenRole_EquivalencePartitioning(t *testing.T) {
	// Arrange.
	ctx := context.WithValue(context.Background(), UserContextKey, &UserContext{UserID: 42, Role: domain.RoleUser})
	req := httptest.NewRequest(http.MethodGet, "/admin", nil).WithContext(ctx)
	handler := RequireRole(domain.RoleAdmin)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("next handler must not be called")
	}))
	recorder := httptest.NewRecorder()

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.Equal(t, http.StatusForbidden, recorder.Code)
	require.Contains(t, recorder.Body.String(), `"code":"FORBIDDEN"`)
}

func TestRecovery_Positive_StateTransition(t *testing.T) {
	// Arrange.
	handler := Recovery(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.Equal(t, http.StatusNoContent, recorder.Code)
}

func TestRecovery_Negative_Panic_Exception(t *testing.T) {
	// Arrange.
	handler := Recovery(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("unexpected failure")
	}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	require.Contains(t, recorder.Body.String(), "Internal Server Error")
}

func TestLogging_Positive_StateTransition(t *testing.T) {
	// Arrange.
	var logOutput bytes.Buffer
	baseLogger := appLogger.NewSlogLogger(slog.LevelDebug, &logOutput, false)
	handler := Logging(baseLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, err := w.Write([]byte("ok"))
		require.NoError(t, err)
	}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.Equal(t, http.StatusOK, recorder.Code)
	require.Equal(t, "ok", recorder.Body.String())
	requireLogStatus(t, logOutput.Bytes(), http.StatusOK)
}

func TestLogging_Negative_ClientError_EquivalencePartitioning(t *testing.T) {
	// Arrange.
	var logOutput bytes.Buffer
	baseLogger := appLogger.NewSlogLogger(slog.LevelDebug, &logOutput, false)
	handler := Logging(baseLogger)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	// Act.
	handler.ServeHTTP(recorder, req)

	// Assert.
	require.Equal(t, http.StatusBadRequest, recorder.Code)
	requireLogStatus(t, logOutput.Bytes(), http.StatusBadRequest)
}

func requireLogStatus(t *testing.T, output []byte, wantStatus int) {
	t.Helper()
	decoder := json.NewDecoder(bytes.NewReader(output))
	found := false
	for {
		var entry map[string]any
		err := decoder.Decode(&entry)
		if errors.Is(err, io.EOF) {
			break
		}
		require.NoError(t, err)
		if status, ok := entry["status"].(float64); ok && int(status) == wantStatus {
			found = true
		}
	}
	require.True(t, found, "expected a completed request log with status %d, got %s", wantStatus, output)
}
