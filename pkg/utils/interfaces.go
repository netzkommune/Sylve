// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package utils

import (
	"reflect"

	"github.com/go-playground/validator/v10"
)

func MapValidationErrors(err error, structType interface{}) []string {
	var validationErrors []string

	if ve, ok := err.(validator.ValidationErrors); ok {
		for _, fieldErr := range ve {
			jsonFieldName := GetJSONFieldName(structType, fieldErr.StructField())
			validationErrors = append(validationErrors, jsonFieldName+" failed validation: "+fieldErr.Tag())
		}
	} else {
		validationErrors = append(validationErrors, err.Error())
	}

	return validationErrors
}

func GetJSONFieldName(structType interface{}, fieldName string) string {
	t := reflect.TypeOf(structType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	field, found := t.FieldByName(fieldName)
	if !found {
		return fieldName
	}

	jsonTag := field.Tag.Get("json")
	if jsonTag == "" {
		return fieldName
	}

	return jsonTag
}
