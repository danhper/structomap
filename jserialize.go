// Package serializer contains
package serializer

import (
	"github.com/fatih/structs"
	"reflect"
)

type mapModifier func(jsonMap) jsonMap

type jsonMap map[string]interface{}
type predicate func(interface{}) bool
type converter func(interface{}) interface{}

type serializer struct {
	raw       interface{}
	modifiers []mapModifier
	reflected reflect.Value
}

// Creates a new serializer
func New(entity interface{}) *serializer {
	return &serializer{
		raw:       entity,
		reflected: reflect.Indirect(reflect.ValueOf(entity)),
	}
}

// Returns the result of the serialization as a map[string]interface{}
func (s *serializer) Result() jsonMap {
	result := make(map[string]interface{})
	for _, modifier := range s.modifiers {
		result = modifier(result)
	}
	return result
}

// Add all the exported fields to the result
func (s *serializer) PickAll() *serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		return structs.Map(s.raw)
	})
	return s
}

// Add the given fields to the result
func (s *serializer) Pick(fields ...string) *serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		for _, field := range fields {
			m[field] = s.reflected.FieldByName(field).Interface()
		}
		return m
	})
	return s
}

// Add the given fields to the result if the predicate returns true
func (s *serializer) PickIf(p predicate, fields ...string) *serializer {
	if p(s.raw) {
		return s.Pick(fields...)
	}
	return s
}

// Omit the given fields from the result
func (s *serializer) Omit(fields ...string) *serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		for _, field := range fields {
			delete(m, field)
		}
		return m
	})
	return s
}

// Omit the given fields from the result if the predicate returns true
func (s *serializer) OmitIf(p predicate, fields ...string) *serializer {
	if p(s.raw) {
		return s.Omit(fields...)
	}
	return s
}

// Add a custom field to the result
func (s *serializer) Add(field string, value interface{}) *serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = value
		return m
	})
	return s
}

// Add a custom field to the result if the predicate returns true
func (s *serializer) AddIf(p predicate, field string, value interface{}) *serializer {
	if p(s.raw) {
		return s.Add(field, value)
	}
	return s
}

// Add a computed custom field to the result
func (s *serializer) AddFunc(field string, f converter) *serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = f(s.raw)
		return m
	})
	return s
}

// Add a computed custom field to the result if the predicate returns true
func (s *serializer) AddFuncIf(p predicate, field string, f converter) *serializer {
	if p(s.raw) {
		return s.AddFunc(field, f)
	}
	return s
}

// Convert the field using the given converter
func (s *serializer) Convert(field string, f converter) *serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = f(s.reflected.FieldByName(field).Interface())
		return m
	})
	return s
}

// Convert the field using the given converter if the predicate returns true
func (s *serializer) ConvertIf(p predicate, field string, f converter) *serializer {
	if p(s.raw) {
		return s.Convert(field, f)
	}
	return s
}
