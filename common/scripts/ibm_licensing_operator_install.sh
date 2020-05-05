#!/bin/bash
#
# Copyright 2020 IBM Corporation
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

# TODO: add options and describe them
usage()
{
   # Display usage
  echo "description: A script to install IBM License Service via Operator."
  echo "usage: $0 [--interactive | -i] [--verbose | -v] [--help | -h]"
  echo "options:"
#  echo "[--interactive | -i] - adds user questions, will ask for versions etc."
  echo "[--verbose | -v] - verbose logs from installation"
  echo "[--olm_version | -o] <version_number> - what version of OLM should be installed if it doesn't exist,"
  echo "by default olm_version=0.13.0"
  echo "[--skip_olm_installation | -s] - skips installation of OLM, but olm global catalog namespace still needs to be found."
  echo "[--olm_global_catalog_namespace | -c] <OLM global catalog namespace> - script will not try to find olm global catalog namespace when set."
  echo "You can read more about OLM global catalog namespace here: https://github.com/operator-framework/operator-lifecycle-manager/blob/master/doc/install/install.md"
  echo "[--help | -h] - shows usage"
  echo
}

verify_command_line_processing(){
  # Test code to verify command line processing
#  if [ "$interactive" = "1" ]; then
#    verbose_output_command echo "interactive is on"
#  else
#    verbose_output_command echo "interactive is off"
#  fi
  verbose_output_command echo "olm version is ${olm_version}"
}

verify_kubectl(){
  if ! verbose_output_command kubectl version; then
    echo "kubectl command does not seems to work"
    echo "try to install it and setup config for your cluster where you want to install IBM License Service"
    exit 2
  fi
}

create_namespace(){
  if ! verbose_output_command kubectl get namespace "${INSTALL_NAMESPACE}"; then
    echo "Creating namespace ${INSTALL_NAMESPACE}"
    if ! kubectl create namespace "${INSTALL_NAMESPACE}"; then
      echo "kubectl command cannot create needed namespace"
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
      echo "CSV CRD exists does not exists, installing OLM with version ${olm_version}"
      if ! curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/"${olm_version}"/install.sh | bash -s "${olm_version}"; then
        echo "Failed to install OLM"
        echo "You can try to install OLM from here https://github.com/operator-framework/operator-lifecycle-manager/releases and continue installation while skipping OLM part"
        exit 5
      fi
    else
      verbose_output_command echo "OLM seems to be installed"
    fi
  else
    verbose_output_command echo "Skipping OLM installation"
  fi
  if [ "${olm_global_catalog_namespace}" == "" ]; then
    verbose_output_command echo "Trying to get namespace where OLM's packageserver is installed"
    if ! olm_namespace=$(verbose_output_command kubectl get csv --all-namespaces -l olm.version -o jsonpath="{.items[?(@.metadata.name=='packageserver')].metadata.namespace}") || [ "${olm_namespace}" == "" ]; then
      echo "Failed to get namespace where OLM's packageserver is installed, which is needed for finding OLM's global catalog namespace, make sure you have OLM installed"
      echo "You can try to install OLM from here https://github.com/operator-framework/operator-lifecycle-manager/releases"
      echo "If you can find OLM's global catalog namespace yourself try setting parameter --olm_global_catalog_namespace parameter of this script"
      echo "On OpenShift Container Platform this probably is 'openshift-marketplace', but for older versions and for custom OLM installation it might be 'olm', but you might verify it by looking for OLM's packageserver deployment configuration"
      exit 6
    else
      verbose_output_command echo "Namespace where OLM's packageserver is installed is: ${olm_namespace}"
    fi
    verbose_output_command echo "Trying to get OLM's global catalog namespace so that catalog needed by IBM Licensing can be accessed in any watched namespace."
    if ! olm_global_catalog_namespace=$(verbose_output_command kubectl get deployment --namespace="${olm_namespace}" packageserver -o yaml | grep -A 1 -i global-namespace | tail -1 | cut -d "-" -f 2- | sed -e 's/^[ \t]*//') || [ "${olm_global_catalog_namespace}" == "" ]; then
      echo "Failed to find OLM's global catalog namespace where catalog for IBM Licensign needs to be installed"
      echo "If you can find it yourself try setting parameter --olm_global_catalog_namespace parameter of this script"
      echo "On OpenShift Container Platform this probably is 'openshift-marketplace', but for older versions and for custom OLM installation it might be 'olm', but you might verify it by looking for OLM's packageserver deployment configuration"
      exit 7
    else
      verbose_output_command echo "OLM's global catalog namespace is: ${olm_global_catalog_namespace}"
    fi
  else
    verbose_output_command echo "OLM global catalog namespace set by user, skipping finding it inside script"
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

#interactive=
verbose=
olm_version=0.13.0
skip_olm_installation=

while [ "$1" != "" ]; do
  OPT=$1
  case $OPT in
    -h | --help )                                       usage
                                                        exit
                                                        ;;
#    -i | --interactive )                                interactive=1
#                                                        ;;
    -v | --verbose )                                    verbose=1
                                                        ;;
    -o | --olm_version )                                shift
                                                        olm_version=$1
                                                        ;;
    -c | --olm_global_catalog_namespace )               shift
                                                        olm_global_catalog_namespace=$1
                                                        ;;
    -s | --skip_olm_installation )                      skip_olm_installation=1
                                                        ;;
    * )                                                 echo "Error, wrong option: $OPT"
                                                        usage
                                                        exit 1
  esac
  if ! shift; then
    echo "Error, did not add needed arguments after option: $OPT"
    usage
    exit 4
  fi
done

##### Main

verify_command_line_processing
verify_kubectl
create_namespace
install_olm
#install_marketplace
#create_operator_source
#create_operator_group
#create_subscription
#create_instance
