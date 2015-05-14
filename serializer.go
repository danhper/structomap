// Package serializer contains
package serializer

import (
	"github.com/fatih/structs"
	"reflect"
)

type mapModifier func(jsonMap) jsonMap

type jsonMap map[string]interface{}
type Predicate func(interface{}) bool
type Converter func(interface{}) interface{}

type Serializer interface {
	// Returns the result of the serialization as a map[string]interface{}
	Result() map[string]interface{}

	// Add all the exported fields to the result
	PickAll() Serializer

	// Add the given fields to the result
	Pick(fields ...string) Serializer

	// Add the given fields to the result if the Predicate returns true
	PickIf(predicate Predicate, fields ...string) Serializer

	// Omit the given fields from the result
	Omit(fields ...string) Serializer

	// Omit the given fields from the result if the Predicate returns true
	OmitIf(predicate Predicate, fields ...string) Serializer

	// Add a custom field to the result
	Add(key string, value interface{}) Serializer

	// Add a custom field to the result if the Predicate returns true
	AddIf(predicate Predicate, key string, value interface{}) Serializer

	// Add a computed custom field to the result
	AddFunc(key string, converter Converter) Serializer

	// Add a computed custom field to the result if the Predicate returns true
	AddFuncIf(predicate Predicate, key string, converter Converter) Serializer

	// Convert the field using the given converter
	Convert(field string, converter Converter) Serializer

	// Convert the field using the given converter if the Predicate returns true
	ConvertIf(predicate Predicate, field string, converter Converter) Serializer
}

// A basic implementation of Serializer
type Base struct {
	raw       interface{}
	modifiers []mapModifier
	reflected reflect.Value
}

// Creates a new serializer
func New(entity interface{}) *Base {
	return &Base{
		raw:       entity,
		reflected: reflect.Indirect(reflect.ValueOf(entity)),
	}
}

func (s *Base) Result() map[string]interface{} {
	result := make(map[string]interface{})
	for _, modifier := range s.modifiers {
		result = modifier(result)
	}
	return result
}

func (s *Base) PickAll() Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		return structs.Map(s.raw)
	})
	return s
}

func (s *Base) Pick(fields ...string) Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		for _, field := range fields {
			m[field] = s.reflected.FieldByName(field).Interface()
		}
		return m
	})
	return s
}

func (s *Base) PickIf(p Predicate, fields ...string) Serializer {
	if p(s.raw) {
		return s.Pick(fields...)
	}
	return s
}

func (s *Base) Omit(fields ...string) Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		for _, field := range fields {
			delete(m, field)
		}
		return m
	})
	return s
}

func (s *Base) OmitIf(p Predicate, fields ...string) Serializer {
	if p(s.raw) {
		return s.Omit(fields...)
	}
	return s
}

func (s *Base) Add(field string, value interface{}) Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = value
		return m
	})
	return s
}

func (s *Base) AddIf(p Predicate, field string, value interface{}) Serializer {
	if p(s.raw) {
		return s.Add(field, value)
	}
	return s
}

func (s *Base) AddFunc(field string, f Converter) Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = f(s.raw)
		return m
	})
	return s
}

func (s *Base) AddFuncIf(p Predicate, field string, f Converter) Serializer {
	if p(s.raw) {
		return s.AddFunc(field, f)
	}
	return s
}

func (s *Base) Convert(field string, f Converter) Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = f(s.reflected.FieldByName(field).Interface())
		return m
	})
	return s
}

func (s *Base) ConvertIf(p Predicate, field string, f Converter) Serializer {
	if p(s.raw) {
		return s.Convert(field, f)
	}
	return s
}
