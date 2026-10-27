//
// Copyright 2026 IBM Corporation
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
	"crypto/x509"
	"encoding/pem"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestGetSelfSignedCertSettingsDefaults(t *testing.T) {
	settings, err := GetSelfSignedCertSettings()
	assert.NoError(t, err)
	assert.Equal(t, DefaultSelfSignedCertKeySize, settings.KeySize)
	assert.Equal(t, DefaultSelfSignedCertValidityDays, settings.ValidityDays)
}

func TestGetSelfSignedCertSettingsOverrides(t *testing.T) {
	t.Setenv(selfSignedCertKeySizeEnvVar, "4096")
	t.Setenv(selfSignedCertValidityDaysEnvVar, "30")

	settings, err := GetSelfSignedCertSettings()
	assert.NoError(t, err)
	assert.Equal(t, 4096, settings.KeySize)
	assert.Equal(t, 30, settings.ValidityDays)
}

func TestGetSelfSignedCertSettingsRejectsInsecureOrInvalidValues(t *testing.T) {
	t.Setenv(selfSignedCertKeySizeEnvVar, "1024")
	_, err := GetSelfSignedCertSettings()
	assert.Error(t, err, "key size below the minimum must be rejected")

	t.Setenv(selfSignedCertKeySizeEnvVar, "not-a-number")
	_, err = GetSelfSignedCertSettings()
	assert.Error(t, err, "non numeric key size must be rejected")

	t.Setenv(selfSignedCertKeySizeEnvVar, "")
	t.Setenv(selfSignedCertValidityDaysEnvVar, "0")
	_, err = GetSelfSignedCertSettings()
	assert.Error(t, err, "non positive validity must be rejected")
}

func TestGenerateSelfSignedCertSecretUsesSettings(t *testing.T) {
	t.Setenv(selfSignedCertKeySizeEnvVar, "2048")
	t.Setenv(selfSignedCertValidityDaysEnvVar, "10")

	secret, err := GenerateSelfSignedCertSecret(types.NamespacedName{Name: "name", Namespace: "ns"}, []string{"svc.ns.svc"})
	assert.NoError(t, err)

	keyBlock, _ := pem.Decode(secret.Data["tls.key"])
	key, err := x509.ParsePKCS1PrivateKey(keyBlock.Bytes)
	assert.NoError(t, err)
	assert.Equal(t, 2048, key.N.BitLen())

	certBlock, _ := pem.Decode(secret.Data["tls.crt"])
	cert, err := x509.ParseCertificate(certBlock.Bytes)
	assert.NoError(t, err)
	assert.WithinDuration(t, time.Now().AddDate(0, 0, 10), cert.NotAfter, time.Minute)
}
