package controllers

import (
	"github.com/revel/revel"
	"scinceHub/app/middleware"
)

func (p Profiles) ShowLoginPage() revel.Result {
	return p.RenderTemplate("login.html")
}

func (p Profiles) ShowRegisterPage() revel.Result {
	return p.RenderTemplate("register.html")
}

func (p Profiles) ShowSettingsPage() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	return p.RenderTemplate("settings.html")
}

func (p Profiles) ShowUserProfile() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}

	return p.RenderTemplate("my_profile.html")
}

func (p Profiles) ShowUserProfilePageByID(id uint64) revel.Result {
	userID, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		//p.Response.Status = http.StatusUnauthorized
		return p.Redirect("/login")
	}
	if userID == id {
		return p.RenderTemplate("my_profile.html")
	}
	p.ViewArgs["id"] = id
	return p.RenderTemplate("profile_by_id.html")
}

func (p Publications) ShowCreatePublicationPage() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	return p.RenderTemplate("create_publication.html")
}

func (p Publications) ShowUpdatePublicationPage() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	return p.RenderTemplate("update_publication.html")

}

func (p Publications) ShowPublications() revel.Result {
	return p.RenderTemplate("publications.html")
}

func (p Profiles) ShowAuthorsPage() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	return p.RenderTemplate("authors.html")
}

func (p Profiles) ShowMySubscribers() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	return p.RenderTemplate("my_subscribers.html")
}

func (p Profiles) ShowMySubscribes() revel.Result {
	_, err := middleware.ValidateJWT(p.Request, "auth_token")
	if err != nil {
		return p.Redirect("/login")
	}
	return p.RenderTemplate("my_subscribes.html")
}
