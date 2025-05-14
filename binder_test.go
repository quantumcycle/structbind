package structbind_test

import (
	"encoding/json"
	"github.com/quantumcycle/structbind"
	"net/http"
	"reflect"
	"strings"
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

type TestInputWithBody struct {
	Body Person `bind:"body"`
}

type ToBeEmbedded struct {
	SubParam1 string `bind:"query=sub_param1"`
}

type TestStructWithEmbedded struct {
	Embedded ToBeEmbedded
	Param1   int    `bind:"query=param1"`
	Header1  string `bind:"header=header1"`
}

type Person struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func createHttpRequestBinder() *structbind.Binder[*http.Request] {
	b := structbind.NewBinder[*http.Request]()
	b.AddBinding("query", func(name string, targetType reflect.Type, req *http.Request) (any, error) {
		return req.URL.Query().Get(name), nil
	})
	b.AddBinding("header", func(name string, targetType reflect.Type, req *http.Request) (any, error) {
		h := req.Header.Get(name)
		if h == "" {
			return nil, nil
		}
		return h, nil
	})
	b.AddBinding("body", func(hint string, targetType reflect.Type, req *http.Request) (any, error) {
		j := reflect.New(targetType).Interface()
		err := json.NewDecoder(req.Body).Decode(&j)
		if err != nil {
			return nil, err
		}
		return j, nil
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

func TestBindAllValuesWithEmbedding(t *testing.T) {
	b := createHttpRequestBinder()
	mockReq, err := http.NewRequest("GET", "http://example.com?param1=23&sub_param1=subvalue1", nil)
	mockReq.Header.Set("header1", "override")
	if err != nil {
		t.Fatal(err)
	}

	result := TestStructWithEmbedded{
		Embedded: ToBeEmbedded{
			SubParam1: "",
		},
		Param1:  1,
		Header1: "default",
	}
	err = b.Bind(mockReq, &result)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "subvalue1", result.Embedded.SubParam1)
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

func TestBindCanUseFieldType(t *testing.T) {
	b := createHttpRequestBinder()
	mockReq, err := http.NewRequest("POST", "http://example.com", strings.NewReader("{\"field1\":\"value\",\"field2\":42}"))
	if err != nil {
		t.Fatal(err)
	}

	result := TestInputWithBody{}
	err = b.Bind(mockReq, &result)
	if err != nil {
		t.Fatal(err)
	}

	assert.Equal(t, "value", result.Body.Field1)
	assert.Equal(t, 42, result.Body.Field2)
}
