package datasciencepipelines

/* This is for DSP */
//+kubebuilder:rbac:groups="datasciencepipelinesapplications.opendatahub.io",resources=datasciencepipelinesapplications/status,verbs=update;patch;get
//+kubebuilder:rbac:groups="datasciencepipelinesapplications.opendatahub.io",resources=datasciencepipelinesapplications/finalizers,verbs=update;patch;get
//+kubebuilder:rbac:groups="datasciencepipelinesapplications.opendatahub.io",resources=datasciencepipelinesapplications,verbs=create;delete;list;update;watch;patch;get
//+kubebuilder:rbac:groups="image.openshift.io",resources=imagestreamtags,verbs=get
//+kubebuilder:rbac:groups="authentication.k8s.io",resources=tokenreviews,verbs=create;get
//+kubebuilder:rbac:groups="authorization.k8s.io",resources=subjectaccessreviews,verbs=create;get
