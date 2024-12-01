package handlers

import (
	"net/http"
	"strings"

	"static-admin/embedded_box/frontend"

	"github.com/gin-gonic/gin"
)

func Index(ctx *gin.Context) {
	file := strings.TrimPrefix(ctx.Request.URL.Path, "/")

	// standardize on swagger/ as path
	if file == "swagger" || file == "swagger/index.html" {
		ctx.Redirect(http.StatusMovedPermanently, "/swagger/")
		return
	}

	// ensure we handle the swagger path correctly
	if file == "swagger/" {
		file = "swagger/index.html"
	}

	if !frontend.Box.Exists(file) {
		file = "index.html"
		ctx.HTML(http.StatusOK, file, gin.H{})
		return
	}

	data := frontend.Box.Get(file)
	contentTypes := map[string]string{
		".css":  "text/css",
		".html": "text/html",
		".ico":  "image/x-icon",
		".js":   "application/javascript",
		".json": "application/json",
		".png":  "image/png",
		".txt":  "text/plain",
	}

	for ext, contentType := range contentTypes {
		if strings.HasSuffix(file, ext) {
			ctx.Header("Content-Type", contentType)
		}
	}

	ctx.String(http.StatusOK, string(data))
}
