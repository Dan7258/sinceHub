package controllers

import (
	"fmt"
	"net/http"
	"sinceHub/app/models"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
)

type Profiles struct {
	*revel.Controller
}

func (p Profiles) CreateProfile() revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	validate := validator.New()
	err = validate.Struct(profile)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.CreateProfile(profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusCreated
	return p.Redirect("/login")
}

func (p Profiles) Login(login, password string) revel.Result {
	user, err := models.GetProfileLoginData(login)
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderTemplate("Profiles/login.html")
	}
	if user.Password != password {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderTemplate("Profiles/login.html")
	}

	p.Session["user"] = fmt.Sprintf("%d", user.ID)
	p.Response.Status = http.StatusFound
	return p.Redirect("/profile")
}

func (p Profiles) ShowLogin() revel.Result {
	return p.RenderTemplate("Profiles/login.html")
}

func (p Profiles) Logout() revel.Result {
	for k := range p.Session {
		delete(p.Session, k)
	}
	p.Response.Status = http.StatusNoContent
	return p.Redirect("/login")
}

func (p Profiles) GetProfileByID(id uint64) revel.Result {
	profile, err := models.GetProfileByID(id)
	if err != nil {
		p.Response.Status = http.StatusNotFound
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profile)
}
func (p Profiles) ShowUserProfile() revel.Result {
	_, ok := p.Session["user"]
	if !ok {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	return p.RenderTemplate("Profiles/profile.html")
}
func (p Profiles) GetUserData() revel.Result {
	user, ok := p.Session["user"]
	if !ok {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	userID, _ := strconv.ParseUint(user.(string), 10, 64)
	profile, _ := models.GetUserProfile(userID)
	fmt.Println(profile)
	return p.RenderJSON(profile)
}

func (p Profiles) DeleteProfileByID(id uint64) revel.Result {
	err := models.DeleteProfileByID(id)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeleteProfileByLogin(login string) revel.Result {
	err := models.DeleteProfileByLogin(login)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) UpdateProfileByID(id uint64) revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	validate := validator.New()
	err = validate.Struct(profile)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.UpdateProfileByID(id, profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) UpdateProfileByLogin(login string) revel.Result {
	profile := new(models.Profiles)
	err := p.Params.BindJSON(profile)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	validate := validator.New()
	err = validate.Struct(profile)
	if err != nil {
		p.Response.Status = http.StatusUnprocessableEntity
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	err = models.UpdateProfileByLogin(login, profile)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}

	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) GetAllProfiles() revel.Result {
	profiles, err := models.GetAllProfiles()
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusOK
	return p.RenderJSON(profiles)
}

func (p Profiles) AddPublicationsToProfile(id uint64) revel.Result {
	var pubIDs []uint64
	err := p.Params.BindJSON(&pubIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddPublicationsToProfile(id, pubIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeletePublicationsFromProfile(id uint64) revel.Result {
	var pubIDs []uint64
	err := p.Params.BindJSON(&pubIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeletePublicationsFromProfile(id, pubIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) AddSubscribersToProfile(id uint64) revel.Result {
	var subIDs []uint64
	err := p.Params.BindJSON(&subIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.AddSubscribersToProfile(id, subIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}

func (p Profiles) DeleteSubscribersFromProfile(id uint64) revel.Result {
	var subIDs []uint64
	err := p.Params.BindJSON(&subIDs)
	if err != nil {
		p.Response.Status = http.StatusBadRequest
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	err = models.DeleteSubscribersFromProfile(id, subIDs)
	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderJSON(map[string]string{"error": err.Error()})
	}
	p.Response.Status = http.StatusNoContent
	return p.RenderJSON(map[string]int{"status": http.StatusNoContent})
}
