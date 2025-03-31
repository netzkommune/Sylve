package utils

import (
	"errors"
	"testing"

	"github.com/go-playground/validator/v10"
)

type SampleInput struct {
	Name  string `json:"name" validate:"required"`
	Email string `json:"email" validate:"required,email"`
	Age   int    `json:"age"`
}

func TestGetJSONFieldName(t *testing.T) {
	type testStruct struct {
		Foo string `json:"foo"`
		Bar string // no json tag
	}

	tests := []struct {
		name       string
		inputField string
		expected   string
	}{
		{"with json tag", "Foo", "foo"},
		{"without json tag", "Bar", "Bar"},
		{"nonexistent field", "Baz", "Baz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetJSONFieldName(testStruct{}, tt.inputField)
			if got != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, got)
			}
		})
	}

	t.Run("pointer to struct", func(t *testing.T) {
		got := GetJSONFieldName(&testStruct{}, "Foo")
		if got != "foo" {
			t.Errorf("expected 'foo', got %q", got)
		}
	})
}

func TestMapValidationErrors(t *testing.T) {
	validate := validator.New()

	input := SampleInput{
		Name:  "",
		Email: "not-an-email",
	}

	err := validate.Struct(input)
	if err == nil {
		t.Fatal("expected validation error, got nil")
	}

	errors := MapValidationErrors(err, input)

	expectedErrors := []string{
		"name failed validation: required",
		"email failed validation: email",
	}

	if len(errors) != len(expectedErrors) {
		t.Fatalf("expected %d errors, got %d", len(expectedErrors), len(errors))
	}

	for i := range expectedErrors {
		if errors[i] != expectedErrors[i] {
			t.Errorf("expected error %q, got %q", expectedErrors[i], errors[i])
		}
	}
}

func TestMapValidationErrors_NonValidationError(t *testing.T) {
	err := errors.New("some random error")
	out := MapValidationErrors(err, struct{}{})

	if len(out) != 1 || out[0] != "some random error" {
		t.Errorf("expected 'some random error', got %v", out)
	}
}
