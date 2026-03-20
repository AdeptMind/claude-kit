package policy

// PolicySpec is the top-level policy definition.
// Exactly one of Profile or Extends must be set.
type PolicySpec struct {
	Profile    string           `yaml:"profile,omitempty"    json:"profile,omitempty"`
	Extends    string           `yaml:"extends,omitempty"    json:"extends,omitempty"`
	Filesystem *FilesystemPolicy `yaml:"filesystem,omitempty" json:"filesystem,omitempty"`
	Network    *NetworkPolicy    `yaml:"network,omitempty"    json:"network,omitempty"`
	Hooks      *HooksPolicy      `yaml:"hooks,omitempty"      json:"hooks,omitempty"`
	Sandbox    *SandboxPolicy    `yaml:"sandbox,omitempty"    json:"sandbox,omitempty"`
	Audit      *AuditPolicy      `yaml:"audit,omitempty"      json:"audit,omitempty"`
}

// FilesystemPolicy controls filesystem access.
type FilesystemPolicy struct {
	Deny  []string `yaml:"deny,omitempty"  json:"deny,omitempty"`
	Allow []string `yaml:"allow,omitempty" json:"allow,omitempty"`
}

// NetworkPolicy controls network access.
type NetworkPolicy struct {
	Mode      string   `yaml:"mode"                json:"mode"`
	Allowlist []string `yaml:"allowlist,omitempty"  json:"allowlist,omitempty"`
}

// HooksPolicy defines pre/post tool-use hooks.
type HooksPolicy struct {
	PreToolUse  []HookDef `yaml:"pre_tool_use,omitempty"  json:"pre_tool_use,omitempty"`
	PostToolUse []HookDef `yaml:"post_tool_use,omitempty" json:"post_tool_use,omitempty"`
}

// HookDef is a single hook definition.
type HookDef struct {
	Matcher string `yaml:"matcher" json:"matcher"`
	Command string `yaml:"command" json:"command"`
}

// SandboxPolicy controls sandbox execution.
type SandboxPolicy struct {
	Enabled         bool     `yaml:"enabled"                    json:"enabled"`
	Runtime         string   `yaml:"runtime,omitempty"          json:"runtime,omitempty"`
	WritableMounts  []string `yaml:"writable_mounts,omitempty"  json:"writable_mounts,omitempty"`
	EnvPassthrough  []string `yaml:"env_passthrough,omitempty"  json:"env_passthrough,omitempty"`
	ImageBase       string   `yaml:"image_base,omitempty"       json:"image_base,omitempty"`
}

// AuditPolicy controls audit logging.
type AuditPolicy struct {
	Enabled      bool     `yaml:"enabled"                json:"enabled"`
	Path         string   `yaml:"path,omitempty"         json:"path,omitempty"`
	MaxSize      string   `yaml:"max_size,omitempty"     json:"max_size,omitempty"`
	IncludeTools []string `yaml:"include_tools,omitempty" json:"include_tools,omitempty"`
}
