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

type Serializer struct {
	raw       interface{}
	modifiers []mapModifier
	reflected reflect.Value
}

// Creates a new Serializer
func New(entity interface{}) *Serializer {
	return &Serializer{
		raw:       entity,
		reflected: reflect.Indirect(reflect.ValueOf(entity)),
	}
}

// Returns the result of the serialization as a map[string]interface{}
func (s *Serializer) Result() jsonMap {
	result := make(map[string]interface{})
	for _, modifier := range s.modifiers {
		result = modifier(result)
	}
	return result
}

// Add all the exported fields to the result
func (s *Serializer) PickAll() *Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		return structs.Map(s.raw)
	})
	return s
}

// Add the given fields to the result
func (s *Serializer) Pick(fields ...string) *Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		for _, field := range fields {
			m[field] = s.reflected.FieldByName(field).Interface()
		}
		return m
	})
	return s
}

// Add the given fields to the result if the predicate returns true
func (s *Serializer) PickIf(p predicate, fields ...string) *Serializer {
	if p(s.raw) {
		return s.Pick(fields...)
	}
	return s
}

// Omit the given fields from the result
func (s *Serializer) Omit(fields ...string) *Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		for _, field := range fields {
			delete(m, field)
		}
		return m
	})
	return s
}

// Omit the given fields from the result if the predicate returns true
func (s *Serializer) OmitIf(p predicate, fields ...string) *Serializer {
	if p(s.raw) {
		return s.Omit(fields...)
	}
	return s
}

// Add a custom field to the result
func (s *Serializer) Add(field string, value interface{}) *Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = value
		return m
	})
	return s
}

// Add a custom field to the result if the predicate returns true
func (s *Serializer) AddIf(p predicate, field string, value interface{}) *Serializer {
	if p(s.raw) {
		return s.Add(field, value)
	}
	return s
}

// Add a computed custom field to the result
func (s *Serializer) AddFunc(field string, f converter) *Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = f(s.raw)
		return m
	})
	return s
}

// Add a computed custom field to the result if the predicate returns true
func (s *Serializer) AddFuncIf(p predicate, field string, f converter) *Serializer {
	if p(s.raw) {
		return s.AddFunc(field, f)
	}
	return s
}

// Convert the field using the given converter
func (s *Serializer) Convert(field string, f converter) *Serializer {
	s.modifiers = append(s.modifiers, func(m jsonMap) jsonMap {
		m[field] = f(s.reflected.FieldByName(field).Interface())
		return m
	})
	return s
}

// Convert the field using the given converter if the predicate returns true
func (s *Serializer) ConvertIf(p predicate, field string, f converter) *Serializer {
	if p(s.raw) {
		return s.Convert(field, f)
	}
	return s
}
