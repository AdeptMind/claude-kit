package roles

// RolesConfig is the top-level structure of roles.yaml.
type RolesConfig struct {
	Roles map[string]RoleDefinition `yaml:"roles"`
}

// RoleDefinition describes a single role's working methods.
type RoleDefinition struct {
	Description string          `yaml:"description"`
	Workflow    WorkflowConfig  `yaml:"workflow"`
	Reporting   ReportingConfig `yaml:"reporting"`
	Analysis    AnalysisConfig  `yaml:"analysis"`
	Phases      PhaseAssignment `yaml:"phases,omitempty"`
}

// WorkflowConfig controls gates and checks for a role.
type WorkflowConfig struct {
	InterRoundGate  bool   `yaml:"inter_round_gate"`
	ClientChallenge bool   `yaml:"client_challenge"`
	ReadinessCheck  string `yaml:"readiness_check"`
	DoneCheck       string `yaml:"done_check"`
}

// ReportingConfig controls how a role reports progress.
type ReportingConfig struct {
	Language string `yaml:"language"`
	Proof    string `yaml:"proof"`
}

// AnalysisConfig controls how a role analyzes requirements.
type AnalysisConfig struct {
	Traceability    string `yaml:"traceability"`
	ChallengeSoThat bool   `yaml:"challenge_so_that"`
}

// PhaseAssignment maps BMAD phases to role names (only used by the "all" role).
type PhaseAssignment struct {
	Break   []string `yaml:"break,omitempty"`
	Model   []string `yaml:"model,omitempty"`
	Analyze []string `yaml:"analyze,omitempty"`
	Act     []string `yaml:"act,omitempty"`
	Deliver []string `yaml:"deliver,omitempty"`
}
