package listener

import (
	"fmt"
	"log"
	"time"

	"github.com/WilfredDube/fxtract-backend/controller"
	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/WilfredDube/fxtract-backend/lib/contracts"
	"github.com/WilfredDube/fxtract-backend/lib/msgqueue"
	persistence "github.com/WilfredDube/fxtract-backend/repository/reposelect"
	"github.com/WilfredDube/fxtract-backend/service"
	"github.com/teris-io/shortid"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EventProcessor struct {
	EventListener         msgqueue.EventListener
	CadFileService        service.CadFileService
	TaskService           service.TaskService
	ToolService           service.ToolService
	ProcessingPlanService service.ProcessingPlanService
	MaterialService       service.MaterialService
	ProjectService        service.ProjectService
	UserService           service.UserService
	Processor             *service.Processor
	pdfService            service.PDFService
}

func (p *EventProcessor) ProcessEvents(events ...string) {
	log.Println("listening for events")

	received, errors, err := p.EventListener.Listen(events...)

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

		for i, bend := range cadFile.BendFeatures {
			tool, err := p.ToolService.FindByAngle(int64(bend.Angle))
			if err != nil {
				log.Fatalf("%s: %s", "Failed to retrieve tool data: ", err)
			}

			cadFile.BendFeatures[i].ToolID = tool.ToolID
		}

		material, err := p.MaterialService.Find(cadFile.Material)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to retrieve material data: ", err)
		}

		maxLength := 0.0
		for _, file := range cadFile.BendFeatures {
			if file.Length > maxLength {
				maxLength = file.Length
			}
		}

		bendingForce := (maxLength * cadFile.FeatureProps.Thickness * material.KFactor * material.TensileStrength) / 8
		cadFile.FeatureProps.BendingForce = bendingForce

		_, err = p.CadFileService.Update(*cadFile)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to update data: ", err)
		}

		PROJECTCADFILES := controller.CADFILECACHE + cadFile.ProjectID.Hex()
		go persistence.ClearCache(cadFile.ProjectID.Hex())
		go persistence.ClearCache(PROJECTCADFILES)

		task, err := p.TaskService.Find(e.TaskID)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to retrieve task data: ", err)
		}

		task.ProcessingTime = e.FeatureProps.FRETime
		task.ProcessedCADFiles = append(task.ProcessedCADFiles, entity.Processed{ID: cadFile.ID, FileName: cadFile.FileName, ProcessType: entity.FeatureRecognition, Status: entity.Complete})

		if task.Quantity == (int64(len(task.ProcessedCADFiles))) {
			task.Status = entity.Complete
		}

		returedTask, err := p.TaskService.Update(task)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to update data: ", err)
		}

		log.Printf("[ User: %s > TaskID: %s > Task status: %s]: CAD file (%s) features saved successfully!", e.UserID, returedTask.TaskID, returedTask.Status, e.CADFileID)
		log.Printf("==========================================================")

		if returedTask.Status == entity.Complete {
			go func() {
				p.Processor.TaskChannel <- returedTask
			}()

			go func() {
				project, err := p.ProjectService.Find(cadFile.ProjectID.Hex())
				if err != nil {
					log.Fatalf("%s: %s", "Project does not exist ", err)
				}

				cadFiles, err := p.CadFileService.FindAll(project.ID.Hex())
				if err != nil {
					log.Fatalf("%s: %s", "Failed to retrieve cad files data: ", err)
				}

				p.Processor.CADFilesChannel <- service.CADFileResponse{UserID: e.UserID, CadFiles: cadFiles}
			}()
		}
	case *contracts.ProcessPlanningComplete:
		log.Printf("event %s created: %s", e.CADFileID, e.TaskID)
		log.Printf("==========================================================")
		fmt.Printf("Received a Processing plan for CAD file ID: %v\n", e.ProcessingPlan.CADFileID)

		p.pdfService = service.NewPDFService()
		processingPlan := entity.ProcessingPlan{}
		processingPlan.ID = primitive.NewObjectID()
		processingPlan.CADFileID = e.ProcessingPlan.CADFileID

		cadFile, err := p.CadFileService.Find(processingPlan.CADFileID.Hex())
		if err != nil {
			log.Fatalf("%s: %s", "Cadfile does not exist ", err)
		}

		project, err := p.ProjectService.Find(cadFile.ProjectID.Hex())
		if err != nil {
			log.Fatalf("%s: %s", "Project does not exist ", err)
		}

		user, err := p.UserService.Profile(project.OwnerID.Hex())
		if err != nil {
			log.Fatalf("%s: %s", "User does not exist ", err)
		}

		processingPlan.Rotations = e.ProcessingPlan.Rotations
		processingPlan.Flips = e.ProcessingPlan.Flips
		processingPlan.Tools = e.ProcessingPlan.Tools
		processingPlan.Modules = e.ProcessingPlan.Modules
		processingPlan.ProcessingTime = e.ProcessingPlan.ProcessingTime
		processingPlan.EstimatedManufacturingTime = e.ProcessingPlan.EstimatedManufacturingTime
		processingPlan.TotalToolDistance = e.ProcessingPlan.TotalToolDistance
		processingPlan.BendingSequences = e.ProcessingPlan.BendingSequences
		processingPlan.Quantity = e.ProcessingPlan.Quantity
		processingPlan.BendingForce = cadFile.FeatureProps.BendingForce
		processingPlan.FileName = cadFile.FileName
		processingPlan.Material = cadFile.Material
		processingPlan.ProjectTitle = project.Title
		processingPlan.Engineer = user.FullName()
		processingPlan.BendFeatures = cadFile.BendFeatures

		sid, err := shortid.New(1, shortid.DefaultABC, 2342)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to save processing plan: ", err)
		}

		processingPlan.PartNo = sid.MustGenerate()
		processingPlan.CreatedAt = time.Now().Unix()

		pdfBuff, err := p.pdfService.GeneratePDF(&processingPlan)
		if err != nil {
			log.Fatalf("%s: %s", "Failed: ", err)
		}

		pdfBlob := service.NewAzureBlobService()
		filename := fmt.Sprintf(project.ID.Hex()+"/%s.pdf", processingPlan.ID.Hex())
		_, url, err := pdfBlob.UploadFromBuffer(&pdfBuff, filename)
		if err != nil {
			log.Fatalf("%s: %s", "Failed: ", err)
		}

		processingPlan.PdfURL = url

		_, err = p.ProcessingPlanService.Create(&processingPlan)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to save processing plan: ", err)
		}

		cadFile.FeatureProps.ProcessLevel = e.ProcessLevel
		_, err = p.CadFileService.Update(*cadFile)
		if err != nil {
			log.Fatalf("%s: %s", "Cadfile update failed ", err)
		}

		PROJECTCADFILES := controller.CADFILECACHE + cadFile.ProjectID.Hex()
		go persistence.ClearCache(cadFile.ProjectID.Hex())
		go persistence.ClearCache(processingPlan.ID.Hex())
		go persistence.ClearCache(PROJECTCADFILES)

		task, err := p.TaskService.Find(e.TaskID)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to retrieve task data: ", err)
		}

		task.ProcessingTime = e.ProcessingPlan.EstimatedManufacturingTime
		task.ProcessedCADFiles = append(task.ProcessedCADFiles, entity.Processed{ID: cadFile.ID, FileName: cadFile.FileName, ProcessType: entity.ProcessPlanning, Status: entity.Complete})

		if task.Quantity == (int64(len(task.ProcessedCADFiles))) {
			task.Status = entity.Complete
		}

		returedTask, err := p.TaskService.Update(task)
		if err != nil {
			log.Fatalf("%s: %s", "Failed to update data: ", err)
		}

		if returedTask.Status == entity.Complete {
			go func() {
				p.Processor.TaskChannel <- returedTask
				log.Printf("[ User: %s > TaskID: %s > Task status: %s]: CAD file (%s) processing plan saved successfully!", e.UserID, returedTask.ID, returedTask.Status, e.CADFileID)
				log.Printf("==========================================================")
			}()

			go func() {
				cadFiles, err := p.CadFileService.FindAll(project.ID.Hex())
				if err != nil {
					log.Fatalf("%s: %s", "Failed to retrieve cad files data: ", err)
				}

				p.Processor.CADFilesChannel <- service.CADFileResponse{UserID: e.UserID, CadFiles: cadFiles}
			}()
		}
	default:
		log.Printf("unknown event type: %T", e)
	}
}
