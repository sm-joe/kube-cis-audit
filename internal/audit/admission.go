package audit

import (
	"fmt"
	"strings"

	admissionv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ruleMatchesPodCreateUpdate reports whether a rule can match CREATE or UPDATE
// requests for core/v1 Pods. It deliberately checks resource matching rather
// than webhook or policy names.
func ruleMatchesPodCreateUpdate(rule admissionv1.RuleWithOperations) bool {
	return ruleMatchesPodOperation(rule, admissionv1.Create) || ruleMatchesPodOperation(rule, admissionv1.Update)
}

func ruleMatchesPodOperation(rule admissionv1.RuleWithOperations, operation admissionv1.OperationType) bool {
	if !containsOperation(rule.Operations, operation) {
		return false
	}
	if !containsValue(rule.APIGroups, "") && !containsValue(rule.APIGroups, "*") {
		return false
	}
	if !containsValue(rule.APIVersions, "v1") && !containsValue(rule.APIVersions, "*") {
		return false
	}
	if rule.Scope != nil && *rule.Scope == admissionv1.ClusterScope {
		return false
	}
	for _, resource := range rule.Resources {
		switch resource {
		case "*", "*/*", "pods":
			return true
		}
	}
	return false
}

func containsOperation(values []admissionv1.OperationType, wanted admissionv1.OperationType) bool {
	for _, value := range values {
		if value == wanted || value == admissionv1.OperationAll {
			return true
		}
	}
	return false
}

func containsValue(values []string, wanted string) bool {
	for _, value := range values {
		if value == wanted {
			return true
		}
	}
	return false
}

func matchResourcesCanMatchPods(match *admissionv1.MatchResources, bindingConstraint bool) bool {
	if match == nil {
		return bindingConstraint
	}
	// For a VAP policy, resourceRules are required. For a binding, an empty
	// resourceRules list adds no resource constraint to policy matching.
	includeCreate := bindingConstraint && len(match.ResourceRules) == 0
	includeUpdate := includeCreate
	for _, namedRule := range match.ResourceRules {
		includeCreate = includeCreate || ruleMatchesPodOperation(namedRule.RuleWithOperations, admissionv1.Create)
		includeUpdate = includeUpdate || ruleMatchesPodOperation(namedRule.RuleWithOperations, admissionv1.Update)
	}
	if !includeCreate && !includeUpdate {
		return false
	}
	// Exclusions take precedence over inclusion rules. A Pod exclusion without
	// resourceNames excludes all Pod names for the matched operation.
	excludeCreate, excludeUpdate := false, false
	for _, excluded := range match.ExcludeResourceRules {
		if len(excluded.ResourceNames) != 0 {
			continue
		}
		excludeCreate = excludeCreate || ruleMatchesPodOperation(excluded.RuleWithOperations, admissionv1.Create)
		excludeUpdate = excludeUpdate || ruleMatchesPodOperation(excluded.RuleWithOperations, admissionv1.Update)
	}
	return (includeCreate && !excludeCreate) || (includeUpdate && !excludeUpdate)
}

func policySelectorSummary(match *admissionv1.MatchResources) string {
	if match == nil {
		return "selectors=none"
	}
	parts := make([]string, 0, 2)
	if match.NamespaceSelector != nil {
		parts = append(parts, fmt.Sprintf("namespaceSelector(labels=%d, expressions=%d)", len(match.NamespaceSelector.MatchLabels), len(match.NamespaceSelector.MatchExpressions)))
	}
	if match.ObjectSelector != nil {
		parts = append(parts, fmt.Sprintf("objectSelector(labels=%d, expressions=%d)", len(match.ObjectSelector.MatchLabels), len(match.ObjectSelector.MatchExpressions)))
	}
	if len(match.ResourceRules) > 0 {
		parts = append(parts, fmt.Sprintf("resourceRules=%d", len(match.ResourceRules)))
	}
	if len(match.ExcludeResourceRules) > 0 {
		parts = append(parts, fmt.Sprintf("excludeResourceRules=%d", len(match.ExcludeResourceRules)))
	}
	if len(parts) == 0 {
		return "selectors=none; resourceRules=0; excludeResourceRules=0"
	}
	return strings.Join(parts, "; ")
}

func failurePolicyValue(value *admissionv1.FailurePolicyType) string {
	if value == nil {
		return "Fail (API default)"
	}
	return string(*value)
}

func labelSelectorState(selector *metav1.LabelSelector) string {
	if selector == nil {
		return "unset"
	}
	return fmt.Sprintf("set(labels=%d, expressions=%d)", len(selector.MatchLabels), len(selector.MatchExpressions))
}
