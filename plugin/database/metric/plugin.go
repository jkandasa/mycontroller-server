package metric

import (
	"fmt"

	"github.com/mycontroller-org/server/v2/pkg/model/cmap"
	metricType "github.com/mycontroller-org/server/v2/plugin/database/metric/type"
)

// CreatorFn func type
type CreatorFn func(config cmap.CustomMap) (metricType.Plugin, error)

// Creators is used for create plugins.
var creators = make(map[string]CreatorFn)

func Register(name string, fn CreatorFn) {
	creators[name] = fn
}

func Create(name string, config cmap.CustomMap) (p metricType.Plugin, err error) {
	if fn, ok := creators[name]; ok {
		p, err = fn(config)
	} else {
		err = fmt.Errorf("metric database plugin [%s] is not registered", name)
	}
	return
}
