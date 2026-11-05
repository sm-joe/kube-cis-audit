package admission

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
type PluginConfig struct {
	Evidence                        Evidence
	EnabledPlugins, DisabledPlugins []string
	ConfigKnown                     bool
}
type Policy struct {
	Name, Kind, Enforcement string
	Ready                   bool
	HasValidation           bool
	HasMutation             bool
	FailurePolicy           string
	MatchResources          []string
}
type PolicyInventory struct {
	Evidence Evidence
	Policies []Policy
}
type Result struct {
	ID, Title string
	Status    Status
	Reason    string
	Source    string
}
