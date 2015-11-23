package axyaserve

import (
	"fmt"
	"mime"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"

	"github.com/asartalo/go-html-transform/h5"
	"github.com/asartalo/go-html-transform/html/transform"
)

type injectorFunc func(string) string

type Injector struct {
	original  http.Handler
	injectors map[string]injectorFunc
}

func NewInjector(original http.Handler) *Injector {
	return &Injector{original, make(map[string]injectorFunc)}
}

func (lr *Injector) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var responseContentType string
	var newContent string
	var contentLength int
	wr := httptest.NewRecorder()

	fmt.Println("Request: ", r.Method, r.URL)

	lr.original.ServeHTTP(wr, r)
	content := wr.Body.String()
	for key, values := range wr.Header() {
		for _, value := range values {
			if key == "Content-Type" {
				responseContentType = value
			}
			w.Header().Add(key, value)
		}
	}
	newContent = content
	contentType, _, err := mime.ParseMediaType(responseContentType)
	if err == nil {
		injectorFunc, ok := lr.injectors[contentType]
		if ok {
			newContent = injectorFunc(string(content))
			contentLength = len([]byte(newContent))
			cLengthValue := w.Header().Get("Content-Length")
			if cLengthValue != "" {
				w.Header().Set("Content-Length", strconv.Itoa(contentLength))
			}
		}
	}
	w.WriteHeader(wr.Code)
	fmt.Println("Response:", wr.Code, contentType)
	w.Write([]byte(newContent))
}

func (lr *Injector) Inject(mimetype string, injectionFunc func(string) string) {
	lr.injectors[mimetype] = injectionFunc
}

func InjectLiveReload(html string) string {
	tree, err := h5.New(strings.NewReader(html))
	if err != nil {
		return html
	}
	t := transform.New(tree)
	script, _ := h5.PartialFromString("<script src=\"http://localhost:35729/livereload.js\"></script>")
	t.Apply(transform.AppendChildren(script...), "body")
	return t.String()
}
