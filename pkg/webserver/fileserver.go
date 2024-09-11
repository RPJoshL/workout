package webserver

import (
	"net/http"
	"strings"

	"git.rpjosh.de/RPJosh/go-logger"
	"git.rpjosh.de/RPJosh/workout/pkg/webserver/httprouter"
)

// FileServer serves the files from the embedded file system and registers
// it to the given path from the router
func FileServer(r *httprouter.Mux, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		logger.Debug("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", http.StatusMovedPermanently).ServeHTTP)
		path += "/"
	}
	routerPath := path + "*"

	r.Get(routerPath, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") {
			http.NotFound(w, r)
			return
		}

		fs := http.StripPrefix(path, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}
