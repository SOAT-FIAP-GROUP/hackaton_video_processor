package http

import (
	"frontend/internal/usecase"
	"log"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	c *usecase.CognitoAuth
}

func NewUserHandler(cognitoAuth *usecase.CognitoAuth) *UserHandler {
	return &UserHandler{
		c: cognitoAuth,
	}
}

func (h *UserHandler) Login(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")

	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	match, err := regexp.MatchString(emailRegex, email)
	if err != nil || !match {
		log.Println("Email format error: invalid email format")
		c.Redirect(http.StatusUnauthorized, "/")
		return
	}

	if len(password) < 6 {
		log.Println("Password length error: password must be at least 6 characters long")
		c.Redirect(http.StatusUnauthorized, "/")
		return
	}

	result, err := h.c.Login(email, password)
	if err != nil {
		log.Println("Login error:", err)
		c.Redirect(http.StatusUnauthorized, "/")
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    *result.AuthenticationResult.AccessToken,
		Path:     "/",
		MaxAge:   900,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})
	c.Redirect(http.StatusSeeOther, "/")
}

func (h *UserHandler) Signup(c *gin.Context) {
	email := c.PostForm("email")
	password := c.PostForm("password")
	name := c.PostForm("name")

	emailRegex := `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

	match, err := regexp.MatchString(emailRegex, email)
	if err != nil || !match {
		log.Println("Email format error: invalid email format")
		c.Redirect(http.StatusUnauthorized, "/")
		return
	}

	if len(password) < 6 {
		log.Println("Password length error: password must be at least 6 characters long")
		c.Redirect(http.StatusUnauthorized, "/")
		return
	}

	if len(name) < 2 || len(name) > 32 {
		log.Println("Name length error: name must be between 2 and 32 characters long")
		c.Redirect(http.StatusUnauthorized, "/")
		return
	}

	result, err := h.c.SignUp(email, password, name)
	if err != nil {
		log.Println("Signup error:", err)
		c.Redirect(http.StatusInternalServerError, "/")
		return
	}

	log.Printf("Signup successful: %v", result)

	c.Redirect(http.StatusSeeOther, "/login")
}

func (h *UserHandler) Logout(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "access_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
	})

	c.Redirect(http.StatusSeeOther, "/login")
}

func (h *UserHandler) HandleLoginPage(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, GetLoginPage())
}

func (h *UserHandler) HandleSignup(c *gin.Context) {
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(http.StatusOK, GetSignupPage())
}
