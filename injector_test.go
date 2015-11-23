package axyaserve_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

	ax "github.com/asartalo/axyaserve"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Injector", func() {
	fmt.Println("YOYOO")
	Context("Given a handler", func() {
		var contentType string
		var w *httptest.ResponseRecorder
		var req *http.Request
		var orig http.HandlerFunc
		var inj *ax.Injector

		var content []byte

		BeforeEach(func() {
			w = httptest.NewRecorder()
			req, _ = http.NewRequest("GET", "http://example.com/foo", nil)

			orig = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", contentType)
				fmt.Fprint(w, "<html><head></head><body>Hello, client</body></html>")
			})

			inj = ax.NewInjector(orig)
			inj.Inject("text/html", func(content string) string {
				return content + "FOO"
			})
		})

		Context("When a request is sent", func() {
			BeforeEach(func() {
				contentType = "text/html; charset=utf-8"
				inj.ServeHTTP(w, req)
				content, _ = ioutil.ReadAll(w.Body)
			})

			It("Injects code to response", func() {
				expected := "<html><head></head><body>Hello, client</body></html>FOO"
				Expect(string(content)).To(Equal(expected))
			})

		})

		Context("When response is of a different type", func() {
			BeforeEach(func() {
				contentType = "text/plain; charset=utf-8"
				inj.ServeHTTP(w, req)
				content, _ = ioutil.ReadAll(w.Body)
			})

			It("Its response should be not be injected", func() {
				Expect(string(content)).To(Equal(string(content)))
			})
		})
	})

	Context("Given an html content", func() {
		var strcontent string
		var result string

		Context("When passed to livereload injector", func() {
			BeforeEach(func() {
				strcontent = "<html><head></head><body>Hello, client</body></html>"
				result = ax.InjectLiveReload(strcontent)
			})

			It("It will contain livereload script", func() {
				expected := "<html><head></head><body>Hello, client<script src=\"http://localhost:35729/livereload.js\"></script></body></html>"
				fmt.Println(result, expected)
				Expect(result).To(Equal(expected))
			})
		})
	})
})
