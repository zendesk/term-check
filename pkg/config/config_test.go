package config

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

type envTestCase struct {
	name           string
	key            string
	backup         string
	env            map[string]string
	expectFatal    bool
	expectedOutput string
}

func TestEnv(t *testing.T) {
	cases := []envTestCase{
		{
			name:           "EnvSuccessfulLookup",
			key:            "ABC",
			backup:         "",
			env:            map[string]string{"ABC": "123"},
			expectFatal:    false,
			expectedOutput: "123",
		},
		{
			name:           "EnvBackupValue",
			key:            "ABC",
			backup:         "EFG",
			env:            make(map[string]string),
			expectFatal:    false,
			expectedOutput: "EFG",
		},
		{
			name:           "EnvPanicNotFound",
			key:            "ABC",
			backup:         "",
			env:            make(map[string]string),
			expectFatal:    true,
			expectedOutput: "",
		},
	}

	for _, tc := range cases {
		mockLookupEnv := func(s string) (string, bool) {
			res, ok := tc.env[s]
			return res, ok
		}

		mockFatalLog := func(format string, a ...interface{}) {
			panic(fmt.Sprintf(format, a))
		}

		c := New(
			WithLookupEnv(mockLookupEnv),
			WithFatalLog(mockFatalLog),
		)

		t.Run(tc.name, func(t *testing.T) {
			if tc.expectFatal {
				expectedMessage := "Environment variable is not set: [ABC]"
				assert.PanicsWithValue(t, expectedMessage, func() { c.Env(tc.key, tc.backup) }, "log.Fatal was not called")
			} else {
				assert.Equal(t, tc.expectedOutput, c.Env(tc.key, tc.backup))
			}
		})
	}
}
