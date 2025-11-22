// Copyright The OpenTelemetry Authors
// SPDX-License-Identifier: Apache-2.0

//go:generate mdatagen metadata.yaml

// Package baggageprocessor implements a processor that handles W3C baggage propagation format.
// It provides functionality to extract, inject, modify, and filter baggage entries according to
// the W3C Baggage specification (https://www.w3.org/TR/baggage/).
package baggageprocessor // import "github.com/open-telemetry/opentelemetry-collector-contrib/processor/baggageprocessor"
