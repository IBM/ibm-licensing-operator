#!/bin/bash
#
# Copyright 2023 IBM Corporation
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
export COMMON_SERVICE_NAMESPACE=ibm-common-services
export LS_NAMESPACE=$COMMON_SERVICE_NAMESPACE$SUFIX

if [ -z "$1" ]
then
   export YOUR_CLUSTER=false
else
   export YOUR_CLUSTER=true
fi

if $MUTEX_ON
then
    echo "add mutex $CN"
    kubectl get configmap |grep "$CNAME"
    export ret=$?
    while [ $ret -eq 0 ] ; do
        echo "Wait. ON cluster there is other job"
        sleep 13
        kubectl get configmap |grep "$CNAME"
        ret=$?
    done
    kubectl create configmap "$CN"
fi

export ret=0
echo "Clear namespace $LS_NAMESPACE inside cluster."
kubectl delete namespace "$LS_NAMESPACE"
while [ $ret -eq 0 ] ; do
    sleep 3
    kubectl get namespace |grep "$LS_NAMESPACE"
    ret=$?
done
echo There is not "$LS_NAMESPACE" namespace.

if $YOUR_CLUSTER
then
    echo "You choose the cluster, we do not removed CRDs and other namespaces"
else
    echo "Check if any other tests works on Cluster and clear from old ibm-licensing resources"

    export ret=0
    kubectl delete namespace "$(kubectl get namespace | grep "$COMMON_SERVICE_NAMESPACE" | awk '{print $1}'| tr '\n' ' , ')"
    while [ $ret -eq 0 ] ; do
        sleep 3
        kubectl get namespace |grep "$COMMON_SERVICE_NAMESPACE"
        ret=$?
    done

    echo DELETE CRD ibmlicensings.operator.ibm.com and meterdefinitions.marketplace.redhat.com
    kubectl delete crd ibmlicensings.operator.ibm.com
    kubectl delete crd meterdefinitions.marketplace.redhat.com

    export ret=0
    while [ $ret -eq 0 ] ; do
       sleep 2
       kubectl get crd ibmlicensings.operator.ibm.com
       ret=$?
    done
fi

if $MUTEX_OFF
then
    echo "remove mutex $CN"
    kubectl delete configmap "$CN"
fi
echo The Cluster is clear