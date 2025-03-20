package controllers

import (
	"sinceHub/app/middleware"
	"sinceHub/app/models"

	"github.com/revel/revel"
)

type App struct {
	*revel.Controller
}

func (c App) Index() revel.Result {
	return c.Render()
}

func init() {
	revel.OnAppStart(models.InitDB)
	revel.OnAppStart(middleware.Init)
}
