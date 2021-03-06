package msgqueue

import (
	"encoding/json"
	"fmt"

	"github.com/WilfredDube/fxtract-backend/lib/contracts"
	"github.com/mitchellh/mapstructure"
)

type StaticEventMapper struct{}

func (e *StaticEventMapper) MapEvent(eventName string, serialized interface{}) (Event, error) {
	var event Event

	switch eventName {
	case "featureRecognitionStarted":
		event = &contracts.FeatureRecognitionStarted{}
	case "featureRecognitionComplete":
		event = &contracts.FeatureRecognitionComplete{}
	case "processPlanningStarted":
		event = &contracts.ProcessPlanningStarted{}
	case "processPlanningComplete":
		event = &contracts.ProcessPlanningComplete{}
	default:
		return nil, fmt.Errorf("unknown event type %s", eventName)
	}

	switch s := serialized.(type) {
	case []byte:
		err := json.Unmarshal(s, event)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal event %s: %s", eventName, err)
		}
	default:
		cfg := mapstructure.DecoderConfig{
			Result:  event,
			TagName: "json",
		}
		dec, err := mapstructure.NewDecoder(&cfg)
		if err != nil {
			return nil, fmt.Errorf("could not initialize decoder for event %s: %s", eventName, err)
		}

		err = dec.Decode(s)
		if err != nil {
			return nil, fmt.Errorf("could not unmarshal event %s: %s", eventName, err)
		}
	}

	return event, nil
}
