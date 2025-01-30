package api

import (
	"harmony/backend/handlers"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
)

func Setup() {
	r := gin.Default()

	r.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "Welcome to Harmony!")
	})

	r.POST("/clip/text", func(c *gin.Context) {
		if c.GetHeader("Content-Type") != "text/plain" {
			c.String(http.StatusBadRequest, "invalid content type")
			return
		}

		user_id := c.Query("user_id")

		data, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, "error reading body")
			return
		}

		err = handlers.UpsertBuffer(user_id, data, handlers.TextType)
		if err != nil {
			log.Println("[error]", err)
		}

		c.String(http.StatusOK, "")
	})

	r.POST("/clip/image", func(c *gin.Context) {
		if c.GetHeader("Content-Type") != "application/octet-stream" {
			c.String(http.StatusBadRequest, "invalid content type")
			return
		}

		user_id := c.Query("user_id")

		buf, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.String(http.StatusInternalServerError, "error reading body")
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
