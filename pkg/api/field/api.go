package field

import (
	"github.com/mycontroller-org/backend/v2/pkg/model"
	ml "github.com/mycontroller-org/backend/v2/pkg/model"
	eventML "github.com/mycontroller-org/backend/v2/pkg/model/bus/event"
	fml "github.com/mycontroller-org/backend/v2/pkg/model/field"
	"github.com/mycontroller-org/backend/v2/pkg/service/mcbus"
	stg "github.com/mycontroller-org/backend/v2/pkg/service/storage"
	ut "github.com/mycontroller-org/backend/v2/pkg/utils"
	busUtils "github.com/mycontroller-org/backend/v2/pkg/utils/bus_utils"
	stgml "github.com/mycontroller-org/backend/v2/plugin/storage"
)

// List by filter and pagination
func List(filters []stgml.Filter, pagination *stgml.Pagination) (*stgml.Result, error) {
	result := make([]fml.Field, 0)
	return stg.SVC.Find(ml.EntityField, &result, filters, pagination)
}

// Get returns a field
func Get(filters []stgml.Filter) (*fml.Field, error) {
	result := &fml.Field{}
	err := stg.SVC.FindOne(ml.EntityField, result, filters)
	return result, err
}

// GetByID returns a field
func GetByID(id string) (*fml.Field, error) {
	filters := []stgml.Filter{
		{Key: model.KeyID, Value: id},
	}
	result := &fml.Field{}
	err := stg.SVC.FindOne(model.EntityFirmware, result, filters)
	return result, err
}

// Save a field details
func Save(field *fml.Field, retainValue bool) error {
	eventType := eventML.TypeUpdated
	if field.ID == "" {
		field.ID = ut.RandUUID()
		eventType = eventML.TypeCreated
	}
	filters := []stgml.Filter{
		{Key: ml.KeyID, Value: field.ID},
	}

	if retainValue && eventType != eventML.TypeCreated {
		fieldOrg, err := GetByID(field.ID)
		if err != nil {
			return err
		}
		field.Current = fieldOrg.Current
		field.Previous = fieldOrg.Previous
	}
	err := stg.SVC.Upsert(ml.EntityField, field, filters)
	if err != nil {
		return err
	}

	if retainValue { // assume this change from HTTP API
		busUtils.PostEvent(mcbus.TopicEventHandler, eventType, model.EntityHandler, field)
	}
	return nil
}

// GetByIDs returns a field details by gatewayID, nodeId, sourceID and fieldName of a message
func GetByIDs(gatewayID, nodeID, sourceID, fieldID string) (*fml.Field, error) {
	filters := []stgml.Filter{
		{Key: ml.KeyGatewayID, Value: gatewayID},
		{Key: ml.KeyNodeID, Value: nodeID},
		{Key: ml.KeySourceID, Value: sourceID},
		{Key: ml.KeyFieldID, Value: fieldID},
	}
	result := &fml.Field{}
	err := stg.SVC.FindOne(ml.EntityField, result, filters)
	return result, err
}

// Delete fields
func Delete(IDs []string) (int64, error) {
	filters := []stgml.Filter{{Key: ml.KeyID, Operator: stgml.OperatorIn, Value: IDs}}
	return stg.SVC.Delete(ml.EntityField, filters)
}
