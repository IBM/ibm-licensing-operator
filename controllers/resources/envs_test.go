//
// Copyright 2023 IBM Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
//

package resources

import (
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGetWatchNamespaceAsList(t *testing.T) {
	envVar := "WATCH_NAMESPACE"

	t.Log("Given the need to read and sanitize the WATCH_NAMESPACE environment variable")
	{
		testCases := []struct {
			name     string
			value    string
			expected []string
		}{
			{"single namespace", "ibm-licensing", []string{"ibm-licensing"}},
			{"multiple namespaces", "ns1,ns2,ns3", []string{"ns1", "ns2", "ns3"}},
			{"surrounding whitespace is trimmed", " ns1 , ns2 ", []string{"ns1", "ns2"}},
			{"empty entries are skipped", "ns1,,ns2,", []string{"ns1", "ns2"}},
			{"whitespace-only entries are skipped", "ns1, ,ns2", []string{"ns1", "ns2"}},
			{"duplicates are removed", "ns1,ns2,ns1", []string{"ns1", "ns2"}},
			{"duplicates after trimming are removed", "ns1, ns1 ,ns2", []string{"ns1", "ns2"}},
			{"input order is preserved", "ns3,ns1,ns2", []string{"ns3", "ns1", "ns2"}},
			{"empty value yields empty list", "", []string{}},
			{"only separators yields empty list", ",,,", []string{}},
		}

		for i, tc := range testCases {
			t.Logf("\tTest %d:\t%s", i, tc.name)
			{
				t.Setenv(envVar, tc.value)

				actual, err := GetWatchNamespaceAsList()
				if err != nil {
					t.Fatalf("\t%s\tShould not get an error : %v", FAIL, err)
				}
				if reflect.DeepEqual(actual, tc.expected) {
					t.Logf("\t%s\tShould get %v", SUCCESS, tc.expected)
				} else {
					t.Errorf("\t%s\tShould get %v : got %v", FAIL, tc.expected, actual)
				}
			}
		}

		t.Log("\tTest:\tWhen env var is not set")
		{
			os.Unsetenv(envVar)

			_, err := GetWatchNamespaceAsList()
			if err == nil {
				t.Fatalf("\t%s\tShould get an error when %s is unset", FAIL, envVar)
			}

			errMsg := envVar + " must be set"
			if strings.Contains(err.Error(), errMsg) {
				t.Logf("\t%s\tShould get an error with a proper message", SUCCESS)
			} else {
				t.Errorf("\t%s\tShould get an error with a proper message : %s", FAIL, "\""+errMsg+"\"")
			}
		}
	}
}

func TestGetReconcileInterval(t *testing.T) {
	envVar := "CRD_RECONCILE_INTERVAL"

	t.Logf("Given the need to read %s environment variable", envVar)
	{
		t.Log("\tTest 0:\tWhen env var is set")
		{
			expectedInterval := 20
			initErr := os.Setenv(envVar, strconv.Itoa(expectedInterval))
			if initErr != nil {
				t.Fatalf("Could not set %s env var : %v", envVar, initErr)
			}

			actualInterval, err := GetCrdReconcileInterval()
			if err != nil {
				t.Fatalf("\t%s\tShould get time interval in seconds : %v", FAIL, err)
			}
			if actualInterval == time.Duration(expectedInterval)*time.Second {
				t.Logf("\t%s\tShould get time interval %s", SUCCESS, actualInterval)
			} else {
				t.Errorf("\t%s\tShould get time interval %ds : %s", SUCCESS, expectedInterval, actualInterval)
			}

			os.Unsetenv(envVar)
		}

		t.Log("\tTest 1:\tWhen env var is not set")
		{
			expectedInterval := 300 * time.Second
			actualInterval, err := GetCrdReconcileInterval()
			if err != nil {
				t.Fatalf("\t%s\tShould get time interval in seconds : %v", FAIL, err)
			}
			if actualInterval == expectedInterval {
				t.Logf("\t%s\tShould get default time interval %s", SUCCESS, actualInterval)
			} else {
				t.Errorf("\t%s\tShould get default time interval %s : %s", SUCCESS, expectedInterval, actualInterval)
			}
		}

		t.Log("\tTest 2:\tWhen env var has forbidden format")
		{
			os.Setenv(envVar, "1.5")
			_, err := GetCrdReconcileInterval()
			if err == nil {
				t.Fatalf("\t%s\tShould get a format error", FAIL)
			}

			errMsg := envVar + " must be a natural number"
			if strings.Contains(err.Error(), errMsg) {
				t.Logf("\t%s\tShould get an error with a proper message", SUCCESS)
			} else {
				t.Errorf("\t%s\tShould get an error with a proper message : %s", FAIL, "\""+errMsg+"\"")
			}
			os.Unsetenv(envVar)
		}
	}

}
