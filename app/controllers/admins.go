package controllers

import (
	"github.com/revel/revel"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
	"time"
)

type Admins struct {
	*revel.Controller
}

func (a Admins) LoginAdmin(login, password string) revel.Result {
	admin, err := models.GetAdminsDataByLogin(login)
	if err != nil {
		a.Response.Status = http.StatusUnauthorized
		return a.RenderTemplate("login_admin.html")
	}
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte(password))
	if err != nil {
		a.Response.Status = http.StatusUnauthorized
		return a.RenderTemplate("login_admin.html")
	}
	token, err := middleware.GenerateAdminJWT(admin.ID)

	if err != nil {
		a.Response.Status = http.StatusInternalServerError
		return a.RenderText("Ошибка генерации токена")
	}
	cookie := &http.Cookie{
		Name:     "auth_token_admin",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	a.SetCookie(cookie)

	a.Response.Status = http.StatusFound
	return a.Redirect("/admin")
}

func (a Admins) LogoutAdmin() revel.Result {
	cookie := &http.Cookie{
		Name:     "auth_token_admin",
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteStrictMode,
	}
	a.SetCookie(cookie)
	return a.Redirect("/")
}
