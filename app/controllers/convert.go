package controllers

import (
	"fmt"
	"github.com/unidoc/unioffice/color"
	"github.com/unidoc/unioffice/common/license"
	"github.com/unidoc/unioffice/document"
	"github.com/unidoc/unioffice/measurement"
	"github.com/unidoc/unioffice/schema/soo/wml"
	"github.com/unidoc/unioffice/spreadsheet"
	"unicode/utf8"

	"os"
	"scinceHub/app/models"
	"strings"
)

func GetFileWithPublicationList(userID uint64, filters models.PublicationFiltres) (string, error) {
	pub := new(models.Publications)
	publications, err := pub.GetPublicationListByFilters(userID, filters)
	if err != nil {
		return "", err
	}
	var filename string
	switch filters.Type {
	case models.Word:
		filename, err = createWordDocument(userID, publications)
	case models.Exel:
		filename, err = createExcelDocument(userID, publications)
	}
	return filename, err

}

func createWordDocument(userID uint64, publications []models.Publications) (string, error) {
	doc := document.New()
	defer doc.Close()

	table := doc.AddTable()
	table.Properties().SetWidthPercent(100)

	borders := table.Properties().Borders()
	borders.SetAll(wml.ST_BorderSingle, color.Auto, 1*measurement.Point)

	row := table.AddRow()
	AddRow(&row, "№")
	AddRow(&row, "Наименование работы")
	AddRow(&row, "Дата публикации")
	AddRow(&row, "Авторы")
	for index, publication := range publications {
		row = table.AddRow()
		AddRow(&row, fmt.Sprint(index+1))
		AddRow(&row, publication.Title)
		AddRow(&row, fmt.Sprint(publication.CreatedAt.Format("02.01.2006")))
		AddRow(&row, getAuthorsFromPublication(publication))
	}
	randomNum, _ := GenerateRandomNumber()
	filename := fmt.Sprintf("public/uploads/%d_%d_list.docx", userID, randomNum)
	err := doc.SaveToFile(filename)
	if err != nil {
		return "", err
	}
	return filename, nil
}

func getAuthorsFromPublication(publication models.Publications) string {
	authors := make([]string, 0)
	for _, profile := range publication.Profiles {
		if profile.MiddleName == "" {
			authors = append(authors, fmt.Sprintf("%s %s", profile.LastName, profile.FirstName))
		} else {
			authors = append(authors, fmt.Sprintf("%s %s %s", profile.LastName, profile.FirstName, profile.MiddleName))
		}
	}
	return strings.Join(authors, ", ")
}

func createExcelDocument(userID uint64, publications []models.Publications) (string, error) {
	exel := spreadsheet.New()
	defer exel.Close()
	sheet := exel.AddSheet()

	boldStyle := exel.StyleSheet.AddCellStyle()
	boldFont := exel.StyleSheet.AddFont()
	boldFont.SetName("Times New Roman")
	boldFont.SetBold(true)
	boldFont.SetSize(12)
	boldStyle.SetFont(boldFont)

	style := exel.StyleSheet.AddCellStyle()
	font := exel.StyleSheet.AddFont()
	font.SetName("Times New Roman")
	font.SetSize(12)
	style.SetFont(font)

	headers := []string{
		"№",
		"Наименование работы",
		"Дата публикации",
		"Авторы",
	}
	row := sheet.AddRow()

	for i, header := range headers {
		SetCellParams(row.AddCell(), boldStyle, header)
		width := measurement.Distance(utf8.RuneCountInString(header))
		if i == 0 {
			sheet.Column(uint32(i) + 1).SetWidth(width * 40)
		} else {
			sheet.Column(uint32(i) + 1).SetWidth(width * 12)
		}
		row.SetHeightAuto()
	}
	for i, publication := range publications {
		row = sheet.AddRow()
		SetCellParams(row.AddCell(), style, fmt.Sprint(i+1))
		SetCellParams(row.AddCell(), style, publication.Title)
		SetCellParams(row.AddCell(), style, fmt.Sprint(publication.CreatedAt.Format("02.01.2006")))
		SetCellParams(row.AddCell(), style, getAuthorsFromPublication(publication))
		row.SetHeightAuto()
	}
	err := exel.Validate()
	if err != nil {
		return "", err
	}
	randomNum, _ := GenerateRandomNumber()
	filename := fmt.Sprintf("public/uploads/%d_%d_list.xlsx", userID, randomNum)
	err = exel.SaveToFile(filename)
	if err != nil {
		return "", err
	}

	return filename, err
}

func SetCellParams(cell spreadsheet.Cell, style spreadsheet.CellStyle, text string) {
	cell.SetString(text)
	cell.SetStyle(style)
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
