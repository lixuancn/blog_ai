package http

import (
	"embed"
	"net/http"
)

//go:embed uiassets/app.css uiassets/app.js
var embeddedAdminAssets embed.FS

func serveAdminCSS(w http.ResponseWriter, r *http.Request) {
	serveAdminAsset(w, r, "uiassets/app.css", "text/css; charset=utf-8")
}

func serveAdminJS(w http.ResponseWriter, r *http.Request) {
	serveAdminAsset(w, r, "uiassets/app.js", "application/javascript; charset=utf-8")
}

func serveAdminAsset(w http.ResponseWriter, _ *http.Request, name, contentType string) {
	raw, err := embeddedAdminAssets.ReadFile(name)
	if err != nil {
		http.Error(w, "asset not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(raw)
}
