package apirbac

import "time"

type Status string

const (
	Pass        Status = "PASS"
	Fail        Status = "FAIL"
	NotAssessed Status = "NOT_ASSESSED"
	Warn        Status = "WARN"
)

type Evidence struct {
	ClusterID   string
	Profile     string
	CollectedAt time.Time
	MaxAge      time.Duration
	Complete    bool
	Source      string
	Errors      []string
}
type APIConfig struct {
	Evidence            Evidence
	AnonymousAuth       *bool
	InsecureBindAddress string
	InsecurePort        *int
	SecurePort          *int
	AuthorizationModes  []string
	AuditEnabled        *bool
}
type Rule struct {
	Verbs           []string
	APIGroups       []string
	Resources       []string
	ResourceNames   []string
	NonResourceURLs []string
}
type Role struct {
	Name, Namespace string
	Cluster         bool
	Rules           []Rule
}
type Binding struct {
	Name, Namespace, RoleName, RoleKind string
	Cluster                             bool
	Subjects                            []Subject
}
type Subject struct{ Kind, Name, Namespace string }
type RBACInventory struct {
	Evidence Evidence
	Roles    []Role
	Bindings []Binding
}
type Result struct {
	ID, Title      string
	Status         Status
	Reason         string
	EvidenceSource string
}
