package status

import (
	conditionsv1 "github.com/openshift/custom-resource-status/conditions/v1"
	corev1 "k8s.io/api/core/v1"
)

// Component init condition.
// type: <component>Ready
// status: Unknown
// reason: ReconcileStart
func InitComponentCondition(componentName string, enabled bool) conditionsv1.Condition {
	managementstatue := "not Managed"
	if enabled {
		managementstatue = "Managed"
	}
	return conditionsv1.Condition{
		Type:    conditionsv1.ConditionType(componentName + PhaseReady),
		Status:  corev1.ConditionUnknown,
		Reason:  ReconcileStartReason,
		Message: "Component managementStatus is " + managementstatue,
	}
}

// Component reconilce success.
// type: <component>Ready
// status: True
// reason: ReconcileCompleted
func SuccessComponentCondition(componentName string) conditionsv1.Condition {
	return conditionsv1.Condition{
		Type:    conditionsv1.ConditionType(componentName + PhaseReady),
		Status:  corev1.ConditionTrue,
		Reason:  ReconcileSuccessReason,
		Message: "Component reconciled successfully",
	}
}

// Component reconcile failed.
// type: <component>Ready
// status: False
// reason: ReconcileFailed
// message: <derive from err>
func FailedComponentCondition(componentName string, err error) (conditionsv1.Condition, error) {
	FailedCondition := setFailedComponentCondition(componentName)
	FailedCondition.Message = err.Error()
	return FailedCondition, err
}

// Component failed reconcile condition, called by FailedComponentCondition().
func setFailedComponentCondition(componentName string) conditionsv1.Condition {
	return conditionsv1.Condition{
		Type:   conditionsv1.ConditionType(componentName + PhaseReady),
		Status: corev1.ConditionFalse,
		Reason: ReconcileFailedReason,
	}
}

// Special handling on DSPA for Argo.
func ArgoExistCondition(reason string, message string) conditionsv1.Condition {
	return conditionsv1.Condition{
		Type:    CapabilityDSPv2Argo,
		Status:  corev1.ConditionFalse,
		Reason:  reason,
		Message: message,
	}
}
