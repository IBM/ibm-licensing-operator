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
export CNAME=ibm-ls-automation-
export CN="$CNAME$SUFIX"
export LS_NAMESPACE=ibm-common-services$SUFIX
export IKS_CLUSTER_ZONE=dal10
#ibmcloud ks vlan ls --zone $IKS_CLUSTER_ZONE
export IKS_CLUSTER_PRIVATE_VLAN=2918270
export IKS_CLUSTER_PUBLIC_VLAN=2918268
export VERSION="4.4_openshift"
export IKS_CLUSTER_FLAVOR=b3c.4x16.encrypted
export clustername=$1
export IKS_CLUSTER_TAG_NAMES="owner:artur.obrzut,team:CP4MCM,Usage:temp,Usage_desc:certification_tests,Review_freq:month"
if [ -z "$clustername" ]
then
   echo "try to find cluster $CNAME"
   if ibmcloud oc cluster ls |grep $CNAME  | grep -m1 normal > once.txt
   then
      awk '{print $1}' < once.txt > ./clustername.txt
      CN=$(cat ./clustername.txt)
      export CN
      echo "Cluster $CN was found"
   fi
else
  echo "You choice to use cluster $clustername"
  echo "$clustername" > ./clustername.txt
  CN=$(cat ./clustername.txt)
  export CN
fi

if ibmcloud oc cluster ls |grep "$CN"
then
   echo "Cluster exists"
else
   echo "Start creating cluster $CN"
   ibmcloud oc cluster create classic --name "$CN" --flavor $IKS_CLUSTER_FLAVOR --hardware shared --workers 1 --zone "$IKS_CLUSTER_ZONE" --public-vlan "$IKS_CLUSTER_PUBLIC_VLAN" --private-vlan "$IKS_CLUSTER_PRIVATE_VLAN"  --version "$VERSION"  --public-service-endpoint
   sleep 10

   ibmcloud oc cluster ls | grep "$CN" | grep normal
   ret=$?
   while [ $ret -ne 0 ] ; do
     echo "Wait for cluster creation 30s"
     sleep 30
     ibmcloud oc cluster ls |grep "$CN" | grep normal
     ret=$?
   done
   echo "Cluster was created $CN"
   echo "now tag it"
   ibmcloud resource tag-attach --resource-name "$CN" --tag-names "$IKS_CLUSTER_TAG_NAMES"

fi

echo "Try to connect to the cluster $CN"
ibmcloud oc cluster config --cluster "$CN"  --yaml --admin
kubectl config current-context > /dev/null

if ! kubectl get nodes
then
  echo "ERROR CANNOT GET NODES!!!"
  exit 1
fi
i=0
kubectl get namespace |grep "$LS_NAMESPACE"
ret=$?
while [ $ret -eq 0 ] ; do
    echo "There is namespace $LS_NAMESPACE inside cluster. Wait 2 minutes for removing it. Please remove it"
    sleep 30
    i=$i+1
    if [[ $i -gt 4 ]]
    then
        echo "Delete tests data like ibm-common-services"
        kubectl delete namespace "$LS_NAMESPACE"

        #kubectl delete crd ibmlicensings.operator.ibm.com
        #kubectl delete crd ibmlicenseservicereporters.operator.ibm.com
    fi
    kubectl get namespace |grep "$LS_NAMESPACE"
    ret=$?
done
echo There is not "$LS_NAMESPACE" namespace.

