package controllers

import (
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"net/http"
	"sinceHub/app/models"
)

type Publications struct {
	*revel.Controller
}

func (p Publications) CreatePublication() revel.Result {
	pub := new(models.Publications)
	err := p.Params.BindJSON(pub)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	if pub.Abstract != "" && len(pub.Abstract) < 2 {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": "Краткое сведение слишком краткое!"})
	}
	validate := validator.New()
	err = validate.Struct(pub)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.CreatePublication(pub)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusCreated
	return p.RenderJSON(map[string]int{"status": http.StatusCreated})
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

	if pub.Title != "" && len(pub.Title) < 2 {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": "Заголовок слишком короткий!"})
	}

	if pub.Content != "" && len(pub.Content) < 2 {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": "В статье мало текста!"})
	}

	if pub.Abstract != "" && len(pub.Abstract) < 2 {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": "Краткое сведение слишком краткое!"})
	}

	err = models.UpdatePublicationByID(id, pub)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Publications) GetAllPublications() revel.Result {
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
