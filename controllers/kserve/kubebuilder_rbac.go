package kserve

// +kubebuilder:rbac:groups="serving.kserve.io",resources=trainedmodels/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=trainedmodels,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=servingruntimes/status,verbs=update;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=servingruntimes/finalizers,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=servingruntimes,verbs=*
// +kubebuilder:rbac:groups="serving.kserve.io",resources=predictors/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=predictors/finalizers,verbs=update;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=predictors,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferenceservices/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferenceservices/finalizers,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferenceservices,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferencegraphs/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=inferencegraphs,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterservingruntimes/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterservingruntimes/finalizers,verbs=create;delete;list;update;watch;patch;get
// +kubebuilder:rbac:groups="serving.kserve.io",resources=clusterservingruntimes,verbs=create;delete;list;update;watch;patch;get

// +kubebuilder:rbac:groups="serving.knative.dev",resources=services/status,verbs=update;patch;delete;get
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services/finalizers,verbs=create;delete;list;watch;update;patch;get
// +kubebuilder:rbac:groups="serving.knative.dev",resources=services,verbs=create;delete;list;watch;update;patch;get

/* Serverless prerequisite */
// +kubebuilder:rbac:groups="networking.istio.io",resources=gateways,verbs=*
// +kubebuilder:rbac:groups="operator.knative.dev",resources=knativeservings,verbs=*
// +kubebuilder:rbac:groups="config.openshift.io",resources=ingresses,verbs=get
