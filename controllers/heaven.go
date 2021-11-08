package controllers

import (
	"Spider/color"
    "Spider/models"
	"github.com/astaxie/beego"
)

type HeavenController struct {
	beego.Controller
}

var (
	heavenpage   = make(chan string, 100)
	heavenmovie  = make(chan models.MovieHeaven, 100)
)

func Heaven() {
	for {
		select {
		case page := <- heavenpage:
			movie := models.GetHeavenContent(page)
			heavenmovie <- movie
		default:
		}
	}
}

func HeavenStoring() {
	for {
		select {
		case movie := <- heavenmovie:
			color.New(color.FgHiGreen).Printf("%s --- ", movie.Url)
			color.New(color.FgHiMagenta).Printf("%s\n", movie.MovieName)
			models.AddMovieHeavenToDatabase(movie)
		default:
		}
	}
}

func (this *HeavenController) MovieHeaven() {
	
	go Heaven()
	go HeavenStoring()
	
	pages := models.GetHeavenList(models.MOIVE_HEAVEN_URL + models.MOIVE_HEAVEN_INDEX)
	for page := pages.Front(); nil != page; page = page.Next() {
		heavenpage <- page.Value.(string)
	}
	this.Ctx.Redirect(302, "/?type=magnet")
}
