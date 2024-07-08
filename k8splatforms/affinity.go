package k8splatforms

import (
	"slices"

	"github.com/wantedly/container-platform-tools/dockerplatforms"
	corev1 "k8s.io/api/core/v1"
)

func PodPlatforms(pod *corev1.Pod, nodePlatforms dockerplatforms.DockerPlatformList) dockerplatforms.DockerPlatformList {
	platforms := make(dockerplatforms.DockerPlatformList, 0, len(nodePlatforms))
	for _, nodePlatform := range nodePlatforms {
		if EvaluatePodAffinity(pod, nodePlatform) {
			platforms = append(platforms, nodePlatform)
		}
	}
	return platforms
}

func EvaluatePodAffinity(pod *corev1.Pod, platform dockerplatforms.DockerPlatform) bool {
	var nodeSelector *corev1.NodeSelector
	if pod.Spec.Affinity != nil {
		if pod.Spec.Affinity.NodeAffinity != nil {
			nodeSelector = pod.Spec.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
		}
	}
	return EvaluateSelectors(pod.Spec.NodeSelector, nodeSelector, platform)
}

func EvaluateSelectors(labelSelector map[string]string, nodeSelector *corev1.NodeSelector, platform dockerplatforms.DockerPlatform) bool {
	return EvaluateLabelSelector(labelSelector, platform) && EvaluateNodeSelector(nodeSelector, platform)
}

func EvaluateLabelSelector(labelSelector map[string]string, platform dockerplatforms.DockerPlatform) bool {
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L308-L310
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L324
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/apimachinery/pkg/labels/selector.go#L938-L954
	for key, expectedValue := range labelSelector {
		value, ok := getLabel(key, platform)
		if ok && value != expectedValue {
			return false
		}
	}
	return true
}

func EvaluateNodeSelector(nodeSelector *corev1.NodeSelector, platform dockerplatforms.DockerPlatform) bool {
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L329
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L81-L103
	// The terms are ORed, but if it is missing, it is considered match-all
	if nodeSelector == nil {
		return true
	}

	for _, term := range nodeSelector.NodeSelectorTerms {
		if evaluateNodeSelectorTerm(term, platform) {
			return true
		}
	}
	return false
}

func evaluateNodeSelectorTerm(nodeSelectorTerm corev1.NodeSelectorTerm, platform dockerplatforms.DockerPlatform) bool {
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L212-L251
	// Note also that, empty case is in fact handled here:
	// - https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L173
	// - https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L194
	// This means the empty expression list is match-all rather than match-none, contrary to what nodeSelectorRequirementsAsSelector suggests

	// Each match expression is ANDed
	for _, expr := range nodeSelectorTerm.MatchExpressions {
		if !evaluateMatchExpression(expr, platform) {
			return false
		}
	}
	// Ignore match fields (which currently only matches against `metadata.name`)
	// as they are not relevant to platform

	return true
}

func evaluateMatchExpression(expr corev1.NodeSelectorRequirement, platform dockerplatforms.DockerPlatform) bool {
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L212-L251
	value, ok := getLabel(expr.Key, platform)
	if !ok {
		// Assume it matches
		return true
	}
	switch expr.Operator {
	case corev1.NodeSelectorOpIn:
		return slices.Contains(expr.Values, value)
	case corev1.NodeSelectorOpNotIn:
		return !slices.Contains(expr.Values, value)
	case corev1.NodeSelectorOpExists:
		return true
	case corev1.NodeSelectorOpDoesNotExist:
		return false
	default:
		// Assume it matches
		return true
	}
}

func getLabel(key string, platform dockerplatforms.DockerPlatform) (string, bool) {
	// https://github.com/kubernetes/kubernetes/blob/v1.30.2/staging/src/k8s.io/component-helpers/scheduling/corev1/nodeaffinity/nodeaffinity.go#L212-L251
	switch key {
	case "kubernetes.io/arch", "beta.kubernetes.io/arch":
		return platform.Architecture, true
	case "kubernetes.io/os", "beta.kubernetes.io/os":
		return platform.OS, true
	default:
		// Assume it matches
		return "", false
	}
}
