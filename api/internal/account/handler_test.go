package account

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func TestHandlerRegisterSetsSecureAuthCookies(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := NewService(&fakeRepo{}, time.Hour, time.Now)
	handler := NewHandler(service, "https://huanyu.example", "hy_session", true)
	router := gin.New()
	handler.RegisterRoutes(router.Group("/api/v1"))
	invite, _ := NewToken()
	body, _ := json.Marshal(map[string]string{
		"inviteCode":  invite.Raw,
		"email":       "dev@example.com",
		"handle":      "dev_user",
		"password":    "long-enough-password",
		"displayName": "开发者",
	})

	request := httptest.NewRequest(http.MethodPost, "/api/v1/account/register", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	cookies := response.Result().Cookies()
	if len(cookies) != 2 {
		t.Fatalf("cookies = %d, want 2", len(cookies))
	}
	if cookies[0].Name != "hy_session" || !cookies[0].HttpOnly || !cookies[0].Secure {
		t.Fatalf("unexpected session cookie: %+v", cookies[0])
	}
	if cookies[1].Name != "hy_csrf" || cookies[1].HttpOnly || !cookies[1].Secure {
		t.Fatalf("unexpected csrf cookie: %+v", cookies[1])
	}
	if bytes.Contains(response.Body.Bytes(), []byte("password")) {
		t.Fatal("response leaked a password field")
	}
}

func TestHandlerMeRequiresSession(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(NewService(&fakeRepo{}, time.Hour, time.Now), "http://localhost:3000", "hy_session", false)
	router := gin.New()
	handler.RegisterRoutes(router.Group("/api/v1"))

	request := httptest.NewRequest(http.MethodGet, "/api/v1/account/me", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", response.Code)
	}
}

func TestHandlerProtectedWriteRequiresCSRF(t *testing.T) {
	router, repo, sessionToken, _ := authenticatedRouter(t, RoleMember)
	body := strings.NewReader(`{"displayName":"林屿","bio":"Go 开发者","location":"上海"}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/account/profile", body)
	request.Header.Set("Content-Type", "application/json")
	request.AddCookie(&http.Cookie{Name: "hy_session", Value: sessionToken})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if repo.profile != nil {
		t.Fatal("profile was updated without csrf validation")
	}
}

func TestHandlerProtectedWriteRejectsForeignOrigin(t *testing.T) {
	router, _, sessionToken, csrfToken := authenticatedRouter(t, RoleMember)
	body := strings.NewReader(`{"displayName":"林屿","bio":"Go 开发者","location":"上海"}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/account/profile", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "https://other.example")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(&http.Cookie{Name: "hy_session", Value: sessionToken})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func TestHandlerProtectedWriteAcceptsSameOriginAndCSRF(t *testing.T) {
	router, repo, sessionToken, csrfToken := authenticatedRouter(t, RoleMember)
	body := strings.NewReader(`{"displayName":"林屿","bio":"Go 开发者","location":"上海"}`)
	request := httptest.NewRequest(http.MethodPatch, "/api/v1/account/profile", body)
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(&http.Cookie{Name: "hy_session", Value: sessionToken})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if repo.profile == nil || repo.profile.Bio != "Go 开发者" {
		t.Fatalf("profile was not persisted: %+v", repo.profile)
	}
}

func TestHandlerAdminRoutesRejectMembers(t *testing.T) {
	router, repo, sessionToken, csrfToken := authenticatedRouter(t, RoleMember)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invites", strings.NewReader(`{"maxUses":1,"ttlDays":7}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(&http.Cookie{Name: "hy_session", Value: sessionToken})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if repo.invite != nil {
		t.Fatal("member created an invitation")
	}
}

func TestHandlerAdminCanCreateInviteWithoutLeakingHash(t *testing.T) {
	router, repo, sessionToken, csrfToken := authenticatedRouter(t, RoleAdmin)
	request := httptest.NewRequest(http.MethodPost, "/api/v1/admin/invites", strings.NewReader(`{"maxUses":2,"ttlDays":7}`))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Origin", "http://localhost:3000")
	request.Header.Set("X-CSRF-Token", csrfToken)
	request.AddCookie(&http.Cookie{Name: "hy_session", Value: sessionToken})
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
	if repo.invite == nil || repo.invite.MaxUses != 2 {
		t.Fatalf("invite was not persisted: %+v", repo.invite)
	}
	if !strings.Contains(response.Body.String(), `"code"`) || strings.Contains(response.Body.String(), "codeHash") {
		t.Fatalf("unexpected invite response: %s", response.Body.String())
	}
}

func TestHandlerForgotPasswordHidesDeliveryFailure(t *testing.T) {
	gin.SetMode(gin.TestMode)
	hash, err := HashPassword("long-enough-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	user := User{
		ID: uuid.New(), Email: "dev@example.com", Handle: "dev_user",
		PasswordHash: hash, Role: RoleMember, Status: StatusActive,
	}
	service := NewService(&fakeRepo{users: map[string]User{"dev@example.com": user}}, time.Hour, time.Now)
	service.SetMailer(&fakeMailer{err: errors.New("smtp unavailable")}, 30*time.Minute)
	handler := NewHandler(service, "http://localhost:3000", "hy_session", false)
	router := gin.New()
	handler.RegisterRoutes(router.Group("/api/v1"))
	request := httptest.NewRequest(http.MethodPost, "/api/v1/account/password/forgot", strings.NewReader(`{"email":"dev@example.com"}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status = %d, body = %s", response.Code, response.Body.String())
	}
}

func authenticatedRouter(t *testing.T, role Role) (*gin.Engine, *fakeRepo, string, string) {
	t.Helper()
	sessionToken, err := NewToken()
	if err != nil {
		t.Fatalf("session token: %v", err)
	}
	csrfToken, err := NewToken()
	if err != nil {
		t.Fatalf("csrf token: %v", err)
	}
	user := User{
		ID: uuid.New(), Handle: "dev_user", DisplayName: "开发者",
		Role: role, Status: StatusActive, CreatedAt: time.Now(),
	}
	session := Session{ID: uuid.New(), UserID: user.ID, CSRFHash: csrfToken.Hash}
	repo := &fakeRepo{authSession: &session, authUser: &user}
	handler := NewHandler(NewService(repo, time.Hour, time.Now), "http://localhost:3000", "hy_session", false)
	router := gin.New()
	handler.RegisterRoutes(router.Group("/api/v1"))
	return router, repo, sessionToken.Raw, csrfToken.Raw
}
