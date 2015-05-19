// Package serializer contains
package serializer

import (
	"fmt"
	"github.com/fatih/structs"
	"github.com/huandu/xstrings"
	"reflect"
)

type KeyCase int

const (
	NotSet     KeyCase = iota
	CamelCase          = iota
	PascalCase         = iota
	SnakeCase          = iota
)

var defaultCase KeyCase = NotSet

// Set the default key case (snake_case, camelCase or PascalCase) for
// all new serializers
func SetDefaultCase(caseType KeyCase) {
	defaultCase = caseType
}

type mapModifier func(jsonMap) jsonMap

type jsonMap map[string]interface{}
type Predicate func(interface{}) bool
type KeyConverter func(string) string
type ValueConverter func(interface{}) interface{}

type Serializer interface {
	// Transform the entity into a map[string]interface{} ready to be serialized
	Transform(entity interface{}) map[string]interface{}

	// Transform the entities into a []map[string]interface{} array ready to be serialized
	// entities must be a slice or an array
	TransformArray(entities interface{}) ([]map[string]interface{}, error)

	// Transform the entities into a []map[string]interface{} array ready to be serialized
	// Panics if entities is not a slice or an array
	MustTransformArray(entities interface{}) []map[string]interface{}

	// Convert all the keys using the given converter
	ConvertKeys(keyConverter KeyConverter) Serializer

	// Use snake_case keys
	UseSnakeCase() Serializer

	// Use camelCase keys
	UseCamelCase() Serializer

	// Use PascalCase keys
	UsePascalCase() Serializer

	// Add all the exported fields to the result
	PickAll() Serializer

	// Add the given fields to the result
	Pick(keys ...string) Serializer

	// Add the given fields to the result if the Predicate returns true
	PickIf(predicate Predicate, keys ...string) Serializer

	// Add the given fields to the result after applying the converter
	PickFunc(converter ValueConverter, keys ...string) Serializer

	// Add the given fields to the result after applying the converter if the predicate returns true
	PickFuncIf(predicate Predicate, converter ValueConverter, keys ...string) Serializer

	// Omit the given fields from the result
	Omit(keys ...string) Serializer

	// Omit the given fields from the result if the Predicate returns true
	OmitIf(predicate Predicate, keys ...string) Serializer

	// Add a custom field to the result
	Add(key string, value interface{}) Serializer

	// Add a custom field to the result if the Predicate returns true
	AddIf(predicate Predicate, key string, value interface{}) Serializer

	// Add a computed custom field to the result
	AddFunc(key string, converter ValueConverter) Serializer

	// Add a computed custom field to the result if the Predicate returns true
	AddFuncIf(predicate Predicate, key string, converter ValueConverter) Serializer
}

func alwaysTrue(u interface{}) bool {
	return true
}

func alwaysFalse(u interface{}) bool {
	return false
}

func identity(u interface{}) interface{} {
	return u
}

// A basic implementation of Serializer
type Base struct {
	raw          interface{}
	modifiers    []mapModifier
	reflected    reflect.Value
	keyConverter KeyConverter
}

// Creates a new serializer
func New() *Base {
	b := &Base{}
	b.addDefaultKeyConverter()
	return b
}

func (b *Base) Transform(entity interface{}) map[string]interface{} {
	b.raw = entity
	b.reflected = reflect.Indirect(reflect.ValueOf(entity))
	return b.result()
}

func (b *Base) TransformArray(entities interface{}) ([]map[string]interface{}, error) {
	s := reflect.ValueOf(entities)
	if s.Kind() != reflect.Slice && s.Kind() != reflect.Array {
		return nil, fmt.Errorf("TransformArray() given a non-slice type")
	}
	var result []map[string]interface{}
	for i := 0; i < s.Len(); i++ {
		result = append(result, b.Transform(s.Index(i).Interface()))
	}
	return result, nil
}

func (b *Base) MustTransformArray(entities interface{}) []map[string]interface{} {
	res, err := b.TransformArray(entities)
	if err != nil {
		panic(err)
	}
	return res
}

func (b *Base) addDefaultKeyConverter() {
	switch defaultCase {
	case PascalCase:
		b.UsePascalCase()
	case SnakeCase:
		b.UseSnakeCase()
	case CamelCase:
		b.UseCamelCase()
	default:
		break
	}
}

func (b *Base) transformedResult(result jsonMap) jsonMap {
	newResult := make(map[string]interface{})
	for key, value := range result {
		newResult[b.keyConverter(key)] = value
	}
	return newResult
}

func (b *Base) result() map[string]interface{} {
	result := make(map[string]interface{})
	for _, modifier := range b.modifiers {
		result = modifier(result)
	}
	if b.keyConverter != nil {
		return b.transformedResult(result)
	} else {
		return result
	}
}

func (b *Base) ConvertKeys(keyConverter KeyConverter) Serializer {
	b.keyConverter = keyConverter
	return b
}

func (b *Base) UsePascalCase() Serializer {
	return b.ConvertKeys(func(k string) string {
		return xstrings.ToCamelCase(k)
	})
}

func (b *Base) UseCamelCase() Serializer {
	return b.ConvertKeys(func(k string) string {
		return xstrings.FirstRuneToLower(xstrings.ToCamelCase(xstrings.ToSnakeCase(k)))
	})
}

func (b *Base) UseSnakeCase() Serializer {
	return b.ConvertKeys(xstrings.ToSnakeCase)
}

func (b *Base) PickAll() Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		return structs.Map(b.raw)
	})
	return b
}

func (b *Base) Pick(keys ...string) Serializer {
	return b.PickFunc(identity, keys...)
}

func (b *Base) PickIf(p Predicate, keys ...string) Serializer {
	return b.PickFuncIf(p, identity, keys...)
}

func (b *Base) PickFunc(converter ValueConverter, keys ...string) Serializer {
	return b.PickFuncIf(alwaysTrue, converter, keys...)
}

func (b *Base) PickFuncIf(p Predicate, converter ValueConverter, keys ...string) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		if p(b.raw) {
			for _, key := range keys {
				m[key] = converter(b.reflected.FieldByName(key).Interface())
			}
		}
		return m
	})
	return b
}

func (b *Base) Omit(keys ...string) Serializer {
	return b.OmitIf(alwaysTrue, keys...)
}

func (b *Base) OmitIf(p Predicate, keys ...string) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		if p(b.raw) {
			for _, key := range keys {
				delete(m, key)
			}
		}
		return m
	})
	return b
}

func (b *Base) Add(key string, value interface{}) Serializer {
	return b.AddIf(alwaysTrue, key, value)
}

func (b *Base) AddIf(p Predicate, key string, value interface{}) Serializer {
	return b.AddFuncIf(p, key, func(m interface{}) interface{} { return value })
}

func (b *Base) AddFunc(key string, f ValueConverter) Serializer {
	return b.AddFuncIf(alwaysTrue, key, f)
}

func (b *Base) AddFuncIf(p Predicate, key string, f ValueConverter) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		if p(b.raw) {
			m[key] = f(b.raw)
		}
		return m
	})
	return b
}
