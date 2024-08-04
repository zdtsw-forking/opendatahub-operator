VERSION=2.15.4444
IMAGE_TAG_BASE=quay.io/wenzhou/opendatahub-operator
E2E_TEST_FLAGS="--skip-deletion=true" -timeout 10m
IMG_TAG=$(VERSION)
IMAGE_BUILD_FLAGS=--build-arg USE_LOCAL=true
DEFAULT_MANIFESTS_PATH=./opt/manifests
OPERATOR_NAMESPACE=openshift-operators