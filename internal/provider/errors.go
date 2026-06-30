// Copyright (c) Joe Lee
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"errors"

	"github.com/joelee2012/go-nacos"
)

// IsNotFoundError checks whether an error from the go-nacos client indicates
// that the requested resource does not exist.
func IsNotFoundError(err error) bool {
	var nacosErr nacos.NacosErr
	if errors.As(err, &nacosErr) {
		return nacosErr.IsNotFound()
	}
	return errors.Is(err, nacos.ErrNotFound)
}
