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
	"strconv"
	"strings"
	"testing"
	"time"
)

const (
	succeed = "\u2713"
	failed  = "\u2717"
)

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
				t.Fatalf("\t%s\tShould get time interval in seconds : %v", failed, err)
			}
			if actualInterval == time.Duration(expectedInterval)*time.Second {
				t.Logf("\t%s\tShould get time interval %s", succeed, actualInterval)
			} else {
				t.Errorf("\t%s\tShould get time interval %ds : %s", succeed, expectedInterval, actualInterval)
			}

			os.Unsetenv(envVar)
		}

		t.Log("\tTest 1:\tWhen env var is not set")
		{
			expectedInterval := 3600 * time.Second
			actualInterval, err := GetCrdReconcileInterval()
			if err != nil {
				t.Fatalf("\t%s\tShould get time interval in seconds : %v", failed, err)
			}
			if actualInterval == 3600*time.Second {
				t.Logf("\t%s\tShould get default time interval %s", succeed, actualInterval)
			} else {
				t.Errorf("\t%s\tShould get default time interval %s : %s", succeed, expectedInterval, actualInterval)
			}
		}

		t.Log("\tTest 2:\tWhen env var has forbidden format")
		{
			os.Setenv(envVar, "1.5")
			_, err := GetCrdReconcileInterval()
			if err == nil {
				t.Fatalf("\t%s\tShould get a format error", failed)
			}

			errMsg := envVar + " must be natural number"
			if strings.Contains(err.Error(), errMsg) {
				t.Logf("\t%s\tShould get an error with a proper message", succeed)
			} else {
				t.Errorf("\t%s\tShould get an error with a proper message : %s", failed, "\""+errMsg+"\"")
			}
			os.Unsetenv(envVar)
		}
	}

}