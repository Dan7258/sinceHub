package controllers

import (
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"net/http"
	"sinceHub/app/models"
)

type Tags struct {
	*revel.Controller
}

func (t Tags) CreateTag() revel.Result {
	tag := new(models.Tags)
	err := t.Params.BindJSON(&tag)
	if err != nil {
		t.Response.Status = http.StatusBadRequest
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}

	validate := validator.New()
	err = validate.Struct(tag)
	if err != nil {
		t.Response.Status = http.StatusUnprocessableEntity
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.CreateTag(tag.Name)

	if err != nil {
		t.Response.Status = http.StatusInternalServerError
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}

	t.Response.Status = http.StatusCreated
	return t.RenderJSON(map[string]int{"status": http.StatusCreated})

}

func (t Tags) GetTagByID(id int) revel.Result {
	tag, err := models.GetTagByID(id)
	if err != nil {
		t.Response.Status = http.StatusNotFound
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	t.Response.Status = http.StatusOK
	return t.RenderJSON(tag)
}

func (t Tags) GetTagByName(name string) revel.Result {
	tag, err := models.GetTagByName(name)
	if err != nil {
		t.Response.Status = http.StatusNotFound
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	t.Response.Status = http.StatusOK
	return t.RenderJSON(tag)
}

func (t Tags) DeleteTagByID(id int) revel.Result {
	err := models.DeleteTagByID(id)
	if err != nil {
		t.Response.Status = http.StatusInternalServerError
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	t.Response.Status = http.StatusNoContent
	return t.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (t Tags) UpdateTagByID(id int) revel.Result {
	tag := new(models.Tags)
	err := t.Params.BindJSON(&tag)
	if err != nil {
		t.Response.Status = http.StatusBadRequest
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}

	validate := validator.New()
	err = validate.Struct(tag)
	if err != nil {
		t.Response.Status = http.StatusUnprocessableEntity
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.UpdateTagByID(id, tag)

	if err != nil {
		t.Response.Status = http.StatusInternalServerError
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}

	t.Response.Status = http.StatusNoContent
	return t.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (t Tags) GetAllTags() revel.Result {
	tags, err := models.GetAllTags()
	if err != nil {
		t.Response.Status = http.StatusInternalServerError
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	t.Response.Status = http.StatusOK
	return t.RenderJSON(tags)
}

func (t Tags) AddPublicationsToTag(id uint64) revel.Result {
	var pubIDs []uint64
	err := t.Params.BindJSON(&pubIDs)
	if err != nil {
		t.Response.Status = http.StatusBadRequest
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddPublicationsToTag(id, pubIDs)
	if err != nil {
		t.Response.Status = http.StatusInternalServerError
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	t.Response.Status = http.StatusNoContent
	return t.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (t Tags) DeletePublicationsFromTag(id uint64) revel.Result {
	var pubIDs []uint64
	err := t.Params.BindJSON(&pubIDs)
	if err != nil {
		t.Response.Status = http.StatusBadRequest
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeletePublicationsFromTag(id, pubIDs)
	if err != nil {
		t.Response.Status = http.StatusInternalServerError
		return t.RenderJSON(map[string]string{"error": err.Error()})
	}
	t.Response.Status = http.StatusNoContent
	return t.RenderJSON(map[string]int{"status": http.StatusNoContent})
}
