// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package internal

import "github.com/go-playground/validator/v10"

var Validate = validator.New()

func ValidationErrorResponse(err error) validator.FieldError {
	if errs, ok := err.(validator.ValidationErrors); ok && len(errs) > 0 {
		return errs[0]
	}
	return nil
}
