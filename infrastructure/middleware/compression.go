package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	return g.writer.Write(data)
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	return g.writer.Write([]byte(s))
}

func CompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// No comprimir rutas de ingestión (se hace hijack o el payload no lo amerita)
		if strings.HasPrefix(c.Request.URL.Path, "/api/iot/") ||
			(c.Request.Method == http.MethodPost && c.Request.URL.Path == "/api/leads") {
			c.Next()
			return
		}

		if !strings.Contains(c.GetHeader("Accept-Encoding"), "gzip") {
			c.Next()
			return
		}

		if c.Request.Method == "HEAD" {
			c.Next()
			return
		}

		contentType := c.GetHeader("Content-Type")
		if strings.Contains(contentType, "image/") ||
			strings.Contains(contentType, "video/") ||
			strings.Contains(contentType, "audio/") {
			c.Next()
			return
		}

		gz, err := gzip.NewWriterLevel(c.Writer, gzip.BestSpeed)
		if err != nil {
			c.Next()
			return
		}
		defer gz.Close()

		c.Header("Content-Encoding", "gzip")
		c.Header("Vary", "Accept-Encoding")

		c.Writer = &gzipWriter{
			ResponseWriter: c.Writer,
			writer:         gz,
		}

		c.Next()

		if c.Writer.Size() > 0 {
			gz.Flush()
		}
	}
}

func DecompressionMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Content-Encoding") == "gzip" {
			reader, err := gzip.NewReader(c.Request.Body)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
					"success": false,
					"message": "Error decomprimiendo request",
				})
				return
			}
			defer reader.Close()

			c.Request.Body = io.NopCloser(reader)
		}

		c.Next()
	}
}
