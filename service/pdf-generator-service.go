package service

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/WilfredDube/fxtract-backend/entity"
	"github.com/johnfercher/maroto/pkg/color"
	"github.com/johnfercher/maroto/pkg/consts"
	"github.com/johnfercher/maroto/pkg/pdf"
	"github.com/johnfercher/maroto/pkg/props"
)

const (
	directory = "pdfs"
)

type PDFService interface {
	GeneratePDF(processingPlan *entity.ProcessingPlan) (string, error)
	getSmallContent(configMap map[int]entity.BendFeature, bendingSequence []entity.BendingSequence) ([]string, [][]string)
}

type pdfService struct{}

func NewPDFService() PDFService {
	_, err := os.Stat(directory)

	if os.IsNotExist(err) {
		errDir := os.MkdirAll(directory, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	// TODO: connect to the cloud via seperate cloud service
	return &pdfService{}
}

func (p *pdfService) GeneratePDF(processingPlan *entity.ProcessingPlan) (string, error) {
	confMap := map[int]entity.BendFeature{}
	for _, v := range processingPlan.BendFeatures {
		confMap[int(v.BendID)] = v
	}

	m := pdf.NewMaroto(consts.Portrait, consts.A4)
	m.SetPageMargins(10, 15, 10)
	// m.SetBorder(true)

	headerSmall, smallContent := p.getSmallContent(confMap, processingPlan.BendingSequences)

	m.SetAliasNbPages("{nb}")
	m.SetFirstPageNb(1)

	m.RegisterHeader(func() {
		m.Row(25, func() {
			m.Col(7, func() {
				m.Text("Fxtract", props.Text{
					Size:  24,
					Top:   8,
					Style: consts.Bold,
					Color: color.Color{
						Red:   3,
						Green: 166,
						Blue:  166,
					},
				})
			})

			m.ColSpace(2)

			m.Col(3, func() {
				m.QrCode("https://fxtract.com/projects/"+processingPlan.CADFileID.Hex()+"/"+processingPlan.CADFileID.Hex(), props.Rect{
					Center:  false,
					Percent: 95,
					Left:    23,
				})
			})
		})

		m.Line(0.2)

		m.Row(12, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Engineer:%43v.", processingPlan.Engineer), props.Text{
					Top:   6,
					Style: consts.Bold,
				})
			})
		})

		layout := "02 / 01 / 2006"
		t := time.Unix(processingPlan.CreatedAt, 0)

		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(fmt.Sprintf("Date:%50v.", t.Format(layout)), props.Text{
					Top:   2,
					Style: consts.Bold,
				})
			})
		})

		m.Line(0.2)

		m.Row(10, func() {
			m.Col(7, func() {
				m.Text(fmt.Sprintf("Project name:%36s", processingPlan.ProjectTitle), props.Text{
					Top:   3,
					Style: consts.Bold,
				})
			})

			m.Col(5, func() {
				m.Text(fmt.Sprintf("Modules:%30d", processingPlan.Modules), props.Text{
					Top:   3,
					Style: consts.Bold,
				})
			})
		})

		m.Row(10, func() {
			m.Col(7, func() {
				m.Text(fmt.Sprintf("Part name:%40s", processingPlan.FileName), props.Text{
					Top:   2,
					Style: consts.Bold,
				})
			})

			m.Col(5, func() {
				m.Text(fmt.Sprintf("Part no:%40s", processingPlan.PartNo), props.Text{
					Top:   2,
					Style: consts.Bold,
				})
			})
		})

		m.Row(10, func() {
			m.Col(7, func() {
				m.Text(fmt.Sprintf("Material:%42s", processingPlan.Material), props.Text{
					Top:   2,
					Style: consts.Bold,
				})
			})

			m.Col(5, func() {
				m.Text(fmt.Sprintf("Bending force:%28.2f", processingPlan.BendingForce), props.Text{
					Top:   2,
					Style: consts.Bold,
				})
			})
		})

		m.Line(0.2)
		m.Row(10, func() {
			m.Row(10, func() {
				m.Col(4, func() {
					m.Text(fmt.Sprintf("Number of tools:%20d", processingPlan.Tools), props.Text{
						Top:   3,
						Style: consts.Bold,
					})
				})
				m.Col(4, func() {
					m.Text(fmt.Sprintf("Number of rotations:%14d", processingPlan.Rotations), props.Text{
						Top:   3,
						Style: consts.Bold,
					})
				})
				m.Col(4, func() {
					m.Text(fmt.Sprintf("Number of flips:%27d", processingPlan.Flips), props.Text{
						Top:   3,
						Style: consts.Bold,
					})
				})
			})
			m.Row(10, func() {
				m.Col(4, func() {
					m.Text(fmt.Sprintf("Quantity:%32d", processingPlan.Quantity), props.Text{
						Top:   3,
						Style: consts.Bold,
					})
				})
				m.Col(4, func() {
					m.Text(fmt.Sprintf("Planning time:%25.3f", processingPlan.ProcessingTime), props.Text{
						Top:   3,
						Style: consts.Bold,
					})
				})
				m.Col(4, func() {
					m.Text(fmt.Sprintf("Estimated production time:%10.1f", processingPlan.EstimatedManufacturingTime), props.Text{
						Top:   3,
						Style: consts.Bold,
					})
				})
			})
			m.Line(0.2)
		})
	})

	m.RegisterFooter(func() {
		m.Row(40, func() {
			m.Col(6, func() {
				m.Signature("Checked by (Name)", props.Font{
					// Family: consts.Courier,
					Style: consts.BoldItalic,
					Size:  8,
				})
			})

			m.Col(6, func() {
				m.Signature("Signature", props.Font{
					// Family: consts.Courier,
					Style: consts.BoldItalic,
					Size:  8,
				})
			})
		})
		m.Row(10, func() {
			m.Col(12, func() {
				m.Text(strconv.Itoa(m.GetCurrentPage())+"/{nb}", props.Text{
					Style: consts.BoldItalic,
					Align: consts.Right,
					Size:  9,
				})
			})
		})
	})

	m.Row(20, func() {
		m.Col(12, func() {
			m.Text("Bending sequence", props.Text{
				Top:   8,
				Style: consts.Bold,
				Align: consts.Center,
				Size:  12,
			})
		})
	})

	m.Line(0.2)
	m.Row(10, func() {})
	m.TableList(headerSmall, smallContent, props.TableList{
		ContentProp: props.TableListContent{
			GridSizes: []uint{1, 1, 2, 2, 2, 2, 2},
		},
		HeaderProp: props.TableListContent{
			GridSizes: []uint{1, 1, 2, 2, 2, 2, 2},
		},
		Align: consts.Center,
	})

	m.Row(10, func() {})
	m.Line(0.2)
	m.Row(10, func() {
		m.Text("NB: All units are in degrees, mm, kN and sec", props.Text{
			Top:   3,
			Style: consts.Italic,
			Align: consts.Center,
			Size:  8,
		})
	})
	m.Line(0.2)

	tempFilePath := directory + "/" + processingPlan.ID.Hex() + ".pdf"
	err := m.OutputFileAndClose(tempFilePath)
	if err != nil {
		return "", fmt.Errorf("could not generate PDF: %v", err)
	}

	// TODO: generate and return the URL to the blob file
	return tempFilePath, nil
}

func (p *pdfService) getSmallContent(configMap map[int]entity.BendFeature, bendingSequence []entity.BendingSequence) ([]string, [][]string) {
	header := []string{"Op", "Bend ID", "Bend Angle", "Length", "Radius", "Direction", "Tool"}

	contents := [][]string{}
	var direction string
	for i, sequence := range bendingSequence {
		feature := configMap[int(sequence.BendID)]
		if feature.Direction == 1 {
			direction = "Inside"
		} else {
			direction = "Outside"
		}

		contents = append(contents, []string{fmt.Sprint(i + 1), fmt.Sprint(feature.BendID), fmt.Sprint(feature.Angle),
			fmt.Sprint(feature.Length), fmt.Sprint(feature.Radius), direction, fmt.Sprint(feature.ToolID)})
	}

	return header, contents
}
