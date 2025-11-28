package assert

import (
	"encoding/xml"
	"fmt"
	"slices"
	"strings"
	"testing"
)

type inNode node

type node struct {
	XMLName xml.Name
	Attrs   []xml.Attr `xml:",any,attr"`
	Content string     `xml:",innerxml"`
	Nodes   []node     `xml:",any"`
}

func (n *node) UnmarshalXML(d *xml.Decoder, start xml.StartElement) error {
	//nolint:musttag // We have to use an alias to avoid infinite recursion. False positive
	if err := d.DecodeElement((*inNode)(n), &start); err != nil {
		return fmt.Errorf("unable to decode element: %w", err)
	}
	if len(n.Nodes) > 0 {
		n.Content = ""
	}
	slices.SortStableFunc(n.Attrs, func(a, b xml.Attr) int {
		if a.Name.Space == b.Name.Space {
			return strings.Compare(a.Name.Local, b.Name.Local)
		}
		return strings.Compare(a.Name.Space, b.Name.Space)
	})
	return nil
}

func XMLEq(t *testing.T, expected, actual string, msgAndArgs ...any) {
	t.Helper()

	var expectedNode node
	if err := xml.Unmarshal([]byte(expected), &expectedNode); err != nil {
		t.Fatalf("Expected value (%s) is not valid xml. XML parsing error: '%s'", expected, err.Error())
		return
	}
	sortNodes(&expectedNode)

	var actualNode node
	if err := xml.Unmarshal([]byte(actual), &actualNode); err != nil {
		t.Fatalf("Input (%s) needs to be valid xml. XML parsing error: '%s'", actual, err.Error())
		return
	}
	sortNodes(&actualNode)

	EqualStruct(t, "XML", expectedNode, actualNode)
}

func sortNodes(nod *node) {
	slices.SortStableFunc(nod.Nodes, func(aaa, bbb node) int {
		if rt := strings.Compare(aaa.XMLName.Local, bbb.XMLName.Local); rt != 0 {
			return rt
		}

		return strings.Compare(aaa.Content, bbb.Content)
	})

	for i := range nod.Nodes {
		sortNodes(&nod.Nodes[i])
	}
}
