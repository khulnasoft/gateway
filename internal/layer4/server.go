// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package layer4

import (
	kengine "github.com/khulnasoft/gateway/internal/kenginev2"
)

// Server represents a Kengine layer4 server.
type Server struct {
	// The network address to bind to. Any Kengine network address
	// is an acceptable value:
	// https://khulnasoft.com/docs/conventions#network-addresses
	Listen []string `json:"listen,omitempty"`

	// Routes express composable logic for handling byte streams.
	Routes RouteList `json:"routes,omitempty"`

	// Maximum time connections have to complete the matching phase (the first terminal handler is matched). Default: 3s.
	MatchingTimeout kengine.Duration `json:"matching_timeout,omitempty"`
}
