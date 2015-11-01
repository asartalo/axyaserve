package axyaserve

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestInjector(t *testing.T) {
	Convey("Given a handler", t, func() {
		var contentType string
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "http://example.com/foo", nil)

		orig := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", contentType)
			fmt.Fprint(w, "<html><head></head><body>Hello, client</body></html>")
		})

		inj := Injector(orig)
		inj.Inject("text/html", func(content string) string {
			return content + "FOO"
		})

		Convey("When a request is sent", func() {
			contentType = "text/html; charset=utf-8"
			inj.ServeHTTP(w, req)
			content, _ := ioutil.ReadAll(w.Body)

			Convey("Its response should be injected with code", func() {
				expected := "<html><head></head><body>Hello, client</body></html>FOO"
				So(string(content), ShouldEqual, expected)
			})

		})

		Convey("When response is of a different type", func() {
			contentType = "text/plain; charset=utf-8"
			inj.ServeHTTP(w, req)
			content, _ := ioutil.ReadAll(w.Body)

			Convey("Its response should be not be injected", func() {
				So(string(content), ShouldEqual, string(content))
			})
		})
	})
}

func TestLiveReloadInjectorFunc(t *testing.T) {
	Convey("Given an html content", t, func() {
		content := "<html><head></head><body>Hello, client</body></html>"

		Convey("When passed to livereload injector", func() {
			result := InjectLiveReload(content)

			Convey("It will contain livereload script", func() {
				expected := "<html><head></head><body>Hello, client<script src=\"http://localhost:35729/livereload.js\"></script></body></html>"
				So(result, ShouldEqual, expected)
			})
		})
	})
}
