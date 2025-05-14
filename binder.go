package structbind

import (
	"fmt"
	"github.com/mitchellh/mapstructure"
	"reflect"
	"strings"
)

type Extractor[T any] func(hint string, targetType reflect.Type, source T) (any, error)

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

	err := b.processFields(req, targetValue, typeOfTarget, &mapValues)
	if err != nil {
		return nil, err
	}

	return mapValues, nil
}

func (b *Binder[T]) processFields(req T, targetValue reflect.Value, typeOfTarget reflect.Type, mapValues *map[string]any) error {
	for i := 0; i < targetValue.NumField(); i++ {
		field := typeOfTarget.Field(i)

		// Process regular fields with "bind" tags
		fieldTag := field.Tag
		if bindDefinition, ok := fieldTag.Lookup("bind"); ok {
			err := b.extractValue(bindDefinition, field.Name, field.Type, req, mapValues)
			if err != nil {
				return err
			}
			continue
		}

		//If the field doesnt have a bind definition and it's a struct, we deep dive
		if field.Type.Kind() == reflect.Struct {
			embeddedValue := targetValue.Field(i)
			embeddedType := field.Type
			embeddedValues := make(map[string]any)
			(*mapValues)[field.Name] = embeddedValues
			err := b.processFields(req, embeddedValue, embeddedType, &embeddedValues)
			if err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func (b *Binder[T]) extractValue(definition, fieldName string, targetType reflect.Type, req T, m *map[string]any) error {
	binding, hint, err := b.parseDefinition(definition)
	if err != nil {
		return err
	}
	extractor, found := b.bindings[binding]
	if !found {
		return fmt.Errorf("no registered binding for: %s", binding)
	}
	value, err := extractor(hint, targetType, req)
	if err != nil {
		return err
	}
	if value != nil {
		(*m)[fieldName] = value
	}
	return nil
}

func (b *Binder[T]) parseDefinition(definition string) (string, string, error) {
	strs := strings.Split(definition, "=")
	if len(strs) == 1 {
		//no hint
		return strs[0], "", nil
	}
	if len(strs) == 2 {
		return strs[0], strs[1], nil
	}
	return "", "", fmt.Errorf("invalid binding definition: %s. should be [binding=hint]", definition)
}
