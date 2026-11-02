package audit

import (
	"fmt"
)

// securityContextFieldAssessment is a field-level observation. It deliberately
// does not represent a CIS benchmark result: a non-zero GID alone does not
// prove least-privilege group membership or effective filesystem permissions.
type securityContextFieldAssessment struct {
	Field    string
	Observed string
	Status   string
	Details  string
}

// evaluateEffectiveGroupSecurityContext assesses the effective primary group
// and Pod fsGroup values resolved by resolveEffectiveSecurityContext.
//
// Status semantics:
//   - FAIL: an invalid negative ID was supplied (not normally accepted by the API).
//   - WARN: GID 0 is explicitly configured.
//   - INFO: an explicit non-zero ID is configured; authorization is not assessed.
//   - NOT_ASSESSED: the field is unset; no runtime or volume defaults are inferred.
//
// runAsGroup and fsGroup are assessed separately. GID 0 does not establish that
// the process runs as UID 0. fsGroup does not guarantee effective ownership or
// permissions for every volume or CSI driver.
func evaluateEffectiveGroupSecurityContext(effective effectiveSecurityContext) []securityContextFieldAssessment {
	return []securityContextFieldAssessment{
		assessGroupID("runAsGroup", effective.RunAsGroup),
		assessFSGroup(effective.FSGroup),
	}
}

func assessGroupID(field string, value *int64) securityContextFieldAssessment {
	if value == nil {
		return securityContextFieldAssessment{
			Field:    "securityContext." + field,
			Observed: field + " unset",
			Status:   "NOT_ASSESSED",
			Details:  "No effective primary group is specified; image and runtime defaults are not inspected.",
		}
	}

	if *value < 0 {
		return securityContextFieldAssessment{
			Field:    "securityContext." + field,
			Observed: fmt.Sprintf("%s=%d (invalid)", field, *value),
			Status:   "FAIL",
			Details:  "The group ID is negative and is not a valid Kubernetes security-context value.",
		}
	}
	if *value == 0 {
		return securityContextFieldAssessment{
			Field:    "securityContext." + field,
			Observed: field + "=0 (root group)",
			Status:   "WARN",
			Details:  "The primary group is GID 0. This does not establish UID 0 or prove that the process runs as root.",
		}
	}

	return securityContextFieldAssessment{
		Field:    "securityContext." + field,
		Observed: fmt.Sprintf("%s=%d", field, *value),
		Status:   "INFO",
		Details:  "An explicit non-zero primary GID is configured; group membership and permissions are not evaluated.",
	}
}

func assessFSGroup(value *int64) securityContextFieldAssessment {
	if value == nil {
		return securityContextFieldAssessment{
			Field:    "securityContext.fsGroup",
			Observed: "fsGroup unset",
			Status:   "NOT_ASSESSED",
			Details:  "No fsGroup is specified. This alone does not establish insecure volume permissions.",
		}
	}

	if *value < 0 {
		return securityContextFieldAssessment{
			Field:    "securityContext.fsGroup",
			Observed: fmt.Sprintf("fsGroup=%d (invalid)", *value),
			Status:   "FAIL",
			Details:  "The fsGroup value is negative and is not a valid Kubernetes security-context value.",
		}
	}
	if *value == 0 {
		return securityContextFieldAssessment{
			Field:    "securityContext.fsGroup",
			Observed: "fsGroup=0 (root group)",
			Status:   "WARN",
			Details:  "Volumes honoring fsGroup may receive group ownership or access associated with GID 0; effective volume permissions are not verified.",
		}
	}

	return securityContextFieldAssessment{
		Field:    "securityContext.fsGroup",
		Observed: fmt.Sprintf("fsGroup=%d", *value),
		Status:   "INFO",
		Details:  "An explicit non-zero fsGroup is configured; effective ownership and access depend on volume and CSI-driver behavior.",
	}
}
