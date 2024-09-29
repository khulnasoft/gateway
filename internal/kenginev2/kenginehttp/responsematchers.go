// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package kenginehttp

import (
	"net/http"
)

// ResponseMatcher is a type which can determine if an
// HTTP response matches some criteria.
type ResponseMatcher struct {
	// If set, one of these status codes would be required.
	// A one-digit status can be used to represent all codes
	// in that class (e.g. 3 for all 3xx codes).
	StatusCode []int `json:"status_code,omitempty"`

	// If set, each header specified must be one of the
	// specified values, with the same logic used by the
	// [request header matcher](/docs/json/apps/http/servers/routes/match/header/).
	Headers http.Header `json:"headers,omitempty"`
}
