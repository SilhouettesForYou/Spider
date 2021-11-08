package controllers


import (
	"bufio"
	"container/list"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"Spider/models"
	"Spider/color"
	"github.com/astaxie/beego"
	"github.com/astaxie/beego/logs"
)

type CrawlController struct {
	beego.Controller
}

var indexType = ""
var RequestId = 1

func IsAllDone() bool {
	if len(page) != 0 {
		return false
	} else if len(magnets) != 0 {
		return false
	} else if len(images) != 0 {
		return false
	} else if len(datas) != 0 {
		return false
	} else if len(saving) != 0 {
		return false
	} else if len(srcs) != 0 {
		return false
	} else if len(loading) != 0 {
		return false
	} else if len(storing) != 0 {
		return false
	} else if len(unsrcs) != 0 {
		return false
	} else if len(unstoring) != 0 {
		return false
	}
	return true
}

func ParseImageIndex(download string, pageIndex *list.List) {
	pageCount := 1
	for p := pageIndex.Front(); nil != p; p = p.Next() {
		pages := models.GetImageList(p.Value.(string))
		color.ColorLog.Logs(
			models.GetDateTime(), 				   color.MagentaLight, 
			" Parsing Index ", 	  				   color.CyanBold, 
			models.ReplaceWithSymbolsEqual(p.Value.(string)), color.GrayLight,
			" Page ", 	  						   color.CyanBold, 
			fmt.Sprintf("%d", pageCount), 	  	   color.RedBold,
			" Items ", 	  					  	   color.CyanBold, 
			fmt.Sprintf("%d\n", pages.Len()), 	   color.RedBold)
		for page := pages.Front(); nil != page; page = page.Next() {
			// logs.Trace("[Parsing Page] ", page.Value.(models.NormalImageInfo).Url)
			if models.IsNewTimesVisit(page.Value.(models.NormalImageInfo).Url) {
				color.ColorLog.Log(" [SKIPPING]\r", color.YellowLight)
				continue
			}
			models.AddToNewTimesVisitedSet(page.Value.(models.NormalImageInfo).Url)
			if download == "0" {
				PushInSrcs(page.Value.(models.NormalImageInfo))
				color.ColorLog.Log(" [PUSHING]\r", color.GreenLight)
			} else if download == "1" {
				PushInSaving(page.Value.(models.NormalImageInfo))
			}
		}
		pageCount++
	}
}

func (this *CrawlController) CrawlMovie() {

	go ProcessMovie()
	go StoringMovie()

	index := models.C_UN_INDEX_URL
	days := 1
	if this.GetString("days") != "" {
		_days, _ := strconv.Atoi(this.GetString("days"))
		days = _days
	}
	today, _ := models.GetToday()
	endDate := models.DaysAgo(today, days)
	for infos, count := models.GetPageContentsBeforeDate(index, endDate); count != 0 && index != ""; index = models.GetNextPage(index) {
		for _, info := range infos {
			PushInContent(info)
			PushInImage(info)
		}
		infos, count = models.GetPageContentsBeforeDate(index, endDate)
	}
	
	for {
		if IsAllDone() {
			processmoviedone <- true
			storemoviedone <- true
			close(page)
			close(magnets)
			close(images)
			close(datas)
			break
		}
	}
	logs.Trace("[End of crawl]")
	this.Ctx.Redirect(302, "/?type=magnet")
}

func (this *CrawlController) CrawlImage() {
	
	go ProcessImage()
	go StoringImage(RequestId)
	RequestId++

	var (
		index 	 = ""
		days 	 = ""
		download = ""
		local    = ""
	)
	
	indexType = this.GetString("type")
	days = this.GetString("days")
	download = this.GetString("download")
	local = this.GetString("local")

	if indexType == models.TABLE_NEW_TIMES {
		index = models.NEW_TIMES_URL
	} else if indexType == models.TABLE_FIRE {
		index = models.FIRE_URL
	}

	pageCount := 1
	if local == "1" {
		var pageIndex = list.New()
		file, err := os.Open(models.NEW_TIMES_PAGE_LIST)
		if err != nil {
			panic(err)
		}
    	defer file.Close()

    	read := bufio.NewReader(file)
		for {
			line, err := read.ReadString('\n') 
			
			if err != nil || io.EOF == err {
				break
			}
			pageIndex.PushBack(strings.Trim(line, "\n"))
		}
		ParseImageIndex(download, pageIndex)
	} else {
		for index != "" {
			var pages = list.New()
			if days != "" {
				_days, _ := strconv.Atoi(days)
				pages = models.GetImageListBeforeDays(index, _days)
			} else {
				pages = models.GetImageList(index)
			}
			
			color.ColorLog.Logs(
				models.GetDateTime(), 				   color.MagentaLight, 
				" Parsing Index ", 	  				   color.CyanBold, 
				models.ReplaceWithSymbolsEqual(index), color.GrayLight,
				" Page ", 	  						   color.CyanBold, 
				fmt.Sprintf("%d", pageCount), 	  	   color.RedBold,
				" Items ", 	  					  	   color.CyanBold, 
				fmt.Sprintf("%d\n", pages.Len()), 	   color.RedBold)
			for page := pages.Front(); nil != page; page = page.Next() {
				// logs.Trace("[Parsing Page] ", page.Value.(models.NormalImageInfo).Url)
				if models.IsNewTimesVisit(page.Value.(models.NormalImageInfo).Url) {
					continue
				}
				models.AddToNewTimesVisitedSet(page.Value.(models.NormalImageInfo).Url)
				if download == "0" {
					PushInSrcs(page.Value.(models.NormalImageInfo))
					color.ColorLog.Log(" PUSHING\r", color.GreenLight)
				} else if download == "1" {
					PushInSaving(page.Value.(models.NormalImageInfo))
				}
			}
	
			if pages.Len() == 0 {
				break;
			}
			pageCount++
			index = models.GetImageNextPage(index)
		}
	}
	
	for {
		if IsAllDone() {
			processmoviedone <- true
			storemoviedone <- true
			close(saving)
			close(srcs)
			close(loading)
			close(storing)
			break
		}
	}
	logs.Trace("[End of crawl]")
	this.Ctx.Redirect(302, "/?type=magnet")
}

func (this *CrawlController) ProcessUnsuccessPages() {
	go ProcessUnImage()
	go StoringUnImage()

	indexType = this.GetString("type")
	unsuccessPages := models.GetUnsuccessPages(indexType)

	for page := unsuccessPages.Front(); nil != page; page = page.Next() {
		imageSrcs := models.GetImageList(page.Value.(string))
		for src := imageSrcs.Front(); nil != src; src = src.Next() {
			PushInUnsrcs(src.Value.(models.NormalImageInfo))
		}
		models.AddToNewTimesVisitedSet(page.Value.(string))
		models.RemoveItemFromRedis("CONTENT_UNSUCCESS_SET", page.Value.(string))
	}

	for {
		if IsAllDone() {
			processunimagedone <- true
			storeunimagedone <- true
			close(unsrcs)
			close(unstoring)
			break
		}
	}
	logs.Trace("[End of crawl]")
	this.Ctx.Redirect(302, "/")
}

func (this *CrawlController) CrawlImagSavePages() {
	var (
		index 	 = ""
		save	 = ""
		filename = ""
	)
	
	indexType = this.GetString("type")
	save = this.GetString("save")

	if indexType == models.TABLE_NEW_TIMES {
		filename = models.NEW_TIMES_PAGE_LIST
	} 

	var pageIndex = list.New()
	if save == "1" {
		var file *os.File
		if !models.IsFileExist(filename) { 
			file, _ = os.Create(filename)
		} 
	
		io.WriteString(file, index + "\n")
		for index != models.FORUM_INDEX {
			index = models.GetImageNextPage(index)
			io.WriteString(file, index + "\n")
			pageIndex.PushBack(index)
		}
	}
	this.Ctx.Redirect(302, "/")
}

func (this *CrawlController) CrawlImageConcurrent() {
	var filename string
	indexType = this.GetString("type")

	if indexType == models.TABLE_NEW_TIMES {
		filename = models.NEW_TIMES_PAGE_LIST
	} 

	var indices = list.New()
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
    defer file.Close()
    read := bufio.NewReader(file)
	for {
		line, err := read.ReadString('\n') 
		
		if err != nil || io.EOF == err {
			break
		}
		indices.PushBack(strings.Trim(line, "\n"))
	}

	e := ConcurrentEngine {
		Scheduler:   &QueueScheduler{},
		WorkerCount: 100,
		PageProcessor: models.GetNormalImagePageConentWithoutPath,
		SrcProcessor: models.AddImageToDatabaseDirectlyConcurrently,
	}

	pageCount := 1
	for p := indices.Front(); nil != p; p = p.Next() {
		pages := models.GetImageList(p.Value.(string))
		color.ColorLog.Logs(
			models.GetDateTime(), 				   color.MagentaLight, 
			" Parsing Index ", 	  				   color.CyanBold, 
			models.ReplaceWithSymbolsEqual(p.Value.(string)), color.GrayLight,
			" Page ", 	  						   color.CyanBold, 
			fmt.Sprintf("%d", pageCount), 	  	   color.RedBold,
			" Items ", 	  					  	   color.CyanBold, 
			fmt.Sprintf("%d\n", pages.Len()), 	   color.RedBold,
		)
		
		e.Run(pages)
		pageCount++
	}
	
}
