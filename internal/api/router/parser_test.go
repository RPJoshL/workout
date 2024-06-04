package router

import (
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type QParser1 struct {
	StrVal string `query:"strVal"`
	IntVal int    `query:"intVal"`
}

type QueryParserEmbedded struct {
	QParser1
	DirectValue string `query:"directVal"`
}

// TestQueryParserValid tests the parsing of a request with
// valid query parameters
func TestQueryParser(t *testing.T) {
	parser := RequestParser{}

	// Default behaviour
	dst := &QParser1{}
	exp := &QParser1{
		StrVal: "Hello World!",
		IntVal: 600,
	}
	req := getRequest(map[string]string{
		"strVal": exp.StrVal, "intVal": "600",
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

// TestQueryParserEmbedded tests the parsing of a request with
// an embedded struct.
//
// It's expected that all query parameters from the embedded struct
// are also parsed as the would be contained directly in the provided struct
func TestQueryParserEmbedded(t *testing.T) {
	parser := RequestParser{}

	// Default behaviour
	dst := &QueryParserEmbedded{}
	exp := &QueryParserEmbedded{
		QParser1: QParser1{
			StrVal: "Hello World!",
		},
		DirectValue: "It's directly here :)",
	}
	req := getRequest(map[string]string{
		"strVal": exp.StrVal, "directVal": exp.DirectValue,
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

// TestQueryParseErr tests the parsing of invalid query parameters
func TestQueryParseErr(t *testing.T) {
	parser := RequestParser{}

	// Default behaviour
	dst := &QParser1{}
	req := getRequest(map[string]string{
		"intVal": "inval",
	})
	parser.Request = req

	err := parser.Parse(dst)
	if err == nil {
		t.Error("Received no error while parsing invalid int", err)
	}
}

// getRequest builds a mock request with the provided query parameters
func getRequest(query map[string]string) *http.Request {
	req, _ := http.NewRequest("GET", "/someData", nil)
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	return req
}
