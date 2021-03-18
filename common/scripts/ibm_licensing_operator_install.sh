#!/bin/bash
#
# Copyright 2021 IBM Corporation
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
# http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.
#

##### Constants

# Namespace where workload will be created (operator and operand)
INSTALL_NAMESPACE=${INSTALL_NAMESPACE:-ibm-common-services}

##### Functions

usage()
{
   # Display usage
  echo "description: A script to install IBM License Service via Operator."
  echo ""
  echo "note: Use this script only for cluster running on x86 architecture."
  echo ""
  echo "usage: $0 [--verbose | -v] [--help | -h] [(--olm_version | -o) <version_number>] [--skip_olm_installation | -s] [(--olm_global_catalog_namespace | -c) <OLM global catalog namespace> ] [(--channel | -l) <subscription channel>] [--no-secret-output | -n]"
  echo "options:"
  echo "[--verbose | -v] - verbose logs from installation"
  echo "[--channel | -l] - do not change unless instructed to. What channel should License Service Operator subscription choose,"
  echo "by default channel=v3"
  echo "[--no-secret-output | -n] - use this option to not show secret at the end of the script"
  echo "[--olm_version | -o] <version_number> - what version of OLM should be installed if it doesn't exist,"
  echo "by default olm_version=0.13.0"
  echo "[--skip_olm_installation | -s] - skips installation of OLM, but olm global catalog namespace still needs to be found."
  echo "[--olm_global_catalog_namespace | -c] <OLM global catalog namespace> - script will not try to find olm global catalog namespace when set."
  echo "You can read more about OLM global catalog namespace here: https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md"
  echo "[--help | -h] - shows usage"
  echo "prerequisite commands: kubectl, git, curl"
}

if [ "$(uname)" == "Darwin" ]; then
  inline_sed(){
    sed -i "" "$@"
  }
else
  inline_sed(){
    sed -i "$@"
  }
fi

verify_command_line_processing(){
  # Test code to verify command line processing
  verbose_output_command echo "olm version is ${olm_version}"
}

verify_kubectl(){
  if ! verbose_output_command kubectl version; then
    echo "Error: kubectl command does not seems to work"
    echo "try to install it and setup config for your cluster where you want to install IBM License Service"
    exit 2
  fi
}

create_namespace(){
  if ! verbose_output_command kubectl get namespace "${INSTALL_NAMESPACE}"; then
    echo "Creating namespace ${INSTALL_NAMESPACE}"
    if ! kubectl create namespace "${INSTALL_NAMESPACE}"; then
      echo "Error: kubectl command cannot create needed namespace"
      echo "make sure you are connected to your cluster where you want to install IBM License Service and have admin permissions"
      exit 3
    fi
  else
    echo "Needed namespace: \"${INSTALL_NAMESPACE}\", already exists"
  fi
}

install_olm(){
  if [ "${skip_olm_installation}" != "1" ]; then
    echo "Check if OLM is installed"
    verbose_output_command echo "Checking if CSV CRD exists"
    if ! verbose_output_command kubectl get crd clusterserviceversions.operators.coreos.com -o name; then
      echo "CSV CRD does not exists, installing OLM with version ${olm_version}"
      if ! curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/"${olm_version}"/install.sh | bash -s "${olm_version}"; then
        echo "Error: Failed to install OLM"
        echo "You can try to install OLM from here https://github.com/operator-framework/operator-lifecycle-manager/releases and continue installation while skipping OLM part"
        exit 5
      fi
    else
      verbose_output_command echo "OLM's needed CRD: CSV exists"
    fi
  else
    verbose_output_command echo "Skipping OLM installation"
  fi
  if [ "${olm_global_catalog_namespace}" == "" ]; then
    verbose_output_command echo "Trying to get namespace where OLM's packageserver is installed"
    if ! olm_namespace=$(kubectl get csv --all-namespaces -l olm.version -o jsonpath="{.items[?(@.metadata.name=='packageserver')].metadata.namespace}") || [ "${olm_namespace}" == "" ]; then
      if [ "${skip_olm_installation}" != "1" ]; then
        echo "OLM CRD was found but packageserver csv was not found, will try to install OLM with version ${olm_version}"
        if ! curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/"${olm_version}"/install.sh | bash -s "${olm_version}"; then
          echo "Error: Failed to install OLM"
          echo "You can try to install OLM from here https://github.com/operator-framework/operator-lifecycle-manager/releases and continue installation while skipping OLM part"
          exit 24
        fi
        verbose_output_command echo "Installed OLM ${olm_version}, will try to get olm_namespace again"
        if ! olm_namespace=$(kubectl get csv --all-namespaces -l olm.version -o jsonpath="{.items[?(@.metadata.name=='packageserver')].metadata.namespace}") || [ "${olm_namespace}" == "" ]; then
          echo "Error: Failed to get namespace where OLM's packageserver is installed, which is needed for finding OLM's global catalog namespace, make sure you have OLM installed"
          echo "You can try to install OLM from here https://github.com/operator-framework/operator-lifecycle-manager/releases"
          echo "If you can find OLM's global catalog namespace yourself try setting parameter --olm_global_catalog_namespace parameter of this script"
          echo "On OpenShift Container Platform this probably is 'openshift-marketplace', but for older versions and for custom OLM installation it might be 'olm', but you might verify it by looking for OLM's packageserver deployment configuration"
          exit 25
        fi
      else
        echo "Error: Failed to get namespace where OLM's packageserver is installed, which is needed for finding OLM's global catalog namespace, make sure you have OLM installed"
        echo "You can try to install OLM from here https://github.com/operator-framework/operator-lifecycle-manager/releases"
        echo "If you can find OLM's global catalog namespace yourself try setting parameter --olm_global_catalog_namespace parameter of this script"
        echo "On OpenShift Container Platform this probably is 'openshift-marketplace', but for older versions and for custom OLM installation it might be 'olm', but you might verify it by looking for OLM's packageserver deployment configuration"
        exit 6
      fi
    else
      verbose_output_command echo "Namespace where OLM's packageserver is installed is: ${olm_namespace}"
    fi
    verbose_output_command echo "Trying to get OLM's global catalog namespace so that catalog needed by IBM Licensing can be accessed in any watched namespace."
    if ! olm_global_catalog_namespace=$(kubectl get deployment --namespace="${olm_namespace}" packageserver -o yaml | grep -A 1 -i global-namespace | tail -1 | cut -d "-" -f 2- | sed -e 's/^[ \t]*//') || [ "${olm_global_catalog_namespace}" == "" ]; then
      echo "Error: Failed to find OLM's global catalog namespace where catalog for IBM Licensing needs to be installed"
      echo "If you can find it yourself try setting parameter --olm_global_catalog_namespace parameter of this script"
      echo "On OpenShift Container Platform this probably is 'openshift-marketplace', but for older versions and for custom OLM installation it might be 'olm', but you might verify it by looking for OLM's packageserver deployment configuration"
      exit 7
    else
      verbose_output_command echo "OLM's global catalog namespace is: ${olm_global_catalog_namespace}"
    fi
  else
    verbose_output_command echo "OLM global catalog namespace set by user, skipping finding it inside script"
  fi
  echo "OLM should be working"
}

handle_catalog_source(){
  if ! verbose_output_command kubectl get CatalogSource opencloud-operators -n "${olm_global_catalog_namespace}"; then
    verbose_output_command echo "Applying opencloud Catalog Source"
    if ! cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: CatalogSource
metadata:
  name: opencloud-operators
  namespace: $olm_global_catalog_namespace
spec:
  displayName: IBMCS Operators
  publisher: IBM
  sourceType: grpc
  image: docker.io/ibmcom/ibm-common-service-catalog:latest
  updateStrategy:
    registryPoll:
      interval: 45m
EOF
    then
      echo "Error: Failed to apply Catalog Source"
      exit 11
    fi
  else
    verbose_output_command echo "opencloud-operators Catalog Source already exists"
  fi
  echo "Waiting for opencloud Catalog Source deployment to be ready"
  retries=50
  until [[ $retries == 0 || $new_cs_state == "READY" ]]; do
    new_cs_state=$(kubectl get catalogsource -n "${olm_global_catalog_namespace}" opencloud-operators -o jsonpath='{.status.connectionState.lastObservedState}' 2>/dev/null || echo "Waiting for Catalog Source to appear")
    if [[ $new_cs_state != "$cs_state" ]]; then
      cs_state=$new_cs_state
      echo "opencloud Catalog Source state: $cs_state"
    fi
    sleep 1
    retries=$((retries - 1))
  done
  if [ $retries == 0 ]; then
      echo "Error: CatalogSource \"opencloud-operators\" failed to reach state READY in 50 retries"
      exit 13
  fi
  echo "opencloud Catalog Source initialized"
}

handle_operator_group(){
  verbose_output_command echo "Counting operatorgroups at namespace $INSTALL_NAMESPACE"
  if ! operatorgroups_in_install_namespace=$(kubectl get OperatorGroup -n "${INSTALL_NAMESPACE}" -o name); then
    echo "Error: Failed to get OperatorGroup at namespace $INSTALL_NAMESPACE"
    exit 26
  fi
  if ! number_of_operatorgroups_in_install_namespace=$(echo "${operatorgroups_in_install_namespace}" | wc -w); then
    echo "Error: Failed to get number of OperatorGroups at namespace $INSTALL_NAMESPACE using 'wc -w' command"
    exit 27
  fi
  if [ "${number_of_operatorgroups_in_install_namespace}" -eq 0 ]; then
    verbose_output_command echo "Applying operatorgroup at namespace $INSTALL_NAMESPACE"
    if ! cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1
kind: OperatorGroup
metadata:
  name: operatorgroup
  namespace: $INSTALL_NAMESPACE
spec:
  targetNamespaces:
  - $INSTALL_NAMESPACE
EOF
    then
      echo "Error: Failed to apply OperatorGroup at namespace $INSTALL_NAMESPACE"
      exit 15
    fi
  elif [ "${number_of_operatorgroups_in_install_namespace}" -gt 1 ]; then
    echo "Error: There are more than one OperatorGroups at namespace ${INSTALL_NAMESPACE}:"
    echo "${operatorgroups_in_install_namespace}"
    echo "For subscription to work there should only exist one OperatorGroup, delete them and let this script create one"
    exit 28
  else
    verbose_output_command echo "OperatorGroup already exists in ${INSTALL_NAMESPACE} namespace, proceeding"
  fi
}

create_subscription(){
  if ! cat <<EOF | kubectl apply -f -
apiVersion: operators.coreos.com/v1alpha1
kind: Subscription
metadata:
  name: ibm-licensing-operator-app
  namespace: $INSTALL_NAMESPACE
spec:
  channel: $channel
  name: ibm-licensing-operator-app
  source: opencloud-operators
  sourceNamespace: $olm_global_catalog_namespace
EOF
  then
    echo "Error: Failed to apply Subscription at namespace $INSTALL_NAMESPACE"
    exit 16
  fi
}

handle_subscription(){
  if ! verbose_output_command kubectl get sub ibm-licensing-operator-app -n "${INSTALL_NAMESPACE}"; then
    create_subscription
  else
    verbose_output_command echo "Subscription already exists"
  fi
  echo "Checking Subscription and CSV status"
  existing_sub_channel=$(kubectl get sub -n "${INSTALL_NAMESPACE}" ibm-licensing-operator-app -o jsonpath='{.spec.channel}')
  if [[ "$existing_sub_channel" != "$channel" ]]; then
    echo "Subscription for License Service already exists but have different channel (found: $existing_sub_channel , expected: $channel ),"
    echo "Either delete existing subscription for License Service Operator or change channel option of the script to the found one"
    exit 22
  fi
  retries=55
  no_csv_name_in_sub_count=0
  until [[ $retries == 0 || $new_csv_phase == "Succeeded" ]]; do
    csv_name=$(kubectl get sub -n "${INSTALL_NAMESPACE}" ibm-licensing-operator-app -o jsonpath='{.status.currentCSV}')
    if [[ "$csv_name" == "" ]]; then
      no_csv_name_in_sub_count=$((no_csv_name_in_sub_count + 1))
      if [ $no_csv_name_in_sub_count -gt 9 ]; then
        no_csv_name_in_sub_count=0
        verbose_output_command "No CSV name in Subscription, deleting Subscription and creating it again"
        kubectl delete sub ibm-licensing-operator-app -n "${INSTALL_NAMESPACE}"
        sleep 5
        create_subscription
      fi
    else
      new_csv_phase=$(kubectl get csv -n "${INSTALL_NAMESPACE}" "${csv_name}" -o jsonpath='{.status.phase}' 2>/dev/null || echo "Waiting for CSV to appear")
      if [[ $new_csv_phase != "$csv_phase" ]]; then
        csv_phase=$new_csv_phase
        echo "$csv_name phase: $csv_phase"
        if [ "$csv_phase" == "Failed" ]; then
          echo "Error: Problem during installation of Subscription, try deleting Subscription and run script again."
          echo "If that won't help, check README for manual installation and troubleshooting"
          exit 17
        fi
      fi
    fi
    sleep 2
    retries=$((retries - 1))
  done
  if [ $retries == 0 ]; then
    echo "Error: CSV \"$csv_name\" failed to reach phase succeeded, try deleting Subscription and run script again."
    echo "If that won't help, check README for manual installation and troubleshooting"
    exit 18
  fi
  echo "Subscription and CSV should work"
}

handle_instance(){
  if ! verbose_output_command kubectl get IBMLicensing instance; then
    if ! cat <<EOF | kubectl apply -f -
apiVersion: operator.ibm.com/v1alpha1
kind: IBMLicensing
metadata:
  name: instance
spec:
  apiSecretToken: ibm-licensing-token
  datasource: datacollector
  httpsEnable: true
  instanceNamespace: $INSTALL_NAMESPACE
EOF
    then
      echo "Error: Failed to apply IBMLicensing instance at namespace $INSTALL_NAMESPACE"
      exit 19
    fi
  else
    verbose_output_command echo "IBMLicensing instance already exists"
  fi
  echo "Checking IBMLicensing instance status"
  retries=36
  until [[ $retries == 0 || $new_ibmlicensing_phase == "Running" ]]; do
    new_ibmlicensing_phase=$(kubectl get IBMLicensing instance -o jsonpath='{.status..phase}' 2>/dev/null || echo "Waiting for IBMLicensing pod to appear")
    if [[ $new_ibmlicensing_phase != "$ibmlicensing_phase" ]]; then
      ibmlicensing_phase=$new_ibmlicensing_phase
      echo "IBMLicensing Pod phase: $ibmlicensing_phase"
      if [ "$ibmlicensing_phase" == "Failed" ] ; then
        echo "Error: Problem during installation of IBMLicensing, try running script again when fixed, check README for post installation section and troubleshooting"
        exit 20
      fi
    fi
    sleep 10
    retries=$((retries - 1))
  done
  if [ $retries == 0 ]; then
    echo "Error: IBMLicensing instance pod failed to reach phase Running"
    exit 21
  fi
}

show_token(){
  if [ "$no_secret_output" != "1" ]; then
    if ! licensing_token=$(kubectl get secret ibm-licensing-token -o jsonpath='{.data.token}' -n "$INSTALL_NAMESPACE" | base64 -d) || [ "${licensing_token}" == "" ]; then
      verbose_output_command echo "Could not get ibm-licensing-token in $INSTALL_NAMESPACE, something might be wrong"
    else
      echo "License Service secret for accessing the API is: $licensing_token"
    fi
  fi
}

show_url(){
  if ! route_url=$(kubectl get route ibm-licensing-service-instance -o jsonpath='{.status.ingress[0].host}' -n "$INSTALL_NAMESPACE") || [ "${route_url}" == "" ]; then
    verbose_output_command echo "Could not get Route for License Service in $INSTALL_NAMESPACE, Route CRD might not be available at your cluster, or ingress option was chosen"
  else
    echo "License Service Route URL for accessing the API is: https://$route_url"
  fi
}

verbose_output_command(){
  if [ "$verbose" = "1" ]; then
    "$@"
  else
    "$@" 1> /dev/null 2>&1
  fi
}

##### Parse arguments

verbose=
olm_version=0.13.0
skip_olm_installation=
olm_global_catalog_namespace=
channel=v3
no_secret_output=

while [ "$1" != "" ]; do
  OPT=$1
  case $OPT in
    -h | --help )                                       usage
                                                        exit
                                                        ;;
    -v | --verbose )                                    verbose=1
                                                        ;;
    -o | --olm_version )                                shift
                                                        olm_version=$1
                                                        ;;
    -c | --olm_global_catalog_namespace )               shift
                                                        olm_global_catalog_namespace=$1
                                                        ;;
    -l | --channel )                                    shift
                                                        channel=$1
                                                        ;;
    -s | --skip_olm_installation )                      skip_olm_installation=1
                                                        ;;
    -n | --no-secret-output )                           no_secret_output=1
                                                        ;;
    * )                                                 echo "Error: wrong option: $OPT"
                                                        usage
                                                        exit 1
  esac
  if ! shift; then
    echo "Error: did not add needed arguments after option: $OPT"
    usage
    exit 4
  fi
done

##### Main

verify_command_line_processing
verify_kubectl
create_namespace
install_olm
handle_catalog_source
handle_operator_group
handle_subscription
handle_instance
show_token
show_url
echo "IBM License Service should be running, you can check post installation section in README to see possible configurations of IBM Licensing instance, and how to configure ingress/route if needed"