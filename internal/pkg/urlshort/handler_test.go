package urlshort

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

const defaultBody = "Hello, testing world!"

type handlerTest struct {
	short string
	long  string
}

var handlerTests = []handlerTest{
	{
		"/urlshort-godoc",
		"https://godoc.org/github.com/gophercises/urlshort",
	},
	{
		"/yaml-godoc",
		"https://godoc.org/gopkg.in/yaml.v2",
	},
}

func TestMapHandlerHappyPath(t *testing.T) {
	mux := defaultMux()
	pathsToUrls := map[string]string{}
	for _, ht := range handlerTests {
		pathsToUrls[ht.short] = ht.long
	}
	mapHandler := MapHandler(pathsToUrls, mux)

	ts := httptest.NewServer(mapHandler)
	defer ts.Close()

	// The body we get back from the short URL should match
	// the body we get back from the long URL.
	for _, ht := range handlerTests {
		shortBody := getURLBody(t, ts.URL+ht.short)
		fullBody := getURLBody(t, ht.long)
		checkResults(t, shortBody, fullBody)
	}
}

func TestMapHandlerNotFound(t *testing.T) {
	mux := defaultMux()
	pathsToUrls := map[string]string{
		"/urlshort-godoc": "https://godoc.org/github.com/gophercises/urlshort",
		"/yaml-godoc":     "https://godoc.org/gopkg.in/yaml.v2",
	}
	mapHandler := MapHandler(pathsToUrls, mux)

	ts := httptest.NewServer(mapHandler)
	defer ts.Close()

	// Key isn't in pathsToUrls, so we should hit
	// the fallback page and get defaultBody
	body := getURLBody(t, ts.URL+"/no such URL")
	checkResults(t, body, defaultBody)
}

func TestYamlHandlerHappyPath(t *testing.T) {
	mux := defaultMux()
	template := `
- path: %s
  url: %s
`
	yaml := ""
	for _, ht := range handlerTests {
		yaml += fmt.Sprintf(template, ht.short, ht.long)
	}
	yamlHandler, err := YAMLHandler([]byte(yaml), mux)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(yamlHandler)
	defer ts.Close()

	// The body we get back from the short URL should match
	// the body we get back from the long URL.
	for _, ht := range handlerTests {
		shortBody := getURLBody(t, ts.URL+ht.short)
		fullBody := getURLBody(t, ht.long)
		checkResults(t, shortBody, fullBody)
	}
}

func TestYamlHandlerNotFound(t *testing.T) {
	mux := defaultMux()
	yaml := `
- path: /urlshort
  url: https://github.com/gophercises/urlshort
- path: /urlshort-final
  url: https://github.com/gophercises/urlshort/tree/solution
`
	yamlHandler, err := YAMLHandler([]byte(yaml), mux)
	if err != nil {
		t.Fatal(err)
	}

	ts := httptest.NewServer(yamlHandler)
	defer ts.Close()

	// Key isn't in yaml, so we should hit
	// the fallback page and get defaultBody
	body := getURLBody(t, ts.URL+"/no such URL")
	checkResults(t, body, defaultBody)
}

func getURLBody(t *testing.T, URL string) string {
	resp, err := http.Get(URL)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	// Get automatically follows redirect, so we can always check for StatusOK
	if status := resp.StatusCode; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusFound)
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(body))
}

func checkResults(t *testing.T, body1, body2 string) {
	if body1 != body2 {
		t.Errorf("expected match: [%s] [%s]", body1, body2)
	}
}

func defaultMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", helloTest)
	return mux
}

func helloTest(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, defaultBody)
}
