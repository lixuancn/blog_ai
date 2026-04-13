package http

import (
	"embed"
	"net/http"
)

//go:embed frontassets/front.css frontassets/front.js
var embeddedFrontAssets embed.FS

func serveFrontCSS(w http.ResponseWriter, r *http.Request) {
	serveFrontAsset(w, r, "frontassets/front.css", "text/css; charset=utf-8")
}

func serveFrontJS(w http.ResponseWriter, r *http.Request) {
	serveFrontAsset(w, r, "frontassets/front.js", "application/javascript; charset=utf-8")
}

func serveFrontAsset(w http.ResponseWriter, _ *http.Request, name, contentType string) {
	raw, err := embeddedFrontAssets.ReadFile(name)
	if err != nil {
		http.Error(w, "asset not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", contentType)
	_, _ = w.Write(raw)
}
