package controllers

import (
	"fmt"
	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/common/license"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
	"os"
	"scinceHub/app/models"
	"strings"
)

func GetFileWithPublicationList(userID uint64, filters models.PublicationFiltres) (string, error) {
	pub := new(models.Publications)
	profile := new(Profiles)
	doc := document.New()
	defer doc.Close()
	publications, err := pub.GetPublicationListByFilters(userID, filters)
	if err != nil {
		return "", err
	}

	table := doc.AddTable()
	table.Properties().SetWidthPercent(100)

	borders := table.Properties().Borders()
	borders.SetAll(wml.ST_BorderSingle, color.Auto, 1*measurement.Point)

	row := table.AddRow()
	AddRow(&row, "№")
	AddRow(&row, "Наименование работы")
	AddRow(&row, "Дата публикации")
	AddRow(&row, "Авторы")
	authors := make([]string, 0)
	for index, publication := range publications {
		row = table.AddRow()
		AddRow(&row, fmt.Sprint(index+1))
		AddRow(&row, publication.Title)
		AddRow(&row, fmt.Sprint(publication.CreatedAt.Format("02.01.2006")))
		for _, profile := range publication.Profiles {
			if profile.MiddleName == "" {
				authors = append(authors, fmt.Sprintf("%s %s", profile.LastName, profile.FirstName))
			} else {
				authors = append(authors, fmt.Sprintf("%s %s %s", profile.LastName, profile.FirstName, profile.MiddleName))
			}
		}
		AddRow(&row, strings.Join(authors, ", "))
		authors = make([]string, 0)
	}
	randomNum, _ := profile.GenerateRandomNumber()
	filename := fmt.Sprintf("public/uploads/%d_%d_list.docx", userID, randomNum)
	err = doc.SaveToFile(filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func AddRow(row *document.Row, text string) {
	run := row.AddCell().AddParagraph().AddRun()
	run.Properties().SetFontFamily("Times New Roman")
	run.Properties().SetSize(14)
	run.AddText(text)
}

func InitLicense() {
	err := license.SetMeteredKey(os.Getenv("UNIDOC_LICENSE_API_KEY"))
	if err != nil {
		panic(err)
	}
}
