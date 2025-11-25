package config

// Config represents the schema of sync.toml.
type Config struct {
	SyncData map[string]Section `toml:"SyncData"`
}

// Section describes folders belonging to an environment root.
type Section struct {
	Folders  []string `toml:"folders"`
	Excludes []string `toml:"excludes"`
}
