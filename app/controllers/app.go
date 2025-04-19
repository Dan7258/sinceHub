package controllers

import (
	"github.com/revel/revel"
	"scinceHub/app/middleware"
	"scinceHub/app/models"
	"scinceHub/app/smtp"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func init() {
	revel.OnAppStart(middleware.InitENV)
	revel.OnAppStart(models.InitDB)
	revel.OnAppStart(smtp.InitSMTP)
	revel.OnAppStart(InitLicense)
}
