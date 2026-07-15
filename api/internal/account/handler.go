package account

import (
	"errors"
	"net/http"
	"strings"
	"time"

	appweb "github.com/ZevixQuatol/whole-zone/api/internal/web"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const principalKey = "account.principal"

type Handler struct {
	service    *Service
	appURL     string
	cookieName string
	secure     bool
}

func NewHandler(service *Service, appURL, cookieName string, secure bool) *Handler {
	return &Handler{service: service, appURL: strings.TrimRight(appURL, "/"), cookieName: cookieName, secure: secure}
}

func (h *Handler) RegisterRoutes(api *gin.RouterGroup) {
	account := api.Group("/account")
	account.POST("/register", h.register)
	account.POST("/login", h.login)
	account.POST("/password/forgot", h.forgotPassword)
	account.POST("/password/reset", h.resetPassword)
	api.GET("/users/:handle", h.publicProfile)

	auth := account.Group("")
	auth.Use(h.authenticate, h.csrf)
	auth.GET("/me", h.me)
	auth.PATCH("/profile", h.updateProfile)
	auth.POST("/password/change", h.changePassword)
	auth.POST("/logout", h.logout)
	auth.POST("/logout-all", h.logoutAll)
	auth.GET("/notifications", h.notifications)
	auth.PATCH("/notifications/:id/read", h.readNotification)

	admin := api.Group("/admin")
	admin.Use(h.authenticate, h.csrf, h.admin)
	admin.GET("/invites", h.invites)
	admin.POST("/invites", h.createInvite)
	admin.POST("/invites/:id/disable", h.disableInvite)
}

func (h *Handler) register(c *gin.Context) {
	var input struct {
		InviteCode  string `json:"inviteCode"`
		Email       string `json:"email"`
		Handle      string `json:"handle"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	if !bind(c, &input) {
		return
	}
	result, err := h.service.Register(c.Request.Context(), RegisterInput{
		InviteCode: input.InviteCode, Email: input.Email, Handle: input.Handle,
		Password: input.Password, DisplayName: input.DisplayName,
	}, c.Request.UserAgent())
	if err != nil {
		h.fail(c, err)
		return
	}
	h.setAuth(c, result)
	c.JSON(http.StatusCreated, gin.H{"user": result.User.Public()})
}

func (h *Handler) login(c *gin.Context) {
	var input LoginInput
	if !bind(c, &input) {
		return
	}
	result, err := h.service.Login(c.Request.Context(), input, c.Request.UserAgent())
	if err != nil {
		h.fail(c, err)
		return
	}
	h.setAuth(c, result)
	c.JSON(http.StatusOK, gin.H{"user": result.User.Public()})
}

func (h *Handler) me(c *gin.Context) {
	principal := mustPrincipal(c)
	c.JSON(http.StatusOK, gin.H{"user": principal.User.Public(), "role": principal.User.Role})
}

func (h *Handler) updateProfile(c *gin.Context) {
	var input ProfileInput
	if !bind(c, &input) {
		return
	}
	user, err := h.service.UpdateProfile(c.Request.Context(), mustPrincipal(c).User.ID, input)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user.Public()})
}

func (h *Handler) publicProfile(c *gin.Context) {
	profile, err := h.service.PublicProfile(c.Request.Context(), c.Param("handle"))
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": profile})
}

func (h *Handler) logout(c *gin.Context) {
	if err := h.service.Logout(c.Request.Context(), mustPrincipal(c).Session.ID); err != nil {
		h.fail(c, err)
		return
	}
	h.clearAuth(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) logoutAll(c *gin.Context) {
	if err := h.service.LogoutAll(c.Request.Context(), mustPrincipal(c).User.ID); err != nil {
		h.fail(c, err)
		return
	}
	h.clearAuth(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) changePassword(c *gin.Context) {
	var input struct {
		Current string `json:"currentPassword"`
		Next    string `json:"newPassword"`
	}
	if !bind(c, &input) {
		return
	}
	if err := h.service.ChangePassword(c.Request.Context(), mustPrincipal(c).User, input.Current, input.Next); err != nil {
		h.fail(c, err)
		return
	}
	h.clearAuth(c)
	c.Status(http.StatusNoContent)
}

func (h *Handler) forgotPassword(c *gin.Context) {
	var input struct {
		Email string `json:"email"`
	}
	if !bind(c, &input) {
		return
	}
	if err := h.service.ForgotPassword(c.Request.Context(), input.Email); err != nil {
		// 找回密码始终返回相同结果，避免通过邮件发送或数据库状态枚举账号。
		_ = c.Error(err)
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) resetPassword(c *gin.Context) {
	var input struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if !bind(c, &input) {
		return
	}
	if err := h.service.ResetPassword(c.Request.Context(), input.Token, input.Password); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) notifications(c *gin.Context) {
	items, err := h.service.Notifications(c.Request.Context(), mustPrincipal(c).User.ID, 50)
	if err != nil {
		h.fail(c, err)
		return
	}
	responses := make([]notificationResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, newNotificationResponse(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": responses})
}

func (h *Handler) readNotification(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		appweb.Abort(c, http.StatusUnprocessableEntity, "INVALID_ID", "通知 ID 无效")
		return
	}
	if err := h.service.ReadNotification(c.Request.Context(), mustPrincipal(c).User.ID, id); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) invites(c *gin.Context) {
	items, err := h.service.Invites(c.Request.Context(), mustPrincipal(c).User, 50)
	if err != nil {
		h.fail(c, err)
		return
	}
	responses := make([]inviteResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, newInviteResponse(item))
	}
	c.JSON(http.StatusOK, gin.H{"items": responses})
}

func (h *Handler) createInvite(c *gin.Context) {
	var input struct {
		MaxUses int `json:"maxUses"`
		TTLDays int `json:"ttlDays"`
	}
	if !bind(c, &input) {
		return
	}
	result, err := h.service.CreateInvite(c.Request.Context(), mustPrincipal(c).User, input.MaxUses, time.Duration(input.TTLDays)*24*time.Hour)
	if err != nil {
		h.fail(c, err)
		return
	}
	c.JSON(http.StatusCreated, gin.H{"invite": newInviteResponse(result.Invite), "code": result.Code})
}

func (h *Handler) disableInvite(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		appweb.Abort(c, http.StatusUnprocessableEntity, "INVALID_ID", "邀请码 ID 无效")
		return
	}
	if err := h.service.DisableInvite(c.Request.Context(), mustPrincipal(c).User, id); err != nil {
		h.fail(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *Handler) authenticate(c *gin.Context) {
	raw, err := c.Cookie(h.cookieName)
	if err != nil {
		appweb.Abort(c, http.StatusUnauthorized, "UNAUTHENTICATED", "请先登录")
		return
	}
	principal, err := h.service.Authenticate(c.Request.Context(), raw)
	if err != nil {
		h.clearAuth(c)
		appweb.Abort(c, http.StatusUnauthorized, "UNAUTHENTICATED", "登录已失效")
		return
	}
	c.Set(principalKey, principal)
}

func (h *Handler) csrf(c *gin.Context) {
	if c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions {
		return
	}
	if origin := c.GetHeader("Origin"); origin != "" && origin != h.appURL {
		appweb.Abort(c, http.StatusForbidden, "ORIGIN_INVALID", "请求来源无效")
		return
	}
	if err := h.service.CheckCSRF(mustPrincipal(c).Session, c.GetHeader("X-CSRF-Token")); err != nil {
		appweb.Abort(c, http.StatusForbidden, "CSRF_INVALID", "请求校验失败")
	}
}

func (h *Handler) admin(c *gin.Context) {
	if mustPrincipal(c).User.Role != RoleAdmin {
		appweb.Abort(c, http.StatusForbidden, "FORBIDDEN", "无权执行此操作")
	}
}

func (h *Handler) setAuth(c *gin.Context, result AuthResult) {
	maxAge := int(time.Until(result.ExpiresAt).Seconds())
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, result.SessionToken, maxAge, "/", "", h.secure, true)
	c.SetCookie("hy_csrf", result.CSRFToken, maxAge, "/", "", h.secure, false)
}

func (h *Handler) clearAuth(c *gin.Context) {
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie(h.cookieName, "", -1, "/", "", h.secure, true)
	c.SetCookie("hy_csrf", "", -1, "/", "", h.secure, false)
}

func (h *Handler) fail(c *gin.Context, err error) {
	var validation ValidationError
	switch {
	case errors.Is(err, ErrAuth):
		appweb.Abort(c, http.StatusUnauthorized, "INVALID_CREDENTIALS", "账号或密码错误")
	case errors.Is(err, ErrDisabled):
		appweb.Abort(c, http.StatusForbidden, "ACCOUNT_DISABLED", "账号已停用")
	case errors.Is(err, ErrForbidden):
		appweb.Abort(c, http.StatusForbidden, "FORBIDDEN", "无权执行此操作")
	case errors.Is(err, ErrInviteInvalid):
		appweb.Abort(c, http.StatusUnprocessableEntity, "INVITE_INVALID", "邀请码无效或已过期")
	case errors.Is(err, ErrConflict):
		appweb.Abort(c, http.StatusConflict, "ACCOUNT_CONFLICT", "邮箱或用户名已被使用")
	case errors.Is(err, ErrNotFound):
		appweb.Abort(c, http.StatusNotFound, "NOT_FOUND", "资源不存在")
	case errors.Is(err, ErrTokenInvalid):
		appweb.Abort(c, http.StatusUnprocessableEntity, "TOKEN_INVALID", "重置链接无效或已过期")
	case errors.As(err, &validation):
		appweb.Abort(c, http.StatusUnprocessableEntity, "VALIDATION_FAILED", validation.Message)
	default:
		appweb.Abort(c, http.StatusInternalServerError, "INTERNAL_ERROR", "系统暂时无法处理请求")
	}
}

func bind(c *gin.Context, target any) bool {
	if err := c.ShouldBindJSON(target); err != nil {
		appweb.Abort(c, http.StatusBadRequest, "INVALID_JSON", "请求格式错误")
		return false
	}
	return true
}

func mustPrincipal(c *gin.Context) Principal {
	value, _ := c.Get(principalKey)
	return value.(Principal)
}

type inviteResponse struct {
	ID         uuid.UUID  `json:"id"`
	MaxUses    int        `json:"maxUses"`
	Uses       int        `json:"uses"`
	ExpiresAt  time.Time  `json:"expiresAt"`
	DisabledAt *time.Time `json:"disabledAt,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func newInviteResponse(invite Invite) inviteResponse {
	return inviteResponse{
		ID: invite.ID, MaxUses: invite.MaxUses, Uses: invite.Uses,
		ExpiresAt: invite.ExpiresAt, DisabledAt: invite.DisabledAt, CreatedAt: invite.CreatedAt,
	}
}

type notificationResponse struct {
	ID        uuid.UUID      `json:"id"`
	Kind      string         `json:"kind"`
	Title     string         `json:"title"`
	Body      string         `json:"body"`
	Data      map[string]any `json:"data"`
	ReadAt    *time.Time     `json:"readAt,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
}

func newNotificationResponse(notification Notification) notificationResponse {
	return notificationResponse{
		ID: notification.ID, Kind: notification.Kind, Title: notification.Title,
		Body: notification.Body, Data: notification.Data, ReadAt: notification.ReadAt, CreatedAt: notification.CreatedAt,
	}
}
