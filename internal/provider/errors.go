// Copyright (c) Joe Lee
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"

	"github.com/joelee2012/go-nacos"
)

// IsNotFoundError checks whether an error from the go-nacos client indicates
// that the requested resource does not exist. Since go-nacos v0.4.0 the client
// returns the sentinel nacos.ErrNotFound for both 404 responses and lookups
// that come up empty, so a single errors.Is check is sufficient.
func IsNotFoundError(err error) bool {
	return errors.Is(err, nacos.ErrNotFound)
}
