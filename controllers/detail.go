package controllers


import (
	"strconv"
	"Spider/models"
	"github.com/astaxie/beego"
)

type DetailController struct {
	beego.Controller
}

func (this *DetailController) ImageDetail() {
	page, _ := strconv.Atoi(this.GetString("page"))
	index, _ := strconv.Atoi(this.GetString("index"))
	isLoadFromFile, _ := strconv.Atoi(this.GetString("file"))
	title := this.GetString("title")
	group := PageCount * (page - 1) + index + 1
	images := models.LoadDetailsFromDatabase(CurrentTable, group)
	this.Data["IsLoadFromFile"] = isLoadFromFile
	this.Data["ImageCount"] = images.Len()
	this.Data["Title"] = title
	this.Data["Json"] = DisplayImageFromLocal(images)
	this.Data["ImageCount"] = images.Len()
	this.TplName = "imagedetails.html"
}
