package giantnetes

import (
	"strconv"
	"testing"

	"github.com/coreos/go-semver/semver"
	"github.com/google/go-cmp/cmp"
)

func Test_stringVersionCmp(t *testing.T) {
	testCases := []struct {
		name         string
		version      string
		def          *semver.Version
		reference    *semver.Version
		supported    bool
		errorMatcher func(error) bool
	}{
		{
			name:         "case 0: patch version is greater than reference",
			version:      "1.0.1",
			def:          semver.New("0.0.0"),
			reference:    semver.New("1.0.0"),
			supported:    true,
			errorMatcher: nil,
		},
		{
			name:         "case 1: patch version is lower than reference",
			version:      "1.0.0",
			def:          semver.New("0.0.0"),
			reference:    semver.New("1.0.1"),
			supported:    false,
			errorMatcher: nil,
		},
		{
			name:         "case 2: minor version is greater than reference",
			version:      "1.1.0",
			def:          semver.New("0.0.0"),
			reference:    semver.New("1.0.0"),
			supported:    true,
			errorMatcher: nil,
		},
		{
			name:         "case 3: minor version is lower than reference",
			version:      "1.0.0",
			def:          semver.New("0.0.0"),
			reference:    semver.New("1.1.0"),
			supported:    false,
			errorMatcher: nil,
		},
		{
			name:         "case 4: major version is greater than reference",
			version:      "2.0.0",
			def:          semver.New("0.0.0"),
			reference:    semver.New("1.1.1"),
			supported:    true,
			errorMatcher: nil,
		},
		{
			name:         "case 5: major version is lower than reference",
			version:      "1.1.1",
			def:          semver.New("0.0.0"),
			reference:    semver.New("2.0.0"),
			supported:    false,
			errorMatcher: nil,
		},
		{
			name:         "case 6: empty version, default version not supported",
			version:      "",
			def:          semver.New("0.0.0"),
			reference:    semver.New("0.0.1"),
			supported:    false,
			errorMatcher: nil,
		},
		{
			name:         "case 7: empty version, default version supported",
			version:      "",
			def:          semver.New("0.0.1"),
			reference:    semver.New("0.0.1"),
			supported:    true,
			errorMatcher: nil,
		},
		{
			name:         "case 7: Invalid version",
			version:      "i_am_an_invalid_version",
			def:          semver.New("0.0.1"),
			reference:    semver.New("0.0.1"),
			supported:    false,
			errorMatcher: IsInvalidVersionError,
		},
	}

	for i, tc := range testCases {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			t.Log(tc.name)

			supported, err := stringVersionCmp(tc.version, tc.def, tc.reference)

			switch {
			case err == nil && tc.errorMatcher == nil:
				// Correct; carry on.
			case err != nil && tc.errorMatcher == nil:
				t.Fatalf("error == %#v, want nil", err)
			case err == nil && tc.errorMatcher != nil:
				t.Fatalf("error == nil, want non-nil")
			case !tc.errorMatcher(err):
				t.Fatalf("error == %#v, want matching", err)
			}

			if !cmp.Equal(supported, tc.supported) {
				t.Fatalf("\n\n%s\n", cmp.Diff(tc.supported, supported))
			}
		})
	}
}

// This is a third party error, not sure how to properly check it without risking to break the tests
// in case the third party changes the error message. In this scenario, just checking the error is present
// would be enough.
func IsInvalidVersionError(err error) bool {
	return err != nil
}
