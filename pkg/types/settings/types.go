package settings

import (
	"time"

	"github.com/mycontroller-org/server/v2/pkg/types/cmap"
)

// key to get a specific settings
const (
	KeySystemSettings        = "system_settings"
	KeySystemJobs            = "system_jobs"
	KeySystemBackupLocations = "system_backup_locations"
	KeyVersion               = "version"
	KeyAnalytics             = "analytics"
	KeySystemDynamicSecrets  = "system_dynamic_secrets"
)

// Settings struct
type Settings struct {
	ID         string                 `json:"id" yaml:"id"`
	Spec       map[string]interface{} `json:"spec" yaml:"spec"`
	ModifiedOn time.Time              `json:"modifiedOn" yaml:"modifiedOn"`
}

// SystemSettings struct
type SystemSettings struct {
	GeoLocation  GeoLocation  `json:"geoLocation" yaml:"geoLocation"`
	Login        Login        `json:"login" yaml:"login"`
	Language     string       `json:"language" yaml:"language"`
	NodeStateJob NodeStateJob `json:"nodeStateJob" yaml:"nodeStateJob"`
}

// GeoLocation struct
type GeoLocation struct {
	AutoUpdate   bool    `json:"autoUpdate" yaml:"autoUpdate"`
	LocationName string  `json:"locationName" yaml:"locationName"`
	Latitude     float64 `json:"latitude" yaml:"latitude"`
	Longitude    float64 `json:"longitude" yaml:"longitude"`
}

// Login settings
type Login struct {
	Message       string `json:"message" yaml:"message"`
	ServerMessage string `json:"serverMessage" yaml:"serverMessage"`
}

// NodeStateJob verifies active node
type NodeStateJob struct {
	ExecutionInterval string `json:"executionInterval" yaml:"executionInterval"`
	InactiveDuration  string `json:"inactiveDuration" yaml:"inactiveDuration"`
}

// VersionSettings struct
type VersionSettings struct {
	Version string `json:"version" yaml:"version"`
}

// SystemJobsSettings cron struct
type SystemJobsSettings struct {
	Sunrise string `json:"sunrise" yaml:"sunrise"` // updates scheduled sunrise sunset jobs on this time
}

// BackupLocations of server
type BackupLocations struct {
	Locations []BackupLocation `json:"locations" yaml:"locations"`
}

// BackupLocation of server
type BackupLocation struct {
	Name   string         `json:"name" yaml:"name"`
	Type   string         `json:"type" yaml:"type"`
	Config cmap.CustomMap `json:"config" yaml:"config"`
}

// AnalyticsConfig keeps data
type AnalyticsConfig struct {
	AnonymousID string `json:"anonymousId" yaml:"anonymousId"`
}

// dynamic secrets used across system
type SystemDynamicSecrets struct {
	JwtAccessSecret string `json:"jwtAccessSecret" yaml:"jwtAccessSecret"`
}
