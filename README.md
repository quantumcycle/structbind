# StructBind

This project is a very simpler binder for structs in Golang.

## Installation

```
go get github.com/quantumcycle/structbind
```

## Example

```go
package main

import (
	"net/http"
    "github.com/quantumcycle/structbind"
)

type MyInputStruct struct {
	Param1  int    `bind:"query=param1"`
	Header1 string `bind:"header=header1"`
}

func main() {
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
    
    mockReq, _ := http.NewRequest("GET", "http://example.com?param1=23", nil)
    mockReq.Header.Set("header1", "test")
    
    result := MyInputStruct{
        Header1: "default value",
    }
    err = b.Bind(mockReq, &result)
	//...
}
```

## Usage

The `AddBinding` adds a new type of binding. You can then use the name in the `bind` tag of the struct to make use of that binding.

When registering a binding, you must provide a function that receives a name argument and a source type. 
The function must return the value of the binding or an error. It can also return nil. In that case, the original value 
in the struct will stay untouched. If you return an empty string, the field will be set to to the type default value.

This lib is using [github.com/mitchellh/mapstructure](https://github.com/mitchellh/mapstructure) underneath so refer to it for more information.

Here is an example using Gorilla MUX to extract path and query parameters
```golang
b := structbind.NewBinder[*http.Request]()
b.AddBinding("path", func(name string, req *http.Request) (any, error) {
    vars := mux.Vars(req)
	return vars[name]
})
b.AddBinding("query", func(name string, req *http.Request) (any, error) {
    return req.URL.Query().Get(name), nil
})
```

Then you could use it on a struct like this to automatically inject values from the request
```golang
type CategoryListingParams struct {
    Category string `bind:"path=category"`
    Page     int    `bind:"query=page"`
}
```

`mapstructure` is doing some weak binding, so it will automatically convert the string query param `page` into an int.
If the conversion fails, it will return an error.
