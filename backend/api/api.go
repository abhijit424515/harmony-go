package api

import (
	"fmt"
	"harmony/backend/handlers"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		token, err := c.Cookie("access_token")
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
			return
		}

		claims, err := handlers.VerifyAndDecodeToken(token)
		if err != nil {
			log.Println("]] HI 2" + err.Error())
			c.AbortWithStatusJSON(http.StatusUnauthorized, nil)
			return
		}

		uid, _ := claims["user_id"].(string)
		email, _ := claims["email"].(string)

		c.Set("user_id", uid)
		c.Set("email", email)
		c.Next()
	}
}

func Setup() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Harmony!")
	})

	r.GET("/user", func(c *gin.Context) {
		e := c.Query("email")
		uid, err := handlers.CreateOrGetUser(e)
		if err != nil {
			log.Println("[error]", err)
			c.String(http.StatusInternalServerError, "[error] creating user")
			return
		}

		payload := make(map[string]interface{})
		payload["email"] = e
		payload["user_id"] = uid

		token, err := handlers.GenerateAccessToken(payload, time.Hour*24)
		if err != nil {
			log.Println("[error]", err)
			c.String(http.StatusInternalServerError, "[error] generating token")
			return
		}

		c.SetCookie("access_token", token, 24*int(time.Hour.Seconds()), "/", "", false, true)
		c.String(http.StatusOK, "")
	})

	r.Use(AuthMiddleware()).GET("/test", func(c *gin.Context) {
		z, _ := c.Get("email")
		email := z.(string)
		z, _ = c.Get("user_id")
		uid := z.(string)

		msg := fmt.Sprintf("You are authorized! Email: %s, User ID: %s", email, uid)
		c.String(http.StatusOK, msg)
	})

	r.Use(AuthMiddleware()).POST("/clip/text", func(c *gin.Context) {
		if c.GetHeader("Content-Type") != "text/plain" {
			c.String(http.StatusBadRequest, "invalid content type")
			return
		}

		z, exists := c.Get("user_id")
		if !exists {
			c.String(http.StatusInternalServerError, "[error] getting user_id")
			return
		}
		user_id := z.(string)

		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, "[error] reading body")
			return
		}

		err = handlers.UpsertBuffer(user_id, data, handlers.TextType)
		if err != nil {
			log.Println("[error]", err)
		}

		c.String(http.StatusOK, "")
	})

	r.Use(AuthMiddleware()).POST("/clip/image", func(c *gin.Context) {
		if c.GetHeader("Content-Type") != "application/octet-stream" {
			c.String(http.StatusBadRequest, "invalid content type")
			return
		}

		z, exists := c.Get("user_id")
		if !exists {
			c.String(http.StatusInternalServerError, "[error] getting user_id")
			return
		}
		user_id := z.(string)

		buf, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, "[error] reading body")
			return
		}

		err = handlers.UpsertBuffer(user_id, buf, handlers.ImageType)
		if err != nil {
			log.Println("[error]", err)
		}

		c.String(http.StatusOK, "")
	})

	r.Run(":" + os.Getenv("PORT"))
}
