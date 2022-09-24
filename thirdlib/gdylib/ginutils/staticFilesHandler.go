package ginutils

import (
	"io/fs"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func HandlerStaticFiles(files fs.FS) gin.HandlerFunc {
	fileServer := http.FileServer(http.FS(files))
	return func(c *gin.Context) {
		staticFile := isStaticFile(http.FS(files), c.Request.URL.Path, true)
		if staticFile {
			fileServer.ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		c.Next()
	}
}

func isStaticFile(fs http.FileSystem, name string, redirect bool) (isFile bool) {
	const indexPage = "/index.html"
	if strings.HasSuffix(name, indexPage) {
		return true
	}
	f, err := fs.Open(name)
	if err != nil {
		return false
	}
	defer f.Close()
	_, err = f.Stat()
	return err == nil
}
