package audit

import "time"

type Status string

const (
	Pass        Status = "PASS"
	Fail        Status = "FAIL"
	NotAssessed Status = "NOT_ASSESSED"
	Warn        Status = "WARN"
)

type Evidence struct {
	ClusterID, Profile, Source string
	CollectedAt                time.Time
	MaxAge                     time.Duration
	Complete                   bool
	Errors                     []string
}
type AuditConfig struct {
	Evidence                 Evidence
	PolicyFileConfigured     *bool
	LogPathConfigured        *bool
	LogMaxAge                *int
	LogMaxBackup             *int
	LogMaxSize               *int
	WebhookBackendConfigured *bool
}
type PolicyRule struct {
	Level                                                            string
	Users, UserGroups, Verbs, Resources, Namespaces, NonResourceURLs []string
}
type PolicyInventory struct {
	Evidence     Evidence
	Rules        []PolicyRule
	DefaultLevel string
	HasCatchAll  bool
}
type Result struct {
	ID, Title string
	Status    Status
	Reason    string
	Source    string
}
