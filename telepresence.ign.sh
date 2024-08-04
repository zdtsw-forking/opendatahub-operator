export NS=${NS:-opendatahub-operator-system}
export DEPLOYMENT=$(oc get deployments -n $NS -o name | cut -d '/' -f 2)

env TELEPRESENCE_USE_OCP_IMAGE=NO telepresence \
      --swap-deployment "$DEPLOYMENT:manager"  \
      --namespace $NS \
      --mount /tmp/tp \
      --expose 9443:9443 \
      --run ./debug.ign.sh