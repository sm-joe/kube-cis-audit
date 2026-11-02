package audit

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
)

func TestEvaluateEffectiveGroupSecurityContext(t *testing.T) {
	tests := []struct {
		name          string
		effective     effectiveSecurityContext
		wantRunStatus string
		wantFSStatus  string
		wantRunSeen   string
		wantFSSeen    string
	}{
		{
			name:          "unset values are not assessed",
			wantRunStatus: "NOT_ASSESSED",
			wantFSStatus:  "NOT_ASSESSED",
			wantRunSeen:   "runAsGroup unset",
			wantFSSeen:    "fsGroup unset",
		},
		{
			name: "non-zero group IDs are informational",
			effective: effectiveSecurityContext{
				RunAsGroup: int64Ptr(10002),
				FSGroup:    int64Ptr(10003),
			},
			wantRunStatus: "INFO",
			wantFSStatus:  "INFO",
			wantRunSeen:   "runAsGroup=10002",
			wantFSSeen:    "fsGroup=10003",
		},
		{
			name: "GID zero warns without claiming root UID",
			effective: effectiveSecurityContext{
				RunAsGroup: int64Ptr(0),
			},
			wantRunStatus: "WARN",
			wantFSStatus:  "NOT_ASSESSED",
			wantRunSeen:   "runAsGroup=0 (root group)",
			wantFSSeen:    "fsGroup unset",
		},
		{
			name: "fsGroup zero warns independently",
			effective: effectiveSecurityContext{
				FSGroup: int64Ptr(0),
			},
			wantRunStatus: "NOT_ASSESSED",
			wantFSStatus:  "WARN",
			wantRunSeen:   "runAsGroup unset",
			wantFSSeen:    "fsGroup=0 (root group)",
		},
		{
			name: "negative group IDs fail validation",
			effective: effectiveSecurityContext{
				RunAsGroup: int64Ptr(-1),
				FSGroup:    int64Ptr(-2),
			},
			wantRunStatus: "FAIL",
			wantFSStatus:  "FAIL",
			wantRunSeen:   "runAsGroup=-1 (invalid)",
			wantFSSeen:    "fsGroup=-2 (invalid)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := evaluateEffectiveGroupSecurityContext(tt.effective)
			if len(got) != 2 {
				t.Fatalf("got %d assessments, want 2", len(got))
			}
			if got[0].Status != tt.wantRunStatus {
				t.Errorf("runAsGroup status = %q, want %q", got[0].Status, tt.wantRunStatus)
			}
			if got[1].Status != tt.wantFSStatus {
				t.Errorf("fsGroup status = %q, want %q", got[1].Status, tt.wantFSStatus)
			}
			if got[0].Observed != tt.wantRunSeen {
				t.Errorf("runAsGroup observed = %q, want %q", got[0].Observed, tt.wantRunSeen)
			}
			if got[1].Observed != tt.wantFSSeen {
				t.Errorf("fsGroup observed = %q, want %q", got[1].Observed, tt.wantFSSeen)
			}
		})
	}
}

func TestGroupAssessmentUsesResolvedPodAndContainerValues(t *testing.T) {
	podSC := &corev1.PodSecurityContext{
		RunAsGroup: int64Ptr(10002),
		FSGroup:    int64Ptr(10003),
	}
	containerSC := &corev1.SecurityContext{
		RunAsGroup: int64Ptr(0),
	}

	effective := resolveEffectiveSecurityContext(podSC, containerSC)
	got := evaluateEffectiveGroupSecurityContext(effective)

	if got[0].Status != "WARN" || got[0].Observed != "runAsGroup=0 (root group)" {
		t.Fatalf("container runAsGroup override not assessed: %+v", got[0])
	}
	if got[1].Status != "INFO" || got[1].Observed != "fsGroup=10003" {
		t.Fatalf("Pod fsGroup inheritance not assessed: %+v", got[1])
	}
}
