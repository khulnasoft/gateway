// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package l4tls

import (
	"github.com/khulnasoft/gateway/internal/kenginev2/kenginetls"
)

type HandlerName string

func (HandlerName) MarshalJSON() ([]byte, error) {
	return []byte(`"tls"`), nil
}

// Handler is a connection handler that terminates TLS.
type Handler struct {
	// Handler is the name of this handler for the JSON config.
	// DO NOT USE this. This is a special value to represent this handler.
	// It will be overwritten when we are marshalled.
	Handler HandlerName `json:"handler"`

	ConnectionPolicies kenginetls.ConnectionPolicies `json:"connection_policies,omitempty"`
}

func (Handler) IAmAHandler() {}
