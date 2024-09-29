// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: Copyright (c) 2024 KhulnaSoft Ltd

package kengine

import (
	gatewayv1 "sigs.k8s.io/gateway-api/apis/v1"

	"github.com/khulnasoft/gateway/internal/kenginev2/kenginehttp"
)

// getPathMatcher .
// ref; https://khulnasoft.com/docs/json/apps/http/servers/routes/match/path/
func (i *Input) getPathMatcher(matcher *kenginehttp.Match, path *gatewayv1.HTTPPathMatch) error {
	if path == nil || path.Value == nil {
		return nil
	}
	value := *path.Value
	if value == "" {
		return nil
	}
	var matchType gatewayv1.PathMatchType
	if path.Type == nil {
		matchType = gatewayv1.PathMatchPathPrefix
	} else {
		matchType = *path.Type
	}

	// If the path is `/` and the match type is a PathPrefix,
	// ignore it. This is just a verbose way of saying "match
	// all paths".
	if value == "/" && matchType == gatewayv1.PathMatchPathPrefix {
		return nil
	}

	switch matchType {
	case gatewayv1.PathMatchExact:
		matcher.Path = kenginehttp.MatchPath{value}
	case gatewayv1.PathMatchPathPrefix:
		matcher.Path = kenginehttp.MatchPath{value + "*"}
	case gatewayv1.PathMatchRegularExpression:
		matcher.PathRE = &kenginehttp.MatchPathRE{
			MatchRegexp: kenginehttp.MatchRegexp{
				Pattern: value,
			},
		}
	}
	return nil
}

// getHeaderMatcher .
// ref; https://khulnasoft.com/docs/json/apps/http/servers/routes/match/header/
func (i *Input) getHeaderMatcher(matcher *kenginehttp.Match, v []gatewayv1.HTTPHeaderMatch) error {
	if v == nil {
		return nil
	}

	// TODO: implement
	return nil
}

// getQueryMatcher .
// ref; https://khulnasoft.com/docs/json/apps/http/servers/routes/match/query/
func (i *Input) getQueryMatcher(matcher *kenginehttp.Match, v []gatewayv1.HTTPQueryParamMatch) error {
	if v == nil {
		return nil
	}

	// TODO: implement
	return nil
}

// getMethodMatcher .
// ref; https://khulnasoft.com/docs/json/apps/http/servers/routes/match/method/
func (i *Input) getMethodMatcher(matcher *kenginehttp.Match, m *gatewayv1.HTTPMethod) error {
	if m == nil {
		return nil
	}
	matcher.Method = kenginehttp.MatchMethod{string(*m)}
	return nil
}
