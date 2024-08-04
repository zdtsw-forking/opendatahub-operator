#!/bin/bash

rm -rf /opt/manifests/
sudo mkdir -p {/var/run/secrets/kubernetes.io/serviceaccount,/opt/manifests}
sudo chown -R $(whoami) {/var/run/secrets/kubernetes.io/serviceaccount,/opt/manifests}

# get manifests
./get_all_manifests.sh
ln -s $(pwd)/opt/manifests/** /opt/manifests/

# get operator pod secrets
export NS=${NS:-opendatahub-operator-system}
export POD=$(oc get pods -n $NS -o json |  jq -r '.items[] | select(.metadata.name | startswith("opendatahub")) | .metadata.name')

oc exec -n $NS $POD -- cat /var/run/secrets/kubernetes.io/serviceaccount/namespace > /var/run/secrets/kubernetes.io/serviceaccount/namespace
oc exec -n $NS $POD -- cat /var/run/secrets/kubernetes.io/serviceaccount/token > /var/run/secrets/kubernetes.io/serviceaccount/token
oc exec -n $NS $POD -- cat /var/run/secrets/kubernetes.io/serviceaccount/ca.crt > /var/run/secrets/kubernetes.io/serviceaccount/ca.crt

# start debugger
air