package structbind_test

import (
	"github.com/quantumcycle/structbind"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestStruct struct {
	Param1  int    `bind:"query=param1"`
	Header1 string `bind:"header=header1"`
}

type TestDifferentNamesStruct struct {
	IncludeValues bool `bind:"query=include_values"`
}

func createHttpRequestBinder() *structbind.Binder[*http.Request] {
	b := structbind.NewBinder[*http.Request]()
	b.AddBinding("query", func(name string, req *http.Request) (any, error) {
		return req.URL.Query().Get(name), nil
	})
	b.AddBinding("header", func(name string, req *http.Request) (any, error) {
		h := req.Header.Get(name)
		if h == "" {
			return nil, nil
		}
		return h, nil
	})
	return b
}

func TestBindAllValues(t *testing.T) {
	b := createHttpRequestBinder()
	mockReq, err := http.NewRequest("GET", "http://example.com?param1=23", nil)
	mockReq.Header.Set("header1", "override")
	if err != nil {
		t.Fatal(err)
	}

	result := TestStruct{
		Param1:  1,
		Header1: "default",
	}
	err = b.Bind(mockReq, &result)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, 23, result.Param1)
	assert.Equal(t, "override", result.Header1)
}

func TestBindShouldIgnoreNils(t *testing.T) {
	b := createHttpRequestBinder()
	mockReq, err := http.NewRequest("GET", "http://example.com?param1=23", nil)
	if err != nil {
		t.Fatal(err)
	}

	result := TestStruct{
		Param1:  1,
		Header1: "default",
	}
	err = b.Bind(mockReq, &result)
	if err != nil {
		t.Fatal(err)
	}

	// Header1 should be ignored when binding since the header binding function returns nil
	assert.Equal(t, "default", result.Header1)
}

func TestBindReturnAnErrorOnInvalidTypeBinding(t *testing.T) {
	b := createHttpRequestBinder()
	mockReq, err := http.NewRequest("GET", "http://example.com?param1=aaa", nil)
	if err != nil {
		t.Fatal(err)
	}

	result := TestStruct{
		Param1:  1,
		Header1: "default",
	}
	err = b.Bind(mockReq, &result)
	assert.Error(t, err)
}

func TestBindOnStructWithDifferentName(t *testing.T) {
	b := createHttpRequestBinder()
	mockReq, err := http.NewRequest("GET", "http://example.com?include_values=true", nil)
	if err != nil {
		t.Fatal(err)
	}

	result := TestDifferentNamesStruct{}
	err = b.Bind(mockReq, &result)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, true, result.IncludeValues)
}
