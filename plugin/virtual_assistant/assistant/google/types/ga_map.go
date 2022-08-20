package types

import (
	vdTY "github.com/mycontroller-org/server/v2/pkg/types/virtual_device"
)

var (
	// https://developers.google.com/assistant/smarthome/guides
	DeviceMap = map[string]string{
		vdTY.DeviceTypeAirConditioner:         "action.devices.types.AC_UNIT",
		vdTY.DeviceTypeAirCooler:              "action.devices.types.AIRCOOLER",
		vdTY.DeviceTypeAirFreshener:           "action.devices.types.AIRFRESHENER",
		vdTY.DeviceTypeAirPurifier:            "action.devices.types.AIRPURIFIER",
		vdTY.DeviceTypeAudioVideoReceiver:     "action.devices.types.AUDIO_VIDEO_RECEIVER",
		vdTY.DeviceTypeAwning:                 "action.devices.types.AWNING",
		vdTY.DeviceTypeBathtub:                "action.devices.types.BATHTUB",
		vdTY.DeviceTypeBed:                    "action.devices.types.BED",
		vdTY.DeviceTypeBlinds:                 "action.devices.types.BLENDER",
		vdTY.DeviceTypeBlender:                "action.devices.types.BLINDS",
		vdTY.DeviceTypeBoiler:                 "action.devices.types.BOILER",
		vdTY.DeviceTypeCamera:                 "action.devices.types.CAMERA",
		vdTY.DeviceTypeCarbonMonoxideDetector: "action.devices.types.CARBON_MONOXIDE_DETECTOR",
		vdTY.DeviceTypeCharger:                "action.devices.types.CHARGER",
		vdTY.DeviceTypeCloset:                 "action.devices.types.CLOSET",
		vdTY.DeviceTypeCoffeeMaker:            "action.devices.types.COFFEE_MAKER",
		vdTY.DeviceTypeCooktop:                "action.devices.types.COOKTOP",
		vdTY.DeviceTypeCurtain:                "action.devices.types.CURTAIN",
		vdTY.DeviceTypeDehumidifier:           "action.devices.types.DEHUMIDIFIER",
		vdTY.DeviceTypeDehydrator:             "action.devices.types.DEHYDRATOR",
		vdTY.DeviceTypeDishwasher:             "action.devices.types.DISHWASHER",
		vdTY.DeviceTypeDoor:                   "action.devices.types.DOOR",
		vdTY.DeviceTypeDoorBell:               "action.devices.types.DOORBELL",
		vdTY.DeviceTypeDrawer:                 "action.devices.types.DRAWER",
		vdTY.DeviceTypeDryer:                  "action.devices.types.DRYER",
		vdTY.DeviceTypeFan:                    "action.devices.types.FAN",
		vdTY.DeviceTypeFaucet:                 "action.devices.types.FAUCET",
		vdTY.DeviceTypeFireplace:              "action.devices.types.FIREPLACE",
		vdTY.DeviceTypeFreezer:                "action.devices.types.FREEZER",
		vdTY.DeviceTypeFryer:                  "action.devices.types.FRYER",
		vdTY.DeviceTypeGarageDoor:             "action.devices.types.GARAGE",
		vdTY.DeviceTypeGate:                   "action.devices.types.GATE",
		vdTY.DeviceTypeGrill:                  "action.devices.types.GRILL",
		vdTY.DeviceTypeHeater:                 "action.devices.types.HEATER",
		vdTY.DeviceTypeHood:                   "action.devices.types.HOOD",
		vdTY.DeviceTypeHumidifier:             "action.devices.types.HUMIDIFIER",
		vdTY.DeviceTypeKettle:                 "action.devices.types.KETTLE",
		vdTY.DeviceTypeLight:                  "action.devices.types.LIGHT",
		vdTY.DeviceTypeLock:                   "action.devices.types.LOCK",
		vdTY.DeviceTypeMicrowave:              "action.devices.types.MICROWAVE",
		vdTY.DeviceTypeMop:                    "action.devices.types.MOP",
		vdTY.DeviceTypeMower:                  "action.devices.types.MOWER",
		vdTY.DeviceTypeMulticooker:            "action.devices.types.MULTICOOKER",
		vdTY.DeviceTypeNetwork:                "action.devices.types.NETWORK",
		vdTY.DeviceTypeOutlet:                 "action.devices.types.OUTLET",
		vdTY.DeviceTypeOven:                   "action.devices.types.OVEN",
		vdTY.DeviceTypePergola:                "action.devices.types.PERGOLA",
		vdTY.DeviceTypePetFeeder:              "action.devices.types.PETFEEDER",
		vdTY.DeviceTypePressureCooker:         "action.devices.types.PRESSURECOOKER",
		vdTY.DeviceTypeRadiator:               "action.devices.types.RADIATOR",
		vdTY.DeviceTypeRefrigerator:           "action.devices.types.REFRIGERATOR",
		vdTY.DeviceTypeRemoteControl:          "action.devices.types.REMOTECONTROL",
		vdTY.DeviceTypeRouter:                 "action.devices.types.ROUTER",
		vdTY.DeviceTypeScene:                  "action.devices.types.SCENE",
		vdTY.DeviceTypeSecuritySystem:         "action.devices.types.SECURITYSYSTEM",
		vdTY.DeviceTypeSensor:                 "action.devices.types.SENSOR",
		vdTY.DeviceTypeSetTopBox:              "action.devices.types.SETTOP",
		vdTY.DeviceTypeShower:                 "action.devices.types.SHOWER",
		vdTY.DeviceTypeShutter:                "action.devices.types.SHUTTER",
		vdTY.DeviceTypeSmokeDetector:          "action.devices.types.SMOKE_DETECTOR",
		vdTY.DeviceTypeSoundbar:               "action.devices.types.SOUNDBAR",
		vdTY.DeviceTypeSousVide:               "action.devices.types.SOUSVIDE",
		vdTY.DeviceTypeSpeaker:                "action.devices.types.SPEAKER",
		vdTY.DeviceTypeSprinkler:              "action.devices.types.SPRINKLER",
		vdTY.DeviceTypeStandMixer:             "action.devices.types.STANDMIXER",
		vdTY.DeviceTypeStreamingBox:           "action.devices.types.STREAMING_BOX",
		vdTY.DeviceTypeStreamingSoundbar:      "action.devices.types.STREAMING_SOUNDBAR",
		vdTY.DeviceTypeStreamingStick:         "action.devices.types.STREAMING_STICK",
		vdTY.DeviceTypeSwitch:                 "action.devices.types.SWITCH",
		vdTY.DeviceTypeThermostat:             "action.devices.types.THERMOSTAT",
		vdTY.DeviceTypeTelevision:             "action.devices.types.TV",
		vdTY.DeviceTypeVacuum:                 "action.devices.types.VACUUM",
		vdTY.DeviceTypeValve:                  "action.devices.types.VALVE",
		vdTY.DeviceTypeWasher:                 "action.devices.types.WASHER",
		vdTY.DeviceTypeWaterHeater:            "action.devices.types.WATERHEATER",
		vdTY.DeviceTypeWaterPurifier:          "action.devices.types.WATERPURIFIER",
		vdTY.DeviceTypeWaterSoftener:          "action.devices.types.WATERSOFTENER",
		vdTY.DeviceTypeWeatherStation:         "action.devices.types.SENSOR", // not available in google
		vdTY.DeviceTypeWindow:                 "action.devices.types.WINDOW",
		vdTY.DeviceTypeYogurtMaker:            "action.devices.types.YOGURTMAKER",
	}

	// https://developers.google.com/assistant/smarthome/traits
	TraitMap = map[string]string{
		vdTY.DeviceTraitAppSelector:        "action.devices.traits.AppSelector",
		vdTY.DeviceTraitArmDisarm:          "action.devices.traits.ArmDisarm",
		vdTY.DeviceTraitBrightness:         "action.devices.traits.Brightness",
		vdTY.DeviceTraitCameraStream:       "action.devices.traits.CameraStream",
		vdTY.DeviceTraitChannel:            "action.devices.traits.Channel",
		vdTY.DeviceTraitColorSetting:       "action.devices.traits.ColorSetting",
		vdTY.DeviceTraitCook:               "action.devices.traits.Cook",
		vdTY.DeviceTraitDispense:           "action.devices.traits.Dispense",
		vdTY.DeviceTraitDock:               "action.devices.traits.Dock",
		vdTY.DeviceTraitEnergyStorage:      "action.devices.traits.EnergyStorage",
		vdTY.DeviceTraitFanSpeed:           "action.devices.traits.FanSpeed",
		vdTY.DeviceTraitFill:               "action.devices.traits.Fill",
		vdTY.DeviceTraitHumiditySetting:    "action.devices.traits.HumiditySetting",
		vdTY.DeviceTraitInputSelector:      "action.devices.traits.InputSelector",
		vdTY.DeviceTraitLightEffects:       "action.devices.traits.LightEffects",
		vdTY.DeviceTraitLocator:            "action.devices.traits.Locator",
		vdTY.DeviceTraitLockUnlock:         "action.devices.traits.LockUnlock",
		vdTY.DeviceTraitMediaState:         "action.devices.traits.MediaState",
		vdTY.DeviceTraitModes:              "action.devices.traits.Modes",
		vdTY.DeviceTraitNetworkControl:     "action.devices.traits.NetworkControl",
		vdTY.DeviceTraitObjectDetection:    "action.devices.traits.ObjectDetection",
		vdTY.DeviceTraitOnOff:              "action.devices.traits.OnOff",
		vdTY.DeviceTraitOpenClose:          "action.devices.traits.OpenClose",
		vdTY.DeviceTraitReboot:             "action.devices.traits.Reboot",
		vdTY.DeviceTraitRotation:           "action.devices.traits.Rotation",
		vdTY.DeviceTraitRunCycle:           "action.devices.traits.RunCycle",
		vdTY.DeviceTraitSensorState:        "action.devices.traits.SensorState",
		vdTY.DeviceTraitScene:              "action.devices.traits.Scene",
		vdTY.DeviceTraitSoftwareUpdate:     "action.devices.traits.SoftwareUpdate",
		vdTY.DeviceTraitStartStop:          "action.devices.traits.StartStop",
		vdTY.DeviceTraitStatusReport:       "action.devices.traits.StatusReport",
		vdTY.DeviceTraitTemperatureControl: "action.devices.traits.TemperatureControl",
		vdTY.DeviceTraitTemperatureSetting: "action.devices.traits.TemperatureSetting",
		vdTY.DeviceTraitTimer:              "action.devices.traits.Timer",
		vdTY.DeviceTraitToggles:            "action.devices.traits.Toggles",
		vdTY.DeviceTraitTransportControl:   "action.devices.traits.TransportControl",
		vdTY.DeviceTraitVolume:             "action.devices.traits.Volume",
	}

	CommandParamsMap = map[string]string{
		"on":         vdTY.DeviceTraitOnOff,
		"brightness": vdTY.DeviceTraitBrightness,
	}
)