package controllers

import (
	"container/list"

	"Spider/models"
	"Spider/color"
)

type PageProcessor func(models.NormalImageInfo, string) *list.List
type SrcProcessor func(models.NormalImageDetail) (int64, error)

type Request struct {
	Type string
	NormalImageInfo models.NormalImageInfo
	NormalImageDetail models.NormalImageDetail
}

type ImageItem struct {
	Id   string 
	Url  string 
	Size string
}

type ParseResult struct {
	Requests *list.List
	StoreItms *list.List
}

type ConcurrentEngine struct {
	Scheduler 	  Scheduler
	WorkerCount   int
	PageProcessor PageProcessor
	SrcProcessor  SrcProcessor
}

type Scheduler interface {
	WorkerReady(chan Request)
	Submit(request Request)
	CreateWorkerChannel() chan Request
	Run()
}

func (this *ConcurrentEngine) Run(pages *list.List) {
	out := make(chan ParseResult)

	this.Scheduler.Run()

	for i := 0; i < this.WorkerCount; i++ {
		this.CreateWorker(this.Scheduler.CreateWorkerChannel(), out, this.Scheduler)
	}


	for page := pages.Front(); nil != page; page = page.Next() {
		// logs.Trace("[Parsing Page] ", page.Value.(models.NormalImageInfo).Url)
		if models.IsNewTimesVisit(page.Value.(models.NormalImageInfo).Url) {
			color.ColorLog.Log(" [SKIPPING]\r", color.YellowLight)
			continue
		}
		models.AddToNewTimesVisitedSet(page.Value.(models.NormalImageInfo).Url)
		this.Scheduler.Submit(Request{
			// Page: page.Value.(models.NormalImageInfo).Url,
			Type: indexType,
			NormalImageInfo: page.Value.(models.NormalImageInfo),
		})
	}

	for {
		results := <- out

		for src := results.Requests.Front(); nil != src; src = src.Next() {
			this.SrcProcessor(src.Value.(models.NormalImageDetail))
		}
	}
}

func (this *ConcurrentEngine) CreateWorker(in chan Request, out chan ParseResult, s Scheduler) {
	go func() {
		for {
			s.WorkerReady(in)
			request := <- in
			srcs := this.PageProcessor(request.NormalImageInfo, request.Type)
			out <- ParseResult{ Requests: srcs }
		}
	}()
}
