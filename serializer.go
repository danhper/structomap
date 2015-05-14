// Package serializer contains
package serializer

import (
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
	// Returns the result of the serialization as a map[string]interface{}
	Result() map[string]interface{}

	// Transform all the keys using the given converter
	TransformKeys(keyConverter KeyConverter) Serializer

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

	// Convert the field using the given converter
	Convert(key string, converter ValueConverter) Serializer

	// Convert the field using the given converter if the Predicate returns true
	ConvertIf(predicate Predicate, key string, converter ValueConverter) Serializer
}

// A basic implementation of Serializer
type Base struct {
	raw          interface{}
	modifiers    []mapModifier
	reflected    reflect.Value
	keyConverter KeyConverter
}

// Creates a new serializer
func New(entity interface{}) *Base {
	b := &Base{
		raw:       entity,
		reflected: reflect.Indirect(reflect.ValueOf(entity)),
	}
	b.addDefaultKeyConverter()
	return b
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

func (b *Base) Result() map[string]interface{} {
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

func (b *Base) TransformKeys(keyConverter KeyConverter) Serializer {
	b.keyConverter = keyConverter
	return b
}

func (b *Base) UsePascalCase() Serializer {
	return b.TransformKeys(func(k string) string {
		return xstrings.ToCamelCase(k)
	})
}

func (b *Base) UseCamelCase() Serializer {
	return b.TransformKeys(func(k string) string {
		return xstrings.FirstRuneToLower(xstrings.ToCamelCase(xstrings.ToSnakeCase(k)))
	})
}

func (b *Base) UseSnakeCase() Serializer {
	return b.TransformKeys(xstrings.ToSnakeCase)
}

func (b *Base) PickAll() Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		return structs.Map(b.raw)
	})
	return b
}

func (b *Base) Pick(keys ...string) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		for _, key := range keys {
			m[key] = b.reflected.FieldByName(key).Interface()
		}
		return m
	})
	return b
}

func (b *Base) PickIf(p Predicate, keys ...string) Serializer {
	if p(b.raw) {
		return b.Pick(keys...)
	}
	return b
}

func (b *Base) Omit(keys ...string) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		for _, key := range keys {
			delete(m, key)
		}
		return m
	})
	return b
}

func (b *Base) OmitIf(p Predicate, keys ...string) Serializer {
	if p(b.raw) {
		return b.Omit(keys...)
	}
	return b
}

func (b *Base) Add(key string, value interface{}) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		m[key] = value
		return m
	})
	return b
}

func (b *Base) AddIf(p Predicate, key string, value interface{}) Serializer {
	if p(b.raw) {
		return b.Add(key, value)
	}
	return b
}

func (b *Base) AddFunc(key string, f ValueConverter) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		m[key] = f(b.raw)
		return m
	})
	return b
}

func (b *Base) AddFuncIf(p Predicate, key string, f ValueConverter) Serializer {
	if p(b.raw) {
		return b.AddFunc(key, f)
	}
	return b
}

func (b *Base) Convert(key string, f ValueConverter) Serializer {
	b.modifiers = append(b.modifiers, func(m jsonMap) jsonMap {
		m[key] = f(b.reflected.FieldByName(key).Interface())
		return m
	})
	return b
}

func (b *Base) ConvertIf(p Predicate, key string, f ValueConverter) Serializer {
	if p(b.raw) {
		return b.Convert(key, f)
	}
	return b
}
