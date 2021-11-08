package controllers

import (
    "container/list"
	"encoding/json"
    "strconv"
    "Spider/models"
	"github.com/astaxie/beego"
	// "github.com/astaxie/beego/logs"
)

type MainController struct {
	beego.Controller
}

type ImageInfo struct {
    Id string `json:"id"`
    Title string `json:"title"`
    Path string `json:"path"`
    Image []byte `json:"image"`
}

type MagnetInfo struct {
    Id string `json:"id"`
    No string `json:"no"`
    Magnets string `json:"magnets"`
}

var TotalPage = 0
var PageCount = 12
var MagnetPageCount = 10
var PageSize = 10
var CurrentTable = models.TABLE_MAGNET

func DisplayImageFromDatabase(items *list.List) string {
    index := 0
    images := make([]ImageInfo, 0, items.Len())
    for item := items.Front(); nil != item; item = item.Next() {
        var image ImageInfo
        image.Id = strconv.Itoa(index)
        image.Title = item.Value.(models.NormalImageData).Title
        // image.Title = "title"
        image.Path = item.Value.(models.NormalImageData).Path
        // image.Path = "path"
        image.Image = item.Value.(models.NormalImageData).Data
        images = append(images, image)
        index++
    }
    data, err := json.Marshal(images)
    if nil == err {
        return string(data)
    }
    return ""
}

func DisplayImageFromLocal(items *list.List) string {
    index := 0
    images := make([]ImageInfo, 0, items.Len())
    for item := items.Front(); nil != item; item = item.Next() {
        var image ImageInfo
        image.Id = strconv.Itoa(index)
        image.Title = item.Value.(models.NormalImageData).Title
        // image.Title = "title"
        image.Path = item.Value.(models.NormalImageData).Path
        // image.Path = "path"
        images = append(images, image)
        index++
    }
    data, err := json.Marshal(images)
    if nil == err {
        return string(data)
    }
    return ""
}

func (this *MainController) Get() {

    currentPage := this.GetString("page")
    pageType := this.GetString("type")
    if pageType == "" {
        pageType = models.TABLE_MAGNET
    }
    // get total page count
    if pageType == models.TABLE_MAGNET {
        PageCount = MagnetPageCount
    }
    TotalPage = models.GetItemsCount(CurrentTable, PageCount)

    var curPage int
    CurrentTable = pageType
	if len(currentPage) == 0 {
        this.Data["PageNo"] = 1
        curPage = 1
	} else {
        this.Data["PageNo"] = currentPage
        curPage, _ = strconv.Atoi(currentPage)
    }
    lower := PageCount * (curPage - 1)
    items := models.LoadItemsFromDatabase(CurrentTable, lower, PageCount)
    if pageType == models.TABLE_NEW_TIMES || pageType == models.TABLE_FIRE {
        this.Data["Json"] = DisplayImageFromDatabase(items)
    } else if pageType == models.TABLE_NEW_TIMES_NO_PIC || pageType == models.TABLE_FIRE_NO_PIC {
        this.Data["Json"] = DisplayImageFromLocal(items)
    } else if pageType == models.TABLE_MAGNET {
        magnets := make([]MagnetInfo, 0, items.Len())
        index := 0
        for item := items.Front(); nil != item; item = item.Next() {
            var magnet MagnetInfo
            magnet.Id = strconv.Itoa(index)
            magnet.No = item.Value.(string)

            // convert list to string
            _magnetList := models.LoadDetailsFromDatabase(CurrentTable, PageCount * (curPage - 1) + index + 1)
            _magnets := make([]models.MagnetData, 0, _magnetList.Len())
            for m := _magnetList.Front(); nil != m; m = m.Next() {
                _magnets = append(_magnets, m.Value.(models.MagnetData))
            }
            __magnets, err := json.Marshal(_magnets)
            if nil == err {
                magnet.Magnets = string(__magnets)
            }

            magnets = append(magnets, magnet)
            index++
        }
        data, err := json.Marshal(magnets)
        if nil == err {
            this.Data["Magnets"] = string(data)
        }
    }
    this.Data["PageType"] = pageType
    this.Data["PageCount"] = items.Len()
	this.Data["PageSize"] = PageSize
    this.Data["TotalPage"] = TotalPage
	this.TplName = "index.html"
}

func init() {
    models.InitDir()
	// connect to redis
    models.ConnectRedis("127.0.0.1:6379")
    // init data
    models.GetCapcity(models.TABLE_NEW_TIMES)
    models.GetCapcity(models.TABLE_NEW_TIMES_NO_PIC)
    models.GetCapcity(models.TABLE_FIRE)
    models.GetCapcity(models.TABLE_MAGNET)
    models.NewTimesTableCount = models.GetItemsTotalCount(models.TABLE_NEW_TIMES)
}
