package controllers


import (
	"time"
	"Spider/models"
	"github.com/astaxie/beego"
)

var (
	A = make(chan string, 100)
	B = make(chan string, 100)
	C = make(chan string, 100)
	D = make(chan string, 100)
)

var _count int64 = 0
var _start int64 = 0

type PrintController struct {
	beego.Controller
}

func IsChannelEmpty() bool {
	if len(A) != 0 {
		return false
	} else if len(B) != 0 {
		return false
	} else if len(C) != 0 {
		return false
	} else if len(D) != 0 {
		return false
	}
	return true
}

func PushInA(a string) {
	A <- a
}

func PushInB(b string) {
	B <- b
}

func PushInC(c string) {
	C <- c
}

func PushInD(d string) {
	D <- d
}

func doSomthing() {
	for {
		select {
		case a := <- A:
			models.ColorPrint(models.GetStatus(&_count), "[AAA" + a + "A...]", models.CountDown(_start), 11)
			PushInB("B")
			PushInC("C")
			time.Sleep(1e8)
		case b := <- B:
			models.ColorPrint(models.GetStatus(&_count), "[BBBB" + b + "...]", models.CountDown(_start), 12)
			time.Sleep(1e8)
		case c := <- C:
			models.ColorPrint(models.GetStatus(&_count), "[CCCC" + c + "...]", models.CountDown(_start), 1)
			PushInD("D")
			time.Sleep(1e8)
		case d := <- D:
			models.ColorPrint(models.GetStatus(&_count), "[DDDD" + d + "...]", models.CountDown(_start), 13)
			time.Sleep(1e8)
		}
	}
}

func (this *PrintController) Print() {
	//models.CreateImageTable("test")
	for i := 0; i < 50; i++ {
		PushInA("A")
		time.Sleep(1e8)
	}
	for {
		if IsChannelEmpty() {
			this.Ctx.Redirect(302, "/?type=magnet")
			break
		}
	}
}

func init() {
	_start = time.Now().UnixNano()
	go doSomthing()
}