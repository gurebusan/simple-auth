package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"time"

	"github.com/go-chi/render"
	"github.com/gurebusan/simple-auth/internal/config"
	"github.com/gurebusan/simple-auth/internal/lib/logger/sl"
	resp "github.com/gurebusan/simple-auth/internal/lib/response"
	"github.com/gurebusan/simple-auth/internal/storage"
)

type issueRequest struct {
	GUID  string `json:"guid"`
	Email string `json:"email"`
}

type Auth interface {
	IssueTokens(GUID, email, ip string) (accessToken, refreshToken string, err error)
	RefreshTokens(GUID, ip, oldRefreshToken string) (newAccessToken, newRefreshToken string, err error)
}

type Handlers struct {
	log  *slog.Logger
	cfg  *config.Config
	auth Auth
}

func New(log *slog.Logger, cfg *config.Config, auth Auth) *Handlers {
	return &Handlers{
		log:  log,
		cfg:  cfg,
		auth: auth,
	}
}

func (h *Handlers) IssueTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Error("Method not allowed", slog.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req issueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.log.Error("Invalid request body", sl.Err(err))
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.GUID == "" {
		h.log.Error("Missing GUID")
		http.Error(w, "GUID is required field", http.StatusBadRequest)
		return
	}
	if !isValidEmail(req.Email) {
		h.log.Error("Missing or invalid email address")
		http.Error(w, "Missing or invalid email address", http.StatusBadRequest)
		return
	}
	ip := r.RemoteAddr
	accessToken, refreshToken, err := h.auth.IssueTokens(req.GUID, req.Email, ip)
	if err != nil {
		h.log.Error("Failed to issue tokens", sl.Err(err))
		http.Error(w, "Failed to issue tokens", http.StatusInternalServerError)
		return
	}
	h.log.Info("Tokens issued", slog.String("guid", req.GUID))

	setRefreshCookie(w, refreshToken, h.cfg.Token.RefreshTTL)
	setGUIDCookie(w, req.GUID, h.cfg.Token.RefreshTTL)
	resp := resp.New(req.GUID, accessToken)
	render.JSON(w, r, resp)
}

func (h *Handlers) RefreshTokens(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		h.log.Error("Method not allowed", slog.String("method", r.Method))
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cookie, err := r.Cookie("refresh_token")
	if err != nil {
		h.log.Error("Missing refresh token cookie")
		http.Error(w, "Missing refresh token", http.StatusUnauthorized)
		return
	}
	refreshToken := cookie.Value
	cookie, err = r.Cookie("guid")
	if err != nil {
		h.log.Error("Missing guid cookie")
		http.Error(w, "Missing guid", http.StatusUnauthorized)
		return
	}
	guid := cookie.Value
	ip := r.RemoteAddr

	newAccessToken, newRefreshToken, err := h.auth.RefreshTokens(guid, ip, refreshToken)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrTokenNotFound):
			h.log.Error("Refresh token not found or already used")
			http.Error(w, "Refresh token not found or already used", http.StatusUnauthorized)
			return
		case errors.Is(err, storage.ErrTokenExpired):
			h.log.Error("Refresh token expired")
			http.Error(w, "Refresh token expired", http.StatusUnauthorized)
			return
		default:
			h.log.Error("Failed to refresh tokens", sl.Err(err))
			http.Error(w, "Failed to refresh tokens", http.StatusInternalServerError)
			return
		}

	}
	h.log.Info("Tokens refreshed", slog.String("guid", guid))
	setRefreshCookie(w, newRefreshToken, h.cfg.Token.RefreshTTL)
	resp := resp.New(guid, newAccessToken)
	render.JSON(w, r, resp)
}

func setRefreshCookie(w http.ResponseWriter, refreshToken string, ttl time.Duration) {
	httpOnlyCookie := http.Cookie{
		Name:     "refresh_token",
		Value:    refreshToken,
		Expires:  time.Now().Add(ttl),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
	}
	http.SetCookie(w, &httpOnlyCookie)
}

func setGUIDCookie(w http.ResponseWriter, guid string, ttl time.Duration) {
	httpOnlyCookie := http.Cookie{
		Name:     "guid",
		Value:    guid,
		Expires:  time.Now().Add(ttl),
		Path:     "/",
		HttpOnly: true,
		Secure:   false, // Set to true in production
	}
	http.SetCookie(w, &httpOnlyCookie)
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}
