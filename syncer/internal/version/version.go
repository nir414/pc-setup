package version

// Version information for the syncer tool
const (
Version = "1.0.0"
Name    = "syncer"
Description = "Windows 설정 백업 도구"
)

// GetVersionString returns the full version string
func GetVersionString() string {
return Name + " v" + Version
}
