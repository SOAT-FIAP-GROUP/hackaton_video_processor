package http

import (
	"frontend/internal/usecase"

	"github.com/gin-gonic/gin"
)

type Router struct {
	Engine *gin.Engine
}

func SetupRouter(vh *VideoHandler, uh *UserHandler, cognito *usecase.CognitoAuth) *Router {
	r := gin.Default()

	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	r.Static("/uploads", "./uploads")
	r.Static("/outputs", "./outputs")

	r.GET("/", AuthMiddleware(cognito), vh.HandleHome)
	r.POST("/upload", AuthMiddleware(cognito), vh.HandleUpload)
	r.GET("/download/:filename", AuthMiddleware(cognito), vh.HandleDownload)
	r.GET("/api/status", AuthMiddleware(cognito), vh.HandleStatus)

	r.GET("/login", uh.HandleLoginPage)
	r.GET("/signup", uh.HandleSignup)
	r.POST("/api/signup", uh.Signup)
	r.POST("/api/login", uh.Login)
	r.GET("/api/logout", AuthMiddleware(cognito), uh.Logout)

	return &Router{Engine: r}
}

func AuthMiddleware(auth *usecase.CognitoAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString, err := c.Cookie("access_token")
		if err != nil {
			c.Redirect(303, "/login")
			c.Abort()
			return
		}

		claims, err := auth.ParseToken(tokenString)
		if err != nil {
			c.Redirect(303, "/login")
			return
		}

		c.Set("userID", claims.UserID)
		c.Set("email", claims.Email)

		c.Next()
	}
}
