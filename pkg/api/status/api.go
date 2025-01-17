package status

import (
	"fmt"
	"os"
	"time"

	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	"go.uber.org/zap"
)

const (
	EnvironmentDocker     = "docker"
	EnvironmentKubernetes = "kubernetes"
	EnvironmentBareMetal  = "bare_metal"
)

var (
	startTime time.Time
)

func init() {
	startTime = time.Now()
}

type Status struct {
	Hostname          string           `json:"hostname"`
	DocumentationURL  string           `json:"documentationUrl"`
	Login             settingsTY.Login `json:"login"`
	StartTime         time.Time        `json:"startTime"`
	ServerTime        time.Time        `json:"serverTime"`
	Uptime            uint64           `json:"uptime"` // in seconds
	MetricsDBDisabled bool             `json:"metricsDBDisabled"`
	Language          string           `json:"language"`
}

func get(minimal bool) Status {
	status := Status{
		DocumentationURL: store.CFG.Web.DocumentationURL,
	}
	status.MetricsDBDisabled = store.CFG.Database.Metric.GetBool(types.KeyDisabled)

	if !minimal {
		hostname, err := os.Hostname()
		if err != nil {
			zap.L().Error("error on getting hostname", zap.Error(err))
			hostname = fmt.Sprintf("error:%s", err.Error())
		}

		status.Hostname = hostname
		status.ServerTime = time.Now()
		status.StartTime = startTime
		status.Uptime = uint64(time.Since(startTime).Seconds())
	}

	// include login message
	login := settingsTY.Login{}
	sysSettings, err := settingsAPI.GetSystemSettings()
	if err != nil {
		zap.L().Error("error on getting system settings", zap.Error(err))
		login.Message = fmt.Sprintf("error on getting login message: %s", err.Error())
	} else {
		login = sysSettings.Login
		status.Language = sysSettings.Language
	}
	status.Login = login

	return status
}

// Get returns status with all fields
func Get() Status {
	return get(false)
}

// GetMinimal returns limited fields, can be used under status rest api (login not required)
func GetMinimal() Status {
	return get(true)
}

// docker creates a .dockerenv file at the root of the directory tree inside the container.
// if this file exists then the viewer is running from inside a container so return true
// With the default configuration, Kubernetes will mount the serviceaccount secrets into pods.
func RunningIn() string {
	if utils.IsFileExists("/.dockerenv") {
		return EnvironmentDocker
	} else if utils.IsDirExists("/var/run/secrets/kubernetes.io") {
		return EnvironmentKubernetes
	}
	return EnvironmentBareMetal
}
