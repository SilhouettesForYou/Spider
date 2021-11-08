package controllers

import (
	// "time"
	"Spider/models"
	"Spider/color"
	
	// "github.com/astaxie/beego/logs"
)

var (
	// channels for movie
	page 	= make(chan models.ListInfo, 100)
	magnets	= make(chan string, 100)
	images  = make(chan models.ListInfo, 100)
	datas   = make(chan models.ImageInfo, 100)
	// channels for normal image
	saving  = make(chan models.NormalImageInfo, 100)
	srcs    = make(chan models.NormalImageInfo, 100)
	loading = make(chan string, 100)
	storing = make(chan models.NormalImageDetail, 100)
	// channel for unseccess pages
	unsrcs    = make(chan models.NormalImageInfo, 100)
	unstoring = make(chan models.NormalImageDetail, 100)
	// channels for stopping
	processimagedone   = make(chan bool)
	processmoviedone   = make(chan bool)
	storeimagedone     = make(chan bool)
	storemoviedone     = make(chan bool)
	processunimagedone = make(chan bool)
	storeunimagedone   = make(chan bool)
)

func PushInContent(info models.ListInfo) {
	page <- info
}

func PushInMagnet(no string) {
	magnets <- no
}

func PushInImage(info models.ListInfo) {
	images <- info
}

func PushInData(info models.ImageInfo) {
	datas <- info
}

func PushInSaving(save models.NormalImageInfo) {
	saving <- save
}

func PushInSrcs(src models.NormalImageInfo) {
	srcs <- src
}

func PushInLoading(load string) {
	loading <- load
}

func PushInStoring(store models.NormalImageDetail) {
	storing <- store
}

func PushInUnsrcs(src models.NormalImageInfo) {
	srcs <- src
}

func PushInUnstoring(store models.NormalImageDetail) {
	storing <- store
}

func ProcessMovie() {
	for {
		select {
		case content := <- page:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "page", len(page))
			PushInMagnet(content.Number)
		case info := <- images:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "images", len(images))
			imageInfos := models.GetImagesOnPage(info.Page, info.Title)
			for src, filename := range imageInfos {
				models.DownloadFile(src, filename)
				PushInData(models.ImageInfo{ info.Number, filename })
			}
		case <- processmoviedone:
			break
		default:
		}
	}
}

func StoringMovie() {
	for {
		select {
		case no := <- magnets:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "magnets", len(magnets))
			if len(no) != 0 {
				pages := models.GetMagnetsPages(no)
				for page := pages.Front(); nil != page; page = page.Next() {
					items := models.GetMagnetsOnPage(page.Value.(string))
					for item := items.Front(); nil != item; item = item.Next() {
						models.AddMagnetToDatabase(no, item.Value.(models.MagnetInfo))
					}
				}
			}
		case data := <- datas:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "datas", len(datas))
			models.AddImageToDatabase(data)
		case <- storemoviedone:
			break
		default:
		}
	}
}

func ProcessImage() {
	for {
		select {
		// save image from net
		case save := <- saving:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "saving", len(saving))
			imageInfos := models.GetNormalImagePageContent(save, indexType)
			for src, filename := range imageInfos {
				//logs.Trace("[Downloading from ] ", src)
				models.DownloadFile(src, filename)
				PushInLoading(filename)
			}
		// storing image into database directly
		case src := <- srcs:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "srcs", len(srcs))
			urls := models.GetNormalImagePageConentWithoutPath(src, indexType)
			for url := urls.Front(); nil != url; url = url.Next() {
				//color.New(color.FgWhite).Printf("[Pushing in storing...]\r")
				PushInStoring(url.Value.(models.NormalImageDetail))
			}
		case <- processimagedone:
			break
		default:
			color.ColorLog.Log(" WAITING\r", color.GreenLight)
		}
	}
}

func StoringImage(id int) {
	for {
		select {
		// stroing image into database through read picture from local
		case load := <- loading:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "loading", len(loading))
			models.AddNormalImageToDatabase(load, indexType)
		case store := <- storing:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "storing ", len(storing))
			models.AddNormalImageToDatabaseDirectly(store.Id, len(storing), id, "storing", store.Title, store.Src, store.Path, store.Table)
			// models.AddNormalImageToDatabaseWithoutPic(store.Title, store.Src, store.Path, "newtimesnopic")
		case <- storeimagedone:
			break
		default:
			color.ColorLog.Log(" WAITING\r", color.YellowLight)
		}
	}
}

func ProcessUnImage() {
	for {
		select {
		// storing image into database directly
		case unsrc := <- unsrcs:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "srcs", len(srcs))
			urls := models.GetNormalImagePageConentWithoutPath(unsrc, indexType)
			for url := urls.Front(); nil != url; url = url.Next() {
				//color.New(color.FgWhite).Printf("[Pushing in storing...]\r")
				PushInUnstoring(url.Value.(models.NormalImageDetail))
			}
		case <- processunimagedone:
			break
		default:
		}
	}
}

func StoringUnImage() {
	for {
		select {
		case unstore := <- unstoring:
			//color.New(color.FgRed).Printf("%s  [Current channel remained number %d]\n", "storing ", len(storing))
			models.AddNormalImageToDatabaseDirectly(unstore.Id, len(unstoring), RequestId, "unstoring", unstore.Title, unstore.Src, unstore.Path, indexType)
			// models.AddNormalImageToDatabaseWithoutPic(store.Title, store.Src, store.Path, "newtimesnopic")
		case <- storeunimagedone:
			break
		default:
		}
	}
}
