package settings

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/json"
	"github.com/mycontroller-org/server/v2/pkg/service/configuration"
	systemJobsHelper "github.com/mycontroller-org/server/v2/pkg/service/system_jobs/helper_utils"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	settingsTY "github.com/mycontroller-org/server/v2/pkg/types/settings"
	webHandlerTY "github.com/mycontroller-org/server/v2/pkg/types/web_handler"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	cloneUtil "github.com/mycontroller-org/server/v2/pkg/utils/clone"
	storageTY "github.com/mycontroller-org/server/v2/plugin/database/storage/types"
	"go.uber.org/zap"
)

// List by filter and pagination
func List(filters []storageTY.Filter, pagination *storageTY.Pagination) (*storageTY.Result, error) {
	result := make([]settingsTY.Settings, 0)
	return store.STORAGE.Find(types.EntitySettings, &result, filters, pagination)
}

// Save a setting details
func Save(settings *settingsTY.Settings) error {
	if settings.ID == "" {
		return errors.New("id should not be nil")
	}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: settings.ID},
	}
	return store.STORAGE.Upsert(types.EntitySettings, settings, filters)
}

// GetByID returns a item
func GetByID(ID string) (*settingsTY.Settings, error) {
	result := &settingsTY.Settings{}
	filters := []storageTY.Filter{
		{Key: types.KeyID, Operator: storageTY.OperatorEqual, Value: ID},
	}
	err := store.STORAGE.FindOne(types.EntitySettings, result, filters)
	if err != nil {
		return nil, err
	}

	// convert map data to struct
	var specStruct interface{}
	switch result.ID {
	case settingsTY.KeySystemSettings:
		settings := settingsTY.SystemSettings{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	case settingsTY.KeySystemJobs:
		settings := settingsTY.SystemJobsSettings{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	case settingsTY.KeySystemBackupLocations:
		exportLocations := settingsTY.BackupLocations{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &exportLocations)
		if err != nil {
			return nil, err
		}
		specStruct = exportLocations

	case settingsTY.KeySystemDynamicSecrets:
		settings := settingsTY.SystemDynamicSecrets{}
		err = utils.MapToStruct(utils.TagNameNone, result.Spec, &settings)
		if err != nil {
			return nil, err
		}
		specStruct = settings

	default:

	}

	// struct to json then json to map
	// this workaround to apply json tags
	if specStruct != nil {
		bytes, err := json.Marshal(specStruct)
		if err != nil {
			return nil, err
		}

		mapSpec := make(map[string]interface{})
		err = json.Unmarshal(bytes, &mapSpec)
		if err != nil {
			return nil, err
		}
		result.Spec = mapSpec
	}

	return result, nil
}

// UpdateSettings config into disk
func UpdateSettings(settings *settingsTY.Settings) error {
	if settings.ID == "" {
		return errors.New("id cannot be empty")
	}

	switch settings.ID {
	case settingsTY.KeySystemSettings:
		return UpdateSystemSettings(settings)

	case settingsTY.KeySystemJobs,
		settingsTY.KeyVersion,
		settingsTY.KeySystemBackupLocations,
		settingsTY.KeyAnalytics,
		settingsTY.KeySystemDynamicSecrets:
		if !configuration.PauseModifiedOnUpdate.IsSet() {
			settings.ModifiedOn = time.Now()
		}
		return update(settings)

	default:
		return fmt.Errorf("unknown settings id:%s", settings.ID)
	}

}

// GetSystemJobs details
func GetSystemJobs() (*settingsTY.SystemJobsSettings, error) {
	settings, err := GetByID(settingsTY.KeySystemJobs)
	if err != nil {
		return nil, err
	}
	systemJobs := &settingsTY.SystemJobsSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemJobs)
	if err != nil {
		return nil, err
	}
	return systemJobs, nil
}

// UpdateSystemSettings config into disk
func UpdateSystemSettings(settings *settingsTY.Settings) error {
	settings.ID = settingsTY.KeySystemSettings
	if !configuration.PauseModifiedOnUpdate.IsSet() {
		settings.ModifiedOn = time.Now()
	}

	// TODO: verify required fields
	err := update(settings)
	if err != nil {
		return err
	}
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return err
	}
	if systemSettings.GeoLocation.AutoUpdate {
		err = AutoUpdateSystemGEOLocation()
		if err != nil {
			return err
		}
	}
	// send geo location updated event

	// post node state updater job change event
	systemJobsHelper.PostEvent(systemJobsHelper.JobTypeNodeStateUpdater)

	return nil
}

// UpdateGeoLocation updates the location details
func UpdateGeoLocation(location *settingsTY.GeoLocation) error {
	settings, err := GetByID(settingsTY.KeySystemSettings)
	if err != nil {
		return err
	}

	// convert spec to system settings
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return err
	}

	// update location
	systemSettings.GeoLocation = *location
	settings.Spec = utils.StructToMap(systemSettings)
	err = update(settings)
	if err != nil {
		return err
	}

	// send geo location updated event
	return nil
}

// GetGeoLocation returns configured latitude and longitude settings to calculate sunrise and sunset
func GetGeoLocation() (*settingsTY.GeoLocation, error) {
	settings, err := GetByID(settingsTY.KeySystemSettings)
	if err != nil {
		return nil, err
	}

	// convert spec to system settings
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	sunrise := systemSettings.GeoLocation
	return &sunrise, nil
}

// GetBackupLocations returns locations set by user
func GetBackupLocations() (*settingsTY.BackupLocations, error) {
	settings, err := GetByID(settingsTY.KeySystemBackupLocations)
	if err != nil {
		return nil, err
	}

	// convert spec to BackupLocations
	systemSettings := &settingsTY.BackupLocations{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	return systemSettings, nil
}

// update is a common function to update a document in settings entity
func update(settings *settingsTY.Settings) error {
	filters := []storageTY.Filter{
		{Key: types.KeyID, Value: settings.ID},
	}
	if !configuration.PauseModifiedOnUpdate.IsSet() {
		settings.ModifiedOn = time.Now()
	}

	// encrypt passwords, tokens, etc
	err := cloneUtil.UpdateSecrets(settings, store.CFG.Secret, "", true, cloneUtil.DefaultSpecialKeys)
	if err != nil {
		zap.L().Error("error on encryption", zap.Error(err))
		return err
	}

	return store.STORAGE.Upsert(types.EntitySettings, settings, filters)
}

// GetAnalytics returns analytics data
func GetAnalytics() (*settingsTY.AnalyticsConfig, error) {
	settings, err := GetByID(settingsTY.KeyAnalytics)
	if err != nil {
		return nil, err
	}

	// convert spec to analytics data
	analyticsData := &settingsTY.AnalyticsConfig{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, analyticsData)
	if err != nil {
		return nil, err
	}

	return analyticsData, nil
}

// GetSystemSettings returns system settings data
func GetSystemSettings() (*settingsTY.SystemSettings, error) {
	settings, err := GetByID(settingsTY.KeySystemSettings)
	if err != nil {
		return nil, err
	}

	// convert spec to analytics data
	systemSettings := &settingsTY.SystemSettings{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSettings)
	if err != nil {
		return nil, err
	}

	return systemSettings, nil
}

func ResetJwtAccessSecret(newSecret string) error {
	if newSecret == "" {
		newSecret = utils.RandUUID()
	}

	systemSecrets := &settingsTY.SystemDynamicSecrets{
		JwtAccessSecret: newSecret,
	}
	spec := utils.StructToMap(systemSecrets)

	settings := &settingsTY.Settings{
		ID:   settingsTY.KeySystemDynamicSecrets,
		Spec: spec,
	}
	err := UpdateSettings(settings)
	if err != nil {
		return err
	}

	return UpdateJwtAccessSecret()
}

func UpdateJwtAccessSecret() error {
	settings, err := GetByID(settingsTY.KeySystemDynamicSecrets)
	if err != nil {
		if err != storageTY.ErrNoDocuments {
			return err
		}
		settings = &settingsTY.Settings{}
	}

	// decrypt passwords, tokens, etc
	err = cloneUtil.UpdateSecrets(settings, store.CFG.Secret, "", false, cloneUtil.DefaultSpecialKeys)
	if err != nil {
		zap.L().Error("error on decryption", zap.Error(err))
		return err
	}

	systemSecret := &settingsTY.SystemDynamicSecrets{}
	err = utils.MapToStruct(utils.TagNameNone, settings.Spec, systemSecret)
	if err != nil {
		return err
	}

	if systemSecret.JwtAccessSecret == "" {
		systemSecret.JwtAccessSecret = utils.RandUUID()
		err = ResetJwtAccessSecret(systemSecret.JwtAccessSecret)
		if err != nil {
			return err
		}
	}

	return os.Setenv(webHandlerTY.EnvJwtAccessSecret, systemSecret.JwtAccessSecret)
}
