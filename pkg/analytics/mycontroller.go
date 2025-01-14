package analytics

import (
	"net/http"
	"time"

	gatewayAPI "github.com/mycontroller-org/server/v2/pkg/api/gateway"
	handlerAPI "github.com/mycontroller-org/server/v2/pkg/api/handler"
	settingsAPI "github.com/mycontroller-org/server/v2/pkg/api/settings"
	statusAPI "github.com/mycontroller-org/server/v2/pkg/api/status"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	httpClient "github.com/mycontroller-org/server/v2/pkg/utils/http_client_json"
	"github.com/mycontroller-org/server/v2/pkg/version"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	gatewayTY "github.com/mycontroller-org/server/v2/plugin/gateway/types"
	handlerTY "github.com/mycontroller-org/server/v2/plugin/handler/types"
	"go.uber.org/zap"
)

const (
	ANALYTICS_ID  = "MC198501010915"
	ANALYTICS_URL = "https://analytics.mycontroller.org/event"
	API_VERSION   = "1"
	timeout       = time.Second * 10
)

// ReportAnalyticsData to the analytics server
func ReportAnalyticsData() {
	zap.L().Debug("collecting analytics data")

	// create and update version details
	ver := version.Get()
	// update the anonymous id
	analytics, err := settingsAPI.GetAnalytics()
	if err != nil {
		zap.L().Debug("error on getting analytics details", zap.Error(err))
		return // if we can't get the anonymous id return from here
	}

	payload := Payload{
		APIVersion:  API_VERSION,
		AnalyticsID: ANALYTICS_ID,
		UserID:      analytics.AnonymousID,
		Application: Application{
			Version:   ver.Version,
			GitCommit: ver.GitCommit,
			BuildDate: ver.BuildDate,
			Platform:  ver.Platform,
			Arch:      ver.Arch,
			GoLang:    ver.GoVersion,
			RunningIn: statusAPI.RunningIn(),
			Uptime:    statusAPI.Get().Uptime,
			Gateways:  []string{},
			Handlers:  []string{},
		},
	}

	// include city, region and country details
	location, err := settingsAPI.GetLocation()
	if err != nil {
		zap.L().Debug("error on getting location details", zap.Error(err))
	} else {
		payload.Location = Location{
			City:    location.City,
			Region:  location.Region,
			Country: location.Country,
		}
	}

	enabledFilter := []storageTY.Filter{{Key: types.KeyEnabled, Operator: storageTY.OperatorEqual, Value: "true"}}
	pagination := &storageTY.Pagination{Limit: 100, Offset: 0} // gets only the first 100

	// update gateways type in use
	result, err := gatewayAPI.List(enabledFilter, pagination)
	if err != nil {
		zap.L().Error("error on getting gateway details", zap.Error(err))
	} else if result.Count > 0 {
		if data, ok := result.Data.(*[]gatewayTY.Config); ok {
			gateways := make([]string, 0)
			for _, gw := range *data {
				gwType := gw.Provider.GetString("type")
				if !utils.ContainsString(gateways, gwType) {
					gateways = append(gateways, gwType)
				}
			}
			payload.Application.Gateways = gateways
		}
	}

	// update handlers type in use
	result, err = handlerAPI.List(enabledFilter, pagination)
	zap.L().Debug("handler", zap.Any("result", result))
	if err != nil {
		zap.L().Error("error on getting handler details", zap.Error(err))
	} else if result.Count > 0 {
		if data, ok := result.Data.(*[]handlerTY.Config); ok {
			handlers := make([]string, 0)
			for _, handler := range *data {
				handlerType := handler.Type
				if !utils.ContainsString(handlers, handlerType) {
					handlers = append(handlers, handlerType)
				}
			}
			payload.Application.Handlers = handlers
		}
	}

	zap.L().Debug("analytics data to be reported", zap.Any("data", payload))

	// publish the data
	client := httpClient.GetClient(false, timeout)
	response, err := client.ExecuteJson(ANALYTICS_URL, http.MethodPost, nil, nil, payload, http.StatusOK)
	if err != nil {
		if response != nil {
			zap.L().Debug("error on sending analytics data", zap.String("response", response.StringBody()), zap.Any("responseConfig", response), zap.Error(err))
		} else {
			zap.L().Debug("error on sending analytics data", zap.Error(err))
		}
	}
}
