// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package kenginehttp

import (
	"github.com/khulnasoft/gateway/internal/kenginev2/kenginehttp/proxyprotocol"
)

// ListenerWrappers .
// TODO: document
type ListenerWrappers []ListenerWrapper

// ListenerWrapper .
// TODO: document
// ref; https://khulnasoft.com/docs/json/apps/http/servers/listener_wrappers/
type ListenerWrapper struct {
	// HTTPRedirect .
	// TODO: document
	// ref; https://khulnasoft.com/docs/json/apps/http/servers/listener_wrappers/http_redirect/
	HTTPRedirect *HTTPRedirectListenerWrapper `json:"http_redirect,omitempty"`

	// TLS .
	// TODO: document
	// ref; https://khulnasoft.com/docs/json/apps/http/servers/listener_wrappers/tls/
	TLS *TLSListenerWrapper `json:"tls,omitempty"`

	// ProxyProtocol .
	// TODO: document
	// ref; https://khulnasoft.com/docs/json/apps/http/servers/listener_wrappers/proxy_protocol/
	ProxyProtocol *proxyprotocol.ListenerWrapper `json:"proxy_protocol,omitempty"`
}
