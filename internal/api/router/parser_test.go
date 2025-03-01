package router

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type QParser1 struct {
	StrVal     string `query:"strVal"`
	IntVal     int    `query:"intVal"`
	FormIntVal int    `form:"intVal"`
	JsonVal    int    `json:"intValJson"`
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
		// Default mode is only query
		FormIntVal: 0,
	}
	req := getRequest(map[string]string{
		"strVal": exp.StrVal, "intVal": "600",
	}, map[string]string{
		"intVal": "700",
	})
	parser.Request = req

	err := parser.Parse(dst, RequestParserOptions{})
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
// are also parsed as they would be contained directly in the provided struct
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
	}, map[string]string{})
	parser.Request = req

	err := parser.Parse(dst, RequestParserOptions{})
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
	}, map[string]string{})
	parser.Request = req

	err := parser.Parse(dst, RequestParserOptions{})
	if err == nil {
		t.Error("Received no error while parsing invalid int", err)
	}
}

// TestModeForm tests the parsing of values based on the
// form tag and form values (mode = ParseModeForm)
func TestModeForm(t *testing.T) {
	parser := RequestParser{}

	dst := &QParser1{}
	exp := &QParser1{
		IntVal:     0,
		FormIntVal: 22,
	}
	req := getRequest(map[string]string{}, map[string]string{
		"intVal": "22",
	})
	parser.Request = req

	err := parser.Parse(dst, RequestParserOptions{
		Mode: ParseModeForm,
	})
	if err != nil {
		t.Errorf("Unexpected error for QueryParser: %s", err)
	}

	if diff := cmp.Diff(exp, dst); diff != "" {
		t.Errorf("Mismatch of query parser (-want +got):\n%s", diff)
	}
}

// TestParseJson tests the parsing of values based on the
// json tag and query values
func TestParseJson(t *testing.T) {
	parser := RequestParser{}

	dst := &QParser1{}
	exp := &QParser1{
		JsonVal: 20,
	}
	req := getRequest(map[string]string{
		"intValJson": "20",
	}, map[string]string{})
	parser.Request = req

	err := parser.Parse(dst, RequestParserOptions{
		InterpreteJson: true,
	})
	if err != nil {
		t.Errorf("Unexpected error for QueryParser: %s", err)
	}

	if diff := cmp.Diff(exp, dst); diff != "" {
		t.Errorf("Mismatch of query parser (-want +got):\n%s", diff)
	}
}

// getRequest builds a mock request with the provided query and
// form parameters
func getRequest(query map[string]string, form map[string]string) *http.Request {
	method := "GET"
	contentType := ""
	var body io.Reader

	if len(form) != 0 {
		method = "POST"

		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		for key, value := range form {
			writer.WriteField(key, value)
		}
		writer.Close()

		body = &buf
		contentType = writer.FormDataContentType()
	}

	req, _ := http.NewRequest(method, "/someData", body)
	q := req.URL.Query()
	for k, v := range query {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()

	// Set headers correctly
	req.Header.Set("Content-Type", contentType)

	return req
}
