package controllers

import (
	"github.com/revel/revel"
	"sinceHub/app/middleware"
	"sinceHub/app/models"
	"sinceHub/app/smtp"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func init() {
	revel.OnAppStart(middleware.Init)
	revel.OnAppStart(models.InitDB)
	revel.OnAppStart(smtp.InitSMTP)
	revel.OnAppStart(InitLicense)
}
