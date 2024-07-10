package labels

import (
	corev1 "k8s.io/api/core/v1"
)

const (
	keyCompositionID = "krateo.io/composition-id"
	keyPatchedBy     = "krateo.io/patched-by"
)

func WasPatchedByKrateo(obj *corev1.Event) bool {
	labels := obj.GetLabels()
	if len(labels) == 0 {
		return false
	}

	_, ok := labels[keyPatchedBy]
	return ok
}

func CompositionID(obj *corev1.Event) string {
	labels := obj.GetLabels()
	if len(labels) == 0 {
		return ""
	}

	if val, ok := labels[keyCompositionID]; ok {
		return val
	}

	return ""
}
