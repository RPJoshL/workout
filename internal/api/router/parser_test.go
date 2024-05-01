package router

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type QueryParser struct {
	StrVal string `query:"strVal"`
	IntVal int    `query:"intVal"`
}

// TestQueryParserValid tests the parsing of a request with
// query parameters
func TestQueryParser(t *testing.T) {
	parser := RequestParser{}

	// Default behaviour
	dst := &QueryParser{}
	exp := &QueryParser{
		StrVal: "Hello world",
		IntVal: 800,
	}
	req := getRequest(map[string]string{
		"strVal": exp.StrVal, "intVal": "800",
	})
	parser.Request = req

	err := parser.Parse(dst)
	if err != nil {
		t.Errorf("Unexpected error for QueryParser: %s", err)
	}

	if diff := cmp.Diff(exp, dst); diff != "" {
		t.Errorf("Mismatch of query parser (-want +got):\n%s", diff)
	}
}

func TestQueryParseErr(t *testing.T) {
	parser := RequestParser{}

	// Default behaviour
	dst := &QueryParser{}
	req := getRequest(map[string]string{
		"intVal": "555inv555",
	})
	parser.Request = req

	err := parser.Parse(dst)
	if err == nil {
		t.Error("Received no error while parsing invalid int", err)
	}
}

func getRequest(query map[string]string) *http.Request {
	req, _ := http.NewRequest("GET", "/someData", nil)
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return req
}
