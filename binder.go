package structbind

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

type Extractor[T any] func(name string, source T) (any, error)

type Binder[T any] struct {
	bindings map[string]Extractor[T]
}

func NewBinder[T any]() *Binder[T] {
	return &Binder[T]{make(map[string]Extractor[T])}
}

func (b *Binder[T]) AddBinding(name string, fn Extractor[T]) *Binder[T] {
	b.bindings[name] = fn
	return b
}

func (b *Binder[T]) Bind(req T, target interface{}) error {
	values, err := b.buildValues(req, target)
	if err != nil {
		return err
	}

	err = mapstructure.WeakDecode(values, target)
	if err != nil {
		return err
	}

	return nil
}

func (b *Binder[T]) buildValues(req T, target interface{}) (map[string]any, error) {
	mapValues := make(map[string]any)
	targetValue := reflect.ValueOf(target).Elem()
	typeOfTarget := targetValue.Type()
	for i := 0; i < targetValue.NumField(); i++ {
		fieldTag := typeOfTarget.Field(i).Tag
		if bindDefinition, ok := fieldTag.Lookup("bind"); ok {
			err := b.extractValue(bindDefinition, req, &mapValues)
			if err != nil {
				return nil, err
			}
		}
	}
	return mapValues, nil
}

func (b *Binder[T]) extractValue(definition string, req T, m *map[string]any) error {
	binding, name, err := b.parseDefinition(definition)
	if err != nil {
		return err
	}
	extractor, found := b.bindings[binding]
	if !found {
		return fmt.Errorf("no registered binding for: %s", binding)
	}
	value, err := extractor(name, req)
	if err != nil {
		return err
	}
	if value != nil {
		(*m)[name] = value
	}
	return nil
}

func (b *Binder[T]) parseDefinition(definition string) (string, string, error) {
	strs := strings.Split(definition, "=")
	if len(strs) != 2 {
		return "", "", fmt.Errorf("invalid binding definition: %s. should be [binding=name]", definition)
	}
	return strs[0], strs[1], nil
}