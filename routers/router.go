package routers

import (
	"Spider/controllers"
	"github.com/astaxie/beego"
)

func init() {
	beego.Router("/", &controllers.MainController{})
	beego.Router("/imagedetail", &controllers.DetailController{}, "get:ImageDetail")
	beego.Router("/crawl/movie", &controllers.CrawlController{}, "*:CrawlMovie")
	beego.Router("/crawl/image", &controllers.CrawlController{}, "*:CrawlImage")
	beego.Router("/crawl/retrieval", &controllers.CrawlController{}, "*:ProcessUnsuccessPages")
	beego.Router("/crawl/savepags", &controllers.CrawlController{}, "*:CrawlImagSavePages")
	beego.Router("/crawl/concurrency", &controllers.CrawlController{}, "*:CrawlImageConcurrent")
	beego.Router("/heaven", &controllers.HeavenController{}, "*:MovieHeaven")
	beego.Router("/clear", &controllers.ManagerController{}, "*:ClearSet")
	beego.Router("/load", &controllers.ManagerController{}, "*:LoadImage")
	beego.Router("/load/to/databse", &controllers.ManagerController{}, "*:LoadToDatabase")
	beego.Router("/print", &controllers.PrintController{}, "*:Print")
}
