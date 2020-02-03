//
// Copyright 2020 IBM Corporation
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

package ibmlicensing

import (
	"context"

	operatorv1alpha1 "github.com/ibm/ibm-licensing-operator/pkg/apis/operator/v1alpha1"
	res "github.com/ibm/ibm-licensing-operator/pkg/resources"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

// cannot set to const due to corev1 types
var defaultSecretMode int32 = 420
var seconds60 int64 = 60

var log = logf.Log.WithName("controller_ibmlicensing")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new IBMLicensing Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileIBMLicensing{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("ibmlicensing-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource IBMLicensing
	err = c.Watch(&source.Kind{Type: &operatorv1alpha1.IBMLicensing{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource "Deployment" and requeue the owner IBMLicensing
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.IBMLicensing{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource "Service" and requeue the owner IBMLicensing
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.IBMLicensing{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &corev1.Secret{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &operatorv1alpha1.IBMLicensing{},
	})
	if err != nil {
		return err
	}
	return nil
}

// blank assignment to verify that ReconcileIBMLicensing implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileIBMLicensing{}

// ReconcileIBMLicensing reconciles a IBMLicensing object
type ReconcileIBMLicensing struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a IBMLicensing object and makes changes based on the state read
// and what is in the IBMLicensing.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Deployment and Service for IBM Licensing
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileIBMLicensing) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling IBMLicensing")

	// Fetch the IBMLicensing instance
	instance := &operatorv1alpha1.IBMLicensing{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			reqLogger.Info("IBMLicensing resource not found. Ignoring since object must be deleted")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get IBMLicensing")
		return reconcile.Result{}, err
	}

	opVersion := instance.Spec.OperatorVersion
	reqLogger.Info("got IBMLicensing instance, version=" + opVersion + ", checking Service")

	// Check if the service already exists
	currentService := &corev1.Service{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.GetResourceName(instance), Namespace: instance.GetNamespace()}, currentService)
	// In case error is cause by non existing service we will create one:
	if err != nil && errors.IsNotFound(err) {
		// Define a new service
		newService := r.newServiceForLicensingCR(instance)
		reqLogger.Info("Creating a new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
		err = r.client.Create(context.TODO(), newService)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Service", "Service.Namespace", newService.Namespace, "Service.Name", newService.Name)
			return reconcile.Result{}, err
		}
		// Service created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Service")
		return reconcile.Result{}, err
	}
	reqLogger.Info("got Service, checking Deployment")

	// Check if the deployment already exists, if not create a new one
	currentDeployment := &appsv1.Deployment{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: res.GetResourceName(instance), Namespace: instance.GetNamespace()}, currentDeployment)
	if err != nil && errors.IsNotFound(err) {
		// Define a new deployment
		newDeployment := r.newDeploymentForLicensingCR(instance)
		reqLogger.Info("Creating a new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
		err = r.client.Create(context.TODO(), newDeployment)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Deployment", "Deployment.Namespace", newDeployment.Namespace, "Deployment.Name", newDeployment.Name)
			return reconcile.Result{}, err
		}
		// Deployment created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment")
		return reconcile.Result{}, err
	}
	reqLogger.Info("got Deployment, checking APISecretToken")

	currentAPISecret := &corev1.Secret{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Spec.APISecretToken, Namespace: instance.GetNamespace()}, currentAPISecret)
	if err != nil && errors.IsNotFound(err) {
		// APISecretToken does not exist
		reqLogger.Info("APISecretToken does not exist, creating secret: " + instance.Spec.APISecretToken)
		newAPISecret := r.newAPISecretToken(instance)
		err = r.client.Create(context.TODO(), newAPISecret)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Secret", "Secret.Namespace", newAPISecret.Namespace, "Secret.Name", newAPISecret.Name)
			return reconcile.Result{}, err
		}
		// Secret created successfully - return and requeue
		return reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Deployment APISecretToken")
		return reconcile.Result{}, err
	}

	reqLogger.Info("all done")
	return reconcile.Result{}, nil
}

func (r *ReconcileIBMLicensing) newAPISecretToken(instance *operatorv1alpha1.IBMLicensing) *corev1.Secret {
	reqLogger := log.WithValues("APISecretToken", "Entry", "instance.GetName()", instance.GetName())
	metaLabels := res.LabelsForLicensingMeta(instance)

	reqLogger.Info("New APISecretToken Entry")
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Spec.APISecretToken,
			Namespace: instance.GetNamespace(),
			Labels:    metaLabels,
		},
		Type:       corev1.SecretTypeOpaque,
		StringData: map[string]string{"token": res.RandString(24)},
	}
	// Set IBMLicensing instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, secret, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Secret APISecretToken")
		return nil
	}
	return secret
}

func (r *ReconcileIBMLicensing) newServiceForLicensingCR(instance *operatorv1alpha1.IBMLicensing) *corev1.Service {
	reqLogger := log.WithValues("serviceForLicensing", "Entry", "instance.GetName()", instance.GetName())
	metaLabels := res.LabelsForLicensingMeta(instance)

	reqLogger.Info("New Service Entry")
	service := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.GetResourceName(instance),
			Namespace: instance.GetNamespace(),
			Labels:    metaLabels,
			// Annotations: map[string]string{"prometheus.io/scrape": "false", "prometheus.io/scheme": "http"},
		},
		Spec: res.GetServiceSpec(instance),
	}

	// Set IBMLicensing instance as the owner and controller of the Service
	err := controllerutil.SetControllerReference(instance, service, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Service")
		return nil
	}
	return service
}

// deploymentForDataMgr returns a DataManager Deployment object
func (r *ReconcileIBMLicensing) newDeploymentForLicensingCR(instance *operatorv1alpha1.IBMLicensing) *appsv1.Deployment {
	reqLogger := log.WithValues("newDeploymentForLicensingCR", "Entry", "instance.GetName()", instance.GetName())

	metaLabels := res.LabelsForLicensingMeta(instance)
	selectorLabels := res.LabelsForLicensingSelector(instance)
	podLabels := res.LabelsForLicensingPod(instance)

	// TODO: maybe add to cr later
	replicas := int32(1)
	reqLogger.Info("image=" + instance.Spec.GetFullImage())

	volumes := []corev1.Volume{}

	apiSecretTokenVolume := corev1.Volume{
		Name: res.APISecretTokenVolumeName,
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: instance.Spec.APISecretToken,
			},
		},
	}

	volumes = append(volumes, apiSecretTokenVolume)

	if instance.Spec.IsMetering() {
		meteringAPICertVolume := corev1.Volume{
			Name: res.MeteringAPICertsVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "icp-metering-api-secret",
					// TODO: validate if good mode, not 0644?
					DefaultMode: &defaultSecretMode,
					Optional:    &res.TrueVar,
				},
			},
		}

		volumes = append(volumes, meteringAPICertVolume)
	}

	if instance.Spec.HTTPSEnable {
		licensingHTTPSCertsVolume := corev1.Volume{
			Name: res.LicensingHTTPSCertsVolumeName,
			VolumeSource: corev1.VolumeSource{
				Secret: &corev1.SecretVolumeSource{
					SecretName: "ibm-licensing-certs",
					// TODO: validate if good mode, not 0644?
					DefaultMode: &defaultSecretMode,
					Optional:    &res.TrueVar,
				},
			},
		}

		volumes = append(volumes, licensingHTTPSCertsVolume)
	}

	//TODO: add init containers later
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      res.GetResourceName(instance),
			Namespace: instance.GetNamespace(),
			Labels:    metaLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: selectorLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: podLabels,
				},
				Spec: corev1.PodSpec{
					// TODO: decide if needed:
					// NodeSelector:                  nodeSelector,
					// PriorityClassName:             "system-cluster-critical",
					TerminationGracePeriodSeconds: &seconds60,
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "beta.kubernetes.io/arch",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"amd64"},
											},
										},
									},
								},
							},
						},
					},
					// TODO: decide if neeeded:
					// Tolerations: []corev1.Toleration{
					// 	{
					// 		Key:      "dedicated",
					// 		Operator: corev1.TolerationOpExists,
					// 		Effect:   corev1.TaintEffectNoSchedule,
					// 	},
					// 	{
					// 		Key:      "CriticalAddonsOnly",
					// 		Operator: corev1.TolerationOpExists,
					// 	},
					// },
					Volumes: volumes,
					Containers: []corev1.Container{
						res.GetLicensingContainer(instance.GetNamespace(), instance.Spec),
					},
				},
			},
		},
	}
	// Set Metering instance as the owner and controller of the Deployment
	err := controllerutil.SetControllerReference(instance, deployment, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set owner for Deployment")
		return nil
	}
	return deployment
}
