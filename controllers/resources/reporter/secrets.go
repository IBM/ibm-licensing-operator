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

package reporter

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	operatorv1alpha1 "github.com/IBM/ibm-licensing-operator/api/v1alpha1"
	"github.com/IBM/ibm-licensing-operator/controllers/resources"
	"github.com/IBM/ibm-licensing-operator/version"
)

const APIReciverSecretTokenKeyName = "token"

func GetZenNginxConf(instance *operatorv1alpha1.IBMLicenseServiceReporter) string {
	return `location /license-service-reporter/ {
  access_by_lua_file /nginx_data/checkjwt.lua;
  set_by_lua $nsdomain 'return os.getenv("NS_DOMAIN")';
  proxy_http_version 1.1;
  proxy_set_header Upgrade $http_upgrade;
  proxy_set_header Connection "upgrade";
  proxy_set_header Host $host;
  proxy_set_header zen-namespace-domain $nsdomain;
  proxy_pass http://ibm-license-service-reporter.` + instance.GetNamespace() + `.svc.cluster.local:3001/license-service-reporter/;
  proxy_read_timeout 10m;
}`
}

const ZenExtensions = `[
  {
    "extension_point_id": "left_menu_item",
    "extension_name": "nav-license-service-reporter",
    "display_name": "Licensing",
    "order_hint": 700,
    "match_permissions": "administrator",
    "meta": {},
    "details": {
		"parent_folder": "dap-header-administer",
		"href": "/license-service-reporter?isZen=true"
    }
  }
]`

func GetAPISecretToken(instance *operatorv1alpha1.IBMLicenseServiceReporter) (*corev1.Secret, error) {
	return resources.GetSecretToken(instance.Spec.APISecretToken, instance.GetNamespace(), APIReciverSecretTokenKeyName, LabelsForMeta(instance))
}

func GetDatabaseSecret(instance *operatorv1alpha1.IBMLicenseServiceReporter) (*corev1.Secret, error) {
	metaLabels := LabelsForMeta(instance)
	randString, err := resources.RandString(8)
	if err != nil {
		return nil, err
	}
	expectedSecret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DatabaseConfigSecretName,
			Namespace: instance.GetNamespace(),
			Labels:    metaLabels,
		},
		Type: corev1.SecretTypeOpaque,
		StringData: map[string]string{
			PostgresPasswordKey:     randString,
			PostgresUserKey:         DatabaseUser,
			PostgresDatabaseNameKey: DatabaseName,
			PostgresPgDataKey:       PgData,
		},
	}
	return expectedSecret, nil
}

func GetZenConfigMap(instance *operatorv1alpha1.IBMLicenseServiceReporter) *corev1.ConfigMap {
	labels := map[string]string{
		"icpdata_addon":         "true",
		"icpdata_addon_version": version.Version,
	}
	expectedCM := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      ZenConfigMapName,
			Namespace: instance.GetNamespace(),
			Labels:    labels,
		},
		Data: map[string]string{
			"nginx.conf": GetZenNginxConf(instance),
			"extensions": ZenExtensions,
		},
	}
	return expectedCM
}
