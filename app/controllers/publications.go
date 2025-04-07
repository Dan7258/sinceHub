package controllers

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"io"
	"net/http"
	"os"
	"sinceHub/app/middleware"
	"sinceHub/app/models"
	"strconv"
)

type Publications struct {
	*revel.Controller
}

func (p Publications) CreatePublication() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	pub := new(models.Publications)
	pub.Title = p.Params.Get("title")
	pub.Abstract = p.Params.Get("abstract")
	validate := validator.New()
	err = validate.Struct(pub)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	fileHeader, ok := p.Params.Files["file"]
	if !ok || len(fileHeader) == 0 {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": "No file uploaded"})
	}

	file, err := fileHeader[0].Open()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось открыть файл"})
	}
	defer file.Close()

	randomNumber, _ := Profiles{}.generateRandomNumber()
	filePath := fmt.Sprintf("public/uploads/%d_%s_%s", userID, randomNumber, fileHeader[0].Filename)
	dst, err := os.Create(filePath)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Не удалось сохранить файл"})
	}
	defer dst.Close()

	_, err = io.Copy(dst, file)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": "Ошибка при сохранении файла"})
	}
	pub.FileLink = filePath

	rawTagIDs := p.Params.Values["tags[]"]
	tagIDs := make(map[uint64]interface{}, 0)
	for _, ID := range rawTagIDs {
		tagID, _ := strconv.ParseUint(ID, 10, 64)
		tagIDs[tagID] = nil
	}
	rawCoauthors := p.Params.Values["coauthors[]"]
	coauthorIDs := make(map[uint64]interface{}, 0)
	coauthorIDs[userID] = nil
	for _, ID := range rawCoauthors {
		coauthorID, _ := strconv.ParseUint(ID, 10, 64)
		coauthorIDs[coauthorID] = nil
	}
	err = models.CreatePublication(pub, tagIDs, coauthorIDs)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusCreated

	return p.Redirect("/profile")
}

//func (p Publications) getPublicationFile(name string) revel.Result {
//
//}

func (p Publications) ShowCreatePublicationPage() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	return p.RenderTemplate("create_publication.html")
}

func (p Publications) GetPublicationByID(id uint64) revel.Result {
	pub, err := models.GetPublicationByID(id)
	if err != nil {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(pub)
}

func (p Publications) DeletePublicationByID(id int) revel.Result {
	err := models.DeletePublicationByID(id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) UpdatePublicationByID(id int) revel.Result {
	pub := new(models.Publications)
	err := p.Params.BindJSON(pub)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	validate := validator.New()
	err = validate.Struct(pub)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.UpdatePublicationByID(id, pub)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) ShowPublications() revel.Result {
	return p.RenderTemplate("publications.html")
}

func (p Publications) GetPublicationsData() revel.Result {
	pub, err := models.GetAllPublications()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(pub)
}

func (p Publications) AddTagsToPublication(id uint64) revel.Result {
	var tagIDs []uint64
	err := p.Params.BindJSON(&tagIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddTagsToPublication(id, tagIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) DeleteTagsFromPublication(id uint64) revel.Result {
	var tagIDs []uint64
	err := p.Params.BindJSON(&tagIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeleteTagsFromPublication(id, tagIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) AddProfilesToPublication(id uint64) revel.Result {
	var profileIDs []uint64
	err := p.Params.BindJSON(&profileIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddProfilesToPublication(id, profileIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) DeleteProfilesFromPublication(id uint64) revel.Result {
	var profileIDs []uint64
	err := p.Params.BindJSON(&profileIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeleteProfilesFromPublication(id, profileIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}
