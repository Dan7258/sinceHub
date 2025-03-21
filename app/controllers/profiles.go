package controllers

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/revel/revel"
	"net/http"
	"sinceHub/app/middleware"
	"sinceHub/app/models"
	"time"
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
		return p.RenderTemplate("login.html")
	}
	if user.Password != password {
		p.Response.Status = http.StatusUnauthorized
		return p.RenderTemplate("login.html")
	}
	token, err := middleware.GenerateJWT(user.ID)

	if err != nil {
		p.Response.Status = http.StatusInternalServerError
		return p.RenderText("Ошибка генерации токена")
	}
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	p.SetCookie(cookie)

	p.Response.Status = http.StatusFound
	return p.Redirect("/profile")
}

func (p Profiles) ShowLoginPage() revel.Result {
	return p.RenderTemplate("login.html")
}

func (p Profiles) ShowSettingsPage() revel.Result {
	return p.RenderTemplate("settings.html")
}

func (p Profiles) Logout() revel.Result {
	cookie := &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	p.SetCookie(cookie)
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
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	return p.RenderTemplate("profile.html")
}
func (p Profiles) GetUserData() revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
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
