package controllers


import (
	"strings"

	"Spider/color"
	"Spider/models"
	"github.com/astaxie/beego"
)

type ManagerController struct {
	beego.Controller
}

func (this *ManagerController) ClearSet() {
	operation := this.GetString("op")
	if operation == "all" {
		this.Ctx.WriteString("all ")
		models.ClearAll()
	} else {
		values := strings.Split(operation, " ")
		for i := 0; i < len(values); i++ {
			function, ok := models.SetsOperation[values[i]]
			if ok {
				this.Ctx.WriteString(values[i] + " ")
				function()
			}
		}
	}
	color.ColorLog.Log("[CLEAR DONE!]", color.RedBold)
	this.Ctx.Redirect(302, "/?type=magnet")
}

func (this *ManagerController) LoadImage() {
	table := this.GetString("table")
	path := this.GetString("path")
	if table == models.TABLE_NEW_TIMES {
		models.LoadImageFromDatabase(table, path)
	}
	color.ColorLog.Log("[LOAD DONE!]", color.RedBold)
	this.Ctx.Redirect(302, "/?type=magnet")
}

func (this *ManagerController) LoadToDatabase() {
	table := this.GetString("table")
	models.SaveDatabaseFromLocal(table)
	color.ColorLog.Log("[LOAD TO DATABASE DONE!]", color.RedBold)
	this.Ctx.Redirect(302, "/?type=magnet")
}