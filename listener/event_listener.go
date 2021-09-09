package listener

import (
	"fmt"
	"log"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/contracts"
	"github.com/WilfredDube/fxtract-backend/lib/msgqueue"
	"github.com/WilfredDube/fxtract-backend/service"
)

type EventProcessor struct {
	EventListener  msgqueue.EventListener
	CadFileService service.CadFileService
	TaskService    service.TaskService
}

func (p *EventProcessor) ProcessEvents() {
	log.Println("listening for events")

	received, errors, err := p.EventListener.Listen("featureRecognitionComplete")

	if err != nil {
		panic(err)
	}

	for {
		select {
		case evt := <-received:
			fmt.Printf("got event %T: \n", evt)
			p.handleEvent(evt)
		case err = <-errors:
			fmt.Printf("got error while receiving event: %s\n", err)
		}
	}
}

func (p *EventProcessor) handleEvent(event msgqueue.Event) {
	switch e := event.(type) {
	case *contracts.FeatureRecognitionComplete:
		log.Printf("event %s created: %s", e.CADFileID, e.TaskID)

		cadFile, err := p.CadFileService.Find(e.CADFileID)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to unmarshal data: ", err)
		}

		cadFile.FeatureProps = e.FeatureProps
		cadFile.BendFeatures = []entity.BendFeature{}
		cadFile.BendFeatures = e.BendFeatures

		_, err = p.CadFileService.Update(*cadFile)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to update data: ", err)
		}

		task, err := p.TaskService.Find(e.TaskID)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to retrieve task data: ", err)
		}

		task.Status = entity.Complete
		task.ProcessingTime = e.FeatureProps.FRETime

		returedTask, err := p.TaskService.Update(*task)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to update data: ", err)
		}

		log.Printf("[ User: %s > TaskID: %s > Task status: %s]: CAD file (%s) features saved successfully!", e.UserID, returedTask.TaskID, returedTask.Status, e.CADFileID)
		log.Printf("==========================================================")
	default:
		log.Printf("unknown event type: %T", e)
	}
}
