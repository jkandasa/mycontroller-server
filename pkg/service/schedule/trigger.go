package schedule

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/mycontroller-org/server/v2/pkg/api/sunrise"
	"github.com/mycontroller-org/server/v2/pkg/store"
	types "github.com/mycontroller-org/server/v2/pkg/types"
	dateTimeTY "github.com/mycontroller-org/server/v2/pkg/types/cusom_datetime"
	scheduleTY "github.com/mycontroller-org/server/v2/pkg/types/schedule"
	"github.com/mycontroller-org/server/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/server/v2/pkg/utils/bus_utils"
	converterUtils "github.com/mycontroller-org/server/v2/pkg/utils/convertor"
	"github.com/mycontroller-org/server/v2/pkg/utils/javascript"
	variablesUtils "github.com/mycontroller-org/server/v2/pkg/utils/variables"
	"go.uber.org/zap"
)

func getScheduleTriggerFunc(cfg *scheduleTY.Config, spec string) func() {
	return func() { scheduleTriggerFunc(cfg, spec) }
}

func scheduleTriggerFunc(cfg *scheduleTY.Config, spec string) {
	// validate schedule
	if !isValidSchedule(cfg) {
		zap.L().Debug("at this time, this is not a valid schedule", zap.String("ScheduleID", cfg.ID), zap.String("spec", spec), zap.Any("validity details", cfg.Validity))
		return
	}

	start := time.Now()

	cfg.State.LastRun = time.Now()
	cfg.State.ExecutedCount++
	cfg.State.LastStatus = true
	cfg.State.Message = ""
	zap.L().Debug("triggered", zap.String("ID", cfg.ID), zap.String("spec", spec))

	executionError := ""

	// disable even there is a error on the schedule
	// call it inside func to get updated "executionError" value
	defer func() { verifyAndDisableSchedule(cfg, time.Since(start), executionError) }()

	// load variables
	variables, err := variablesUtils.LoadVariables(cfg.Variables, store.CFG.Secret)
	if err != nil {
		zap.L().Error("error on loading variables", zap.String("schedulerID", cfg.ID), zap.Error(err))
		// update triggered count and update state
		cfg.State.LastStatus = false
		cfg.State.Message = fmt.Sprintf("error: %s", err.Error())
		busUtils.SetScheduleState(cfg.ID, *cfg.State)
		executionError = err.Error()
		return
	}

	variables[types.KeySchedule] = cfg // include schedule in to the variables list

	switch cfg.CustomVariableType {
	case scheduleTY.CustomVariableTypeNone, "":
		// no action needed

	case scheduleTY.CustomVariableTypeJavascript:
		if cfg.CustomVariableConfig.Javascript != "" {
			result, err := javascript.Execute(cfg.CustomVariableConfig.Javascript, variables)
			if err != nil {
				zap.L().Error("error on executing javascript", zap.String("schedulerID", cfg.ID), zap.Error(err))
				cfg.State.LastStatus = false
				cfg.State.Message = fmt.Sprintf("error: %s", err.Error())
				busUtils.SetScheduleState(cfg.ID, *cfg.State)
				executionError = err.Error()
				return
			}

			// if the response is a map type merge it with variables
			if resultMap, ok := result.(map[string]interface{}); ok {
				variables = variablesUtils.Merge(variables, resultMap)
			}
		}

	case scheduleTY.CustomVariableTypeWebhook:
		customMap := loadWebhookVariables(cfg.ID, cfg.CustomVariableConfig, variables)
		if len(customMap) > 0 {
			variables = variablesUtils.Merge(variables, customMap)
		}

	default:
		zap.L().Error("unknown custom variable loader type", zap.String("type", cfg.CustomVariableType))
	}

	// post to handlers
	parameters := variablesUtils.UpdateParameters(variables, cfg.HandlerParameters)
	busUtils.PostToHandler(cfg.Handlers, parameters)

	cfg.State.Message = fmt.Sprintf("time taken: %s", time.Since(start).String())
	// update triggered count and update state
	busUtils.SetScheduleState(cfg.ID, *cfg.State)
}

func verifyAndDisableSchedule(cfg *scheduleTY.Config, timeTaken time.Duration, executionError string) {
	switch cfg.Type {

	// if repeat type job, verify repeat count
	case scheduleTY.TypeRepeat:
		spec := &scheduleTY.SpecRepeat{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			zap.L().Error("error on convert to repeat spec", zap.Error(err), zap.String("ScheduleID", cfg.ID), zap.Any("spec", cfg.Spec))
			return
		}
		if spec.RepeatCount != 0 && cfg.State.ExecutedCount >= spec.RepeatCount {
			zap.L().Debug("reached the maximum execution count, disabling this job", zap.String("ScheduleID", cfg.ID), zap.Any("spec", cfg.Spec))
			busUtils.DisableSchedule(cfg.ID)
			// Sometimes setState updates as enabled
			// To avoid this adding small sleep, but this is not good fix.
			utils.SmartSleep(200 * time.Millisecond)
			cfg.State.Message = fmt.Sprintf("time taken: %s, reached maximum execution count", timeTaken.String())
			if executionError != "" {
				cfg.State.Message = fmt.Sprintf("%s, executionError:%s", cfg.State.Message, executionError)
			}
			busUtils.SetScheduleState(cfg.ID, *cfg.State)
			return
		}

	// disable the schedule if it is a on date job
	case scheduleTY.TypeSimple, scheduleTY.TypeSunrise, scheduleTY.TypeSunset:
		spec := &scheduleTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			zap.L().Error("error on loading spec", zap.String("schedulerID", cfg.ID), zap.Error(err))
			cfg.State.LastStatus = false
			cfg.State.Message = fmt.Sprintf("error: %s", err.Error())
			if executionError != "" {
				cfg.State.Message = fmt.Sprintf("%s, executionError:%s", cfg.State.Message, executionError)
			}
			busUtils.SetScheduleState(cfg.ID, *cfg.State)
			return
		}
		if spec.Frequency == scheduleTY.FrequencyOnDate {
			busUtils.DisableSchedule(cfg.ID)
		}
	}

}

func getCronSpec(cfg *scheduleTY.Config) (string, error) {
	switch cfg.Type {
	case scheduleTY.TypeRepeat:
		spec := &scheduleTY.SpecRepeat{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("@every %s", spec.Interval), nil

	case scheduleTY.TypeCron:
		spec := &scheduleTY.SpecCron{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return spec.CronExpression, nil

	case scheduleTY.TypeSimple:
		spec := &scheduleTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}
		return toCronExpression(spec)

	case scheduleTY.TypeSunrise, scheduleTY.TypeSunset:
		spec := &scheduleTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return "", err
		}

		var suntime *time.Time

		if cfg.Type == scheduleTY.TypeSunrise {
			sunrise, err := sunrise.GetSunriseTime()
			if err != nil {
				return "", err
			}
			suntime = sunrise
		}
		if cfg.Type == scheduleTY.TypeSunset {
			sunset, err := sunrise.GetSunsetTime()
			if err != nil {
				return "", err
			}
			suntime = sunset
		}
		offset, err := time.ParseDuration(spec.Offset)
		if err != nil {
			return "", err
		}

		updatedTime := suntime.Add(offset)
		spec.Time = updatedTime.Format("15:04:05")
		return toCronExpression(spec)

	default:
		return "", fmt.Errorf("invalid schedule type: %s", cfg.Type)
	}
}

func toCronExpression(spec *scheduleTY.SpecSimple) (string, error) {
	cronRaw := struct {
		Seconds    string
		Minutes    string
		Hours      string
		DayOfMonth string
		Month      string
		DayOfWeek  string
	}{}

	cronRaw.Month = "*"

	switch spec.Frequency {
	case scheduleTY.FrequencyDaily, scheduleTY.FrequencyWeekly:
		cronRaw.DayOfMonth = "*"
		cronRaw.DayOfWeek = spec.DayOfWeek

	case scheduleTY.FrequencyMonthly:
		cronRaw.DayOfWeek = "*"
		cronRaw.DayOfMonth = converterUtils.ToString(spec.DateOfMonth)

	case scheduleTY.FrequencyOnDate:
		date, err := time.Parse(dateTimeTY.CustomDateFormat, spec.Date)
		if err != nil {
			return "", err
		}
		cronRaw.DayOfMonth = converterUtils.ToString(date.Day())
		cronRaw.Month = converterUtils.ToString(int(date.Month()))
		cronRaw.DayOfWeek = "*"

	default:
		return "", fmt.Errorf("invalid frequency: %s", spec.Frequency)
	}

	time := strings.Split(strings.TrimSpace(spec.Time), ":")
	if len(time) > 3 {
		return "", fmt.Errorf("invalid time: %s", spec.Time)
	}

	// update hour and minute
	cronRaw.Hours = time[0]
	cronRaw.Minutes = time[1]

	// update seconds
	if len(time) == 3 {
		cronRaw.Seconds = time[2]
	} else {
		cronRaw.Seconds = "0"
	}

	// format: "Seconds Minutes Hours DayOfMonth Month DayOfWeek"
	cron := fmt.Sprintf("%s %s %s %s %s %s", cronRaw.Seconds, cronRaw.Minutes, cronRaw.Hours, cronRaw.DayOfMonth, cronRaw.Month, cronRaw.DayOfWeek)
	return cron, nil
}

func isValidSchedule(cfg *scheduleTY.Config) bool {
	if !cfg.Validity.Enabled {
		return true
	}

	fromDate := time.Time(cfg.Validity.Date.From.Time)
	toDate := time.Time(cfg.Validity.Date.To.Time)
	fromTime := time.Time(cfg.Validity.Time.From.Time)
	toTime := time.Time(cfg.Validity.Time.To.Time)

	now := time.Now()

	// update from date with time
	if !fromDate.IsZero() {
		if fromTime.IsZero() { // set time to start of the day
			fromTime = time.Date(fromTime.Year(), fromTime.Month(), fromTime.Day(),
				0, 0, 0, 0, now.Location())
		} else { // set the time from defined data
			fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(),
				fromTime.Hour(), fromTime.Minute(), fromTime.Second(), fromTime.Nanosecond(),
				now.Location())
		}

		// update timezone to system timezone
		fromDate = time.Date(fromDate.Year(), fromDate.Month(), fromDate.Day(),
			fromDate.Hour(), fromDate.Minute(), fromDate.Second(), fromDate.Nanosecond(),
			now.Location())
	}

	// update to date with time
	if !toDate.IsZero() {
		if toTime.IsZero() { // set the time to end of the day
			toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
				23, 59, 59, 999999999, now.Location())
		} else { // set the time from defined data
			toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
				toTime.Hour(), toTime.Minute(), toTime.Second(), toTime.Nanosecond(),
				now.Location())
		}

		// update timezone to system timezone
		toDate = time.Date(toDate.Year(), toDate.Month(), toDate.Day(),
			toDate.Hour(), toDate.Minute(), toDate.Second(), toDate.Nanosecond(),
			now.Location())
	}

	// validate from date and time
	if !fromDate.IsZero() && now.Before(fromDate) {
		zap.L().Debug("failed", zap.Any("fromDate", fromDate), zap.Any("now", now))
		return false
	}

	// validate to date and time
	if !toDate.IsZero() && now.After(toDate) {
		zap.L().Debug("failed", zap.Any("toDate", toDate), zap.Any("now", now))
		return false
	}

	// if every date time validation enabled
	if cfg.Validity.ValidateTimeEveryday && (!fromTime.IsZero() || !toTime.IsZero()) {
		timeFormat := "150405"
		nowTimeInt, _ := strconv.ParseUint(now.Format(timeFormat), 10, 64)

		// validate from time
		if !fromTime.IsZero() {
			fromTimeInt, _ := strconv.ParseUint(fromTime.Format(timeFormat), 10, 64)
			if nowTimeInt < fromTimeInt {
				zap.L().Debug("failed", zap.Any("fromTime", fromTime))
				return false
			}
		}

		// validate to time
		if !toTime.IsZero() {
			toTimeInt, _ := strconv.ParseUint(toTime.Format(timeFormat), 10, 64)
			if nowTimeInt > toTimeInt {
				zap.L().Debug("failed", zap.Any("toTime", toTime))
				return false
			}
		}
	}
	return true
}

func updateOnDateJobValidity(cfg *scheduleTY.Config) error {
	if cfg.Type == scheduleTY.TypeSimple ||
		cfg.Type == scheduleTY.TypeSunrise ||
		cfg.Type == scheduleTY.TypeSunset {

		spec := &scheduleTY.SpecSimple{}
		err := utils.MapToStruct(utils.TagNameNone, cfg.Spec, spec)
		if err != nil {
			return err
		}
		if spec.Frequency != scheduleTY.FrequencyOnDate {
			return nil
		}
		date, err := time.Parse(dateTimeTY.CustomDateFormat, spec.Date)
		if err != nil {
			return nil
		}
		cfg.Validity.Enabled = true
		cfg.Validity.Date.From = dateTimeTY.CustomDate{Time: date}
		cfg.Validity.Date.To = dateTimeTY.CustomDate{Time: date}
	}
	return nil
}
