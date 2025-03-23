package controllers

import (
	"sinceHub/app/middleware"
	"sinceHub/app/models"
	"sinceHub/app/smtp"

	"github.com/revel/revel"
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
}
