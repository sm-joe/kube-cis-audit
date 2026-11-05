package conformance

type Mapping struct {
	ControlID    string `json:"control_id"`
	Title        string `json:"title"`
	Profile      string `json:"profile"`
	Status       string `json:"status"`
	EvidenceKind string `json:"evidence_kind,omitempty"`
	SourceFile   string `json:"source_file,omitempty"`
}
type Profile struct {
	Name             string    `json:"name"`
	BenchmarkVersion string    `json:"benchmark_version"`
	KubernetesRange  string    `json:"kubernetes_range,omitempty"`
	Controls         []Mapping `json:"controls"`
}
type Inventory struct {
	Profiles []Profile `json:"profiles"`
}
type Finding struct{ Profile, ControlID, Code, Detail string }
type Report struct {
	Profiles     int       `json:"profiles"`
	Controls     int       `json:"controls"`
	Candidate    int       `json:"candidate"`
	Verified     int       `json:"verified"`
	Disabled     int       `json:"disabled"`
	Findings     []Finding `json:"findings"`
	ReleaseReady bool      `json:"release_ready"`
}
