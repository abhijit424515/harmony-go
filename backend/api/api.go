package api

import (
	"harmony/backend/cache"
	"harmony/backend/handlers"
	"harmony/backend/utils"
	"io"
	"net/http"
	"os"
	"strconv"
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

		claims, err := utils.VerifyAndDecodeToken(token)
		if err != nil {
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
			c.String(http.StatusInternalServerError, "[error] creating user")
			return
		}

		payload := make(map[string]any)
		payload["email"] = e
		payload["user_id"] = uid

		token, err := utils.GenerateAccessToken(payload, time.Hour*24)
		if err != nil {
			c.String(http.StatusInternalServerError, "[error] generating token")
			return
		}

		c.SetCookie("access_token", token, 24*int(time.Hour.Seconds()), "/", "", false, true)
		c.String(http.StatusOK, "")
	})

	r.Use(AuthMiddleware())

	r.GET("/user/check", func(c *gin.Context) {
		c.String(http.StatusOK, "")
	})

	r.GET("/buffer", func(c *gin.Context) {
		z, exists := c.Get("user_id")
		if !exists {
			c.String(http.StatusInternalServerError, "[error] getting user_id")
			return
		}
		user_id := z.(string)

		t := c.Query("ttl")
		if t != "" {
			ts, err := strconv.ParseInt(t, 10, 64)
			if err != nil {
				c.String(http.StatusBadRequest, "invalid timestamp")
				return
			}

			lts := cache.Get(user_id)
			if lts <= ts {
				c.String(http.StatusNotModified, "")
				return
			}
		}

		buf, bt, ttl, err := handlers.GetBuffer(user_id)
		if err != nil {
			c.String(http.StatusNoContent, "[error] buffer expired")
			return
		}

		ct := "text/plain"
		if bt == handlers.ImageType {
			ct = "application/octet-stream"
		}

		cache.Set(user_id, ttl)
		c.Header("X-Buffer-TTL", strconv.FormatInt(ttl, 10))
		c.Data(http.StatusOK, ct, buf)
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

		ttl, err := handlers.UpsertBuffer(user_id, data, handlers.TextType)
		if err != nil {
			c.String(http.StatusInternalServerError, "[error] upserting buffer")
			return
		}

		cache.Set(user_id, ttl)
		c.Data(http.StatusOK, "text/plain", []byte(strconv.FormatInt(ttl, 10)))
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

		ttl, err := handlers.UpsertBuffer(user_id, buf, handlers.ImageType)
		if err != nil {
			c.String(http.StatusInternalServerError, "[error] upserting buffer")
			return
		}

		cache.Set(user_id, ttl)
		c.Data(http.StatusOK, "text/plain", []byte(strconv.FormatInt(ttl, 10)))
	})

	r.Run(":" + os.Getenv("PORT"))
}
