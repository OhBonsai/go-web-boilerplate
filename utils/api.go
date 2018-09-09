package utils

import (
	"net/http"
	"crypto"
	"net/url"
	"html/template"
	"github.com/OhBonsai/go-web-boilerplate/model"
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func RenderWebAppError(w http.ResponseWriter, r *http.Request, err *model.AppError, s crypto.Signer) {
	RenderWebError(w, r, err.StatusCode, url.Values{
		"message": []string{err.Message},
	}, s)
}

func RenderWebError(w http.ResponseWriter, r *http.Request, status int, params url.Values, s crypto.Signer) {
	queryString := params.Encode()

	h := crypto.SHA256
	sum := h.New()
	sum.Write([]byte("/error?" + queryString))
	signature, err := s.Sign(rand.Reader, sum.Sum(nil), h)
	if err != nil {
		http.Error(w, "", http.StatusInternalServerError)
		return
	}
	destination := "/error?" + queryString + "&s=" + base64.URLEncoding.EncodeToString(signature)

	if status >= 300 && status < 400 {
		http.Redirect(w, r, destination, status)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	fmt.Fprintln(w, `<!DOCTYPE html><html><head></head>`)
	fmt.Fprintln(w, `<body onload="window.location = '`+template.HTMLEscapeString(template.JSEscapeString(destination))+`'">`)
	fmt.Fprintln(w, `<noscript><meta http-equiv="refresh" content="0; url=`+template.HTMLEscapeString(destination)+`"></noscript>`)
	fmt.Fprintln(w, `<a href="`+template.HTMLEscapeString(destination)+`" style="color: #c0c0c0;">...</a>`)
	fmt.Fprintln(w, `</body></html>`)
}

