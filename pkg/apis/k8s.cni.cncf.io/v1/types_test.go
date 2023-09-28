package v1

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

type testCase struct {
	description    string
	input          string
	expectedOutput NetworkSelectionElement
	expectedError  error
}

func TestNetworkSelectionElementUnmarshaller(t *testing.T) {

	testCases := []testCase{
		{
			description:   "ip request + IPAMClaims",
			input:         "{\"name\":\"yo!\",\"ips\":[\"asd\"],\"ipam-claim-reference\":\"woop\"}",
			expectedError: TooManyIPSources,
		},
		{
			description: "successfully deserialize a simple struct",
			input:       "{\"name\":\"yo!\",\"ips\":[\"an IP\"]}",
			expectedOutput: NetworkSelectionElement{
				Name:      "yo!",
				IPRequest: []string{"an IP"},
			},
			expectedError: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.description, func(t *testing.T) {
			if err := run(tc); err != nil {
				t.Errorf("failed test %q: %v", tc.description, err)
			}
		})
	}
}

func run(tc testCase) error {
	inputBytes := []byte(tc.input)

	var nse NetworkSelectionElement
	err := json.Unmarshal(inputBytes, &nse)
	if tc.expectedError != nil {
		if !errors.Is(err, tc.expectedError) {
			return fmt.Errorf("unexpected error: %v. Expected error: %v", err, tc.expectedError)
		}
	}
	if !reflect.DeepEqual(nse, tc.expectedOutput) {
		return fmt.Errorf("parsed object is wrong: %v. Expected object: %v", nse, tc.expectedOutput)
	}
	return nil
}
