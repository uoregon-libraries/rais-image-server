package main

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseSchemeMap(t *testing.T) {
	var tests = map[string]struct {
		input     string
		hasError  bool
		extraMaps map[string]string
	}{
		"simple": {
			input: "foo=file:///var/local bar=s3://bucket/path/", hasError: false,
			extraMaps: map[string]string{"foo": "file:///var/local/", "bar": "s3://bucket/path/"},
		},
		"double map": {
			input: "foo=file:///var/local bar=s3://bucket/path/ foo=file:///etc", hasError: true,
			extraMaps: map[string]string{"foo": "file:///var/local/", "bar": "s3://bucket/path/"},
		},
		"remap internals":       {input: "file=file:///", hasError: true, extraMaps: nil},
		"invalid prefix scheme": {input: "file2=/var/local", hasError: true, extraMaps: nil},
		"file with a host":      {input: "file2=file://host/path", hasError: true, extraMaps: nil},
		"s3 with no host":       {input: "coll1=s3:///path", hasError: true, extraMaps: nil},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			var actual = NewImageHandler("/tilepath", "/iiif")
			var err = parseSchemeMap(actual, tc.input)
			if err == nil && tc.hasError {
				t.Errorf("expected error, got nil")
			}
			if err != nil && !tc.hasError {
				t.Errorf("expected no error, got %s", err)
			}

			var expected = NewImageHandler("/tilepath", "url")
			for scheme, prefix := range tc.extraMaps {
				expected.schemeMap[scheme] = prefix
			}
			var diff = cmp.Diff(expected.schemeMap, actual.schemeMap)
			if diff != "" {
				t.Errorf(diff)
			}
		})
	}
}
