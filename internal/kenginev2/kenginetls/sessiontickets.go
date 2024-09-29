// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package kenginetls

import (
	kengine "github.com/khulnasoft/gateway/internal/kenginev2"
)

// SessionTicketService configures and manages TLS session tickets.
type SessionTicketService struct {
	// KeySource is the method by which Kengine produces or obtains
	// TLS session ticket keys (STEKs). By default, Kengine generates
	// them internally using a secure pseudorandom source.
	// TODO: type this
	KeySource any `json:"key_source,omitempty"`

	// How often Kengine rotates STEKs. Default: 12h.
	RotationInterval kengine.Duration `json:"rotation_interval,omitempty"`

	// The maximum number of keys to keep in rotation. Default: 4.
	MaxKeys int `json:"max_keys,omitempty"`

	// Disables STEK rotation.
	DisableRotation bool `json:"disable_rotation,omitempty"`

	// Disables TLS session resumption by tickets.
	Disabled bool `json:"disabled,omitempty"`
}
