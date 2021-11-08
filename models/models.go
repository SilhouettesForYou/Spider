package models

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"database/sql"
	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
	"github.com/PuerkitoBio/goquery"
	"golang.org/x/text/encoding/simplifiedchinese"
	"github.com/astaxie/beego/logs"

	"Spider/color"
)

type MovieInfo struct {
	Actress  	 string
	Name 		 string
	Number   	 string
	ProductTime  string
	DownloadLink string
}

type ListInfo struct {
	Title  string
	Number string
	Date   string
	Page   string
}

type MagnetInfo struct {
	Time   string
	Size   string
	Magnet string
	Hash   string
}

type ImageInfo struct {
	Number string
	Path   string
}

type NormalImageInfo struct {
	Title string
	Url   string
}

type NormalImageInfoConcurrently struct {
	Title string
	Url   string
}

type NormalImageDetail struct {
	Id 	  int
	Title string
	Src   string
	Path  string
	Table string
}

type MovieHeaven struct {
	TranslateName string
	MovieName	  string
	Year		  string
	Country		  string
	Category	  string
	Language 	  string
	Subtitle	  string
	Url			  string
	Poster		  string
	Image		  string
	Magnet		  string
}

type NormalImageData struct {
	Id int64
	Title string
	Path string
	Data []byte
}

type NormalImageWithoutData struct {
	Id int64
	Title string
	Path string
}

type MagnetData struct {
	Id int64 `json:"id"`
	Number string `json:"no"`
	Time string `json:"time"`
	Size string `json:"size"`
	Magnet string `json:"magnet"`
	Hash string `json:"hash"`
}

type NormalInfoNoPic struct {
	Title string
	Path string
}

type Charset string

const (
	UTF8    = Charset("UTF-8")
	GB18030 = Charset("GB18030")
)

const (
	TABLE_NEW_TIMES 	   = "newtimes"
	TABLE_NEW_TIMES_NO_PIC = "newtimesnopic"
	TABLE_FIRE 	 		   = "fire"
	TABLE_FIRE_NO_PIC 	   = "firenopic"
	TABLE_MAGNET		   = "magnet"
)

type ContentIndex int32

const (
	NAME  int16 = 1
	ACTOR int16 = 2
	SIZE  int16 = 4
)

const (
	C_INDEX_URL    = "https://cc.spdh.icu"
	C_UN_INDEX_URL = "https://cc.spdh.icu/thread0806.php?fid=2"
	M_INDEX_URL    = "http://dog.magnet404.xyz"
	M_SEARCH_URL   = "http://dog.magnet404.xyz/search/"
	DOWNLOAD_INDEX = "http://rmdown.com/download.php?"
	FORUM_INDEX	   = "http://www.1899bbs.com/"
	NEW_TIMES_URL  = FORUM_INDEX + "99bbs-78-1.html"
	FIRE_URL	   = FORUM_INDEX + "99bbs-53-1.html"
	MOIVE_HEAVEN_URL = "https://www.dytt8.net"
	MOIVE_HEAVEN_INDEX = "/index0.html"
)

const (
	IMG_DIR = "./static/img/"
	MOVIE_IMAGES_DIR = "./static/img/movie/"
	NEW_TIMES_PAGE_LIST = "./static/img/newtimes.txt"
)

const (
	_TranslateName int = 0 + 1
	_MovieName	   int = 1 + 1
	_Year		   int = 2 + 1
	_Country	   int = 3 + 1
	_Category	   int = 4 + 1
	_Language 	   int = 5 + 1
	_Subtitle	   int = 6 + 1
	_Poster	
	_Image	  
	_Magnet		  
)

var MaxItemCount int = 5000

var (
	NewTimesTableCount int = 0
)

var RegExp = [...]string{ `[A-Za-z]+-[a-z]+[-| ][a-z]+[-]?[\d]+`,
						  `[\w]+[ ]*[\d]*[ |_|-][\w]+[[-]?[\d]+]?` }

type Clear func()
var SetsOperation = map[string]Clear {
	"cs":  ClearContentSuccess,
	"cus": ClearContentUnsuccess,
	"cv":  ClearContentVisited,
	"ms":  ClearMagnetSuccess,
	"cms": ClearMagnetSuccess,
	"mv":  ClearMagnetVisited,
	"hv":  ClearHeavenVisited,
	"ntv": ClearNewTimesVisited,
}

var (
	NewTimesCapcity = make(map[int]int)
	NewTimesCountIncrease = make(map[int]int)
	NewTimesCapcityNoPic = make(map[int]int)
	NewTimesCountIncreaseNoPic = make(map[int]int)
	FireCapcity = make(map[int]int)
	FireCountIncrease = make(map[int]int)
	FireCapcityNoPic = make(map[int]int)
	FireCountIncreaseNoPic = make(map[int]int)
	MagnetCapcity = make(map[int]int)
	MagnetCountIncrease = make(map[int]int)
)

var db *sql.DB

func init() {
	orm.Debug = true 
	db, _ = sql.Open("mysql", "root:123456@tcp(127.0.0.1:3306)/crawl?charsetutf8")
}

func DecodeToGBK(text string) (string, error) {

	dst := make([]byte, len(text) * 2)
	tr := simplifiedchinese.GB18030.NewDecoder()
	nDst, _, err := tr.Transform(dst, []byte(text), true)
	if err != nil {
		return text, err
	}

	return string(dst[:nDst]), nil
}

func ConvertByte2String(byte []byte, charset Charset) string {

	var str string
	switch charset {
	case GB18030:
		var decodeBytes, _ = simplifiedchinese.GB18030.NewDecoder().Bytes(byte)
		str= string(decodeBytes)
	case UTF8:
		fallthrough
	default:
		str = string(byte)
	}

	return str
}

func RequestWithHeaders(url string) (*goquery.Document, error) {
	// logs.Trace("[Requesting] ", url)
	// client := new(http.Client)
	reqest, err := http.NewRequest("GET", url, nil)

	// timeout
	client := http.Client {
		Transport: &http.Transport {
			Dial: func(netw, addr string) (net.Conn, error) {
			deadline := time.Now().Add(30 * time.Second)
			c, err := net.DialTimeout(netw, addr, time.Second * 30)
			if err != nil {
				logs.Trace("[TIME OUT] requesting  ▶ ", url)
				return nil, err
			}
			c.SetDeadline(deadline)
				return c, nil
			},
		},
	}

    reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:69.0) Gecko/20100101 Firefox/69.0")
	reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	reqest.Header.Add("Upgrade-Insecure-Requests", "1")

	response, err := client.Do(reqest)
	doc, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		logs.Trace("[ERROR] request error: ", err, " ▶ ", url)
	}
	return doc, err
}

func IsFileExist(fileName string) bool {
	_, err := os.Stat(fileName)
	if os.IsNotExist(err) {
		return false
	}

	os.Remove(fileName)
	return false
}

func DownloadFile(url string, localPath string) error {
	var (
		buf = make([]byte, 32 * 1024)
		written int64
	)

	tempFilePath := localPath + ".download"

	client := new(http.Client)
	// client.TimeOut = time.Second * 60
	req, err := http.NewRequest("GET", url, nil)

    req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:69.0) Gecko/20100101 Firefox/69.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return err
	}

	IsFileExist(localPath)

	file, err := os.Create(tempFilePath)
	if err != nil {
		return err
	}

	defer file.Close()
	if resp.Body == nil {
		return errors.New("body is null")
	}

	defer resp.Body.Close()

	for {
		// read bytes
		read, errorRead := resp.Body.Read(buf)
		if read > 0 {
			// write bytes
			write, errorWrite := file.Write(buf[0:read])
			if write > 0 {
				written += int64(write)
			}

			if errorWrite != nil {
				err = errorWrite
				break
			}

			if read != write {
				err = io.ErrShortWrite
				break
			}
		}
		if errorRead != nil {
			if errorRead != io.EOF {
				err = errorRead
			}
			break
		}
		//fb(fsize, written)
	}
	
	if err == nil {
		file.Close()
		err = os.Rename(tempFilePath, localPath)
	}
	return err
}

func ResouceToBlob(url string) ([]byte, error) {
	var (
		buf = make([]byte, 32 * 1024)
		blob []byte
	)

	client := new(http.Client)
	// client.TimeOut = time.Second * 60
	req, err := http.NewRequest("GET", url, nil)

    req.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:69.0) Gecko/20100101 Firefox/69.0")
	req.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	req.Header.Add("Upgrade-Insecure-Requests", "1")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.Body == nil {
		return nil, errors.New("body is null")
	}

	defer resp.Body.Close()

	for {
		// read bytes
		read, errorRead := resp.Body.Read(buf)
		if read > 0 {
			// append bytes
			temp := [][]byte{ blob, buf[0:read] }
			blob = bytes.Join(temp, []byte{})
		}
		if errorRead != nil {
			if errorRead != io.EOF {
				err = errorRead
			}
			break
		}
		color.ColorLog.Log(fmt.Sprintf(" %s\r", SizeToString(len(blob))), color.YellowBold)
	}
	
	return blob, err
}

// next page
func GetNextPage(url string) string {
	next := ""
	doc, _ := RequestWithHeaders(url)
	pageField := doc.Find(".pages")
	pageField.Find("a").Each(func(i int, s *goquery.Selection) {
		text := ConvertByte2String([]byte(s.Text()), GB18030)
		reg := regexp.MustCompile(`下`)
		content := reg.FindString(text)
		if len(content) != 0 {
			next, _ = s.Attr("href")
			return
		}
	})
	if strings.Index(next, "#") != -1 {
		return ""
	} else {
		return C_INDEX_URL + "/" + next
	}
}

// previous page
func GetPrevPage(url string) string {
	prev := ""
	doc, _ := RequestWithHeaders(url)
	pageField := doc.Find(".pages")
	pageField.Find("a").Each(func(i int, s *goquery.Selection) {
		text := ConvertByte2String([]byte(s.Text()), GB18030)
		reg := regexp.MustCompile(`上`)
		content := reg.FindString(text)
		if len(content) != 0 {
			prev, _ = s.Attr("href")
			return
		}
	})
	if strings.Index(prev, "#") != -1 {
		return ""
	} else {
		return C_INDEX_URL + "/" + prev
	}
}

func ParseNumber(url string, title string) string {
	var no = ""
	text := ConvertByte2String([]byte(title), GB18030)
	for _, regExp := range RegExp {
		reg := regexp.MustCompile(regExp)
		content := reg.FindAllString(text, -1)
		if len(content) != 0 {
			// logs.Trace("[No.] ", content)
			no = content[0]
			break;
		} else {
			AddToContentUnsuccess(url)
		}
	}
	return no
}

// page list
func GetPageContents(url string) map[int]ListInfo {
	var infos = make(map[int]ListInfo)
	doc, err := RequestWithHeaders(url)
	if nil != err {
		logs.Trace("[Error] ", err, "▶", url)
		return infos
	}
	index := 0
	doc.Find("td[class=tal]").Each(func(i int, table *goquery.Selection) {
		title := table.Find("h3").Find("a").Text()
		page, _ := table.Find("h3").Find("a").Attr("href")
		if !UrlFilter(page) {
			infos[index] = ListInfo{ Title:title, Date:"", Number:ParseNumber(page, title), Page:C_INDEX_URL + "/" + page}
			index++
		}
	})
	index = 0
	doc.Find("td").Each(func(i int, table *goquery.Selection) {
		date := table.Find("a[class=f10]").Text()
		if len(date) > 0 {
			infos[index] = ListInfo{ Title:infos[index].Title, Number:infos[index].Number, Date: date, Page:infos[index].Page }
			index++
		}
	})
	return infos
}

func GetPageContentsBeforeDate(url string, date string) (map[int]ListInfo, int) {
	var infos = make(map[int]ListInfo)
	_infos := GetPageContents(url)
	count := 0
	for _, info := range _infos {
		if CompareDate(info.Date, date) {
			continue
		}
		infos[count] = info
		count++
	}
	return infos, count
}

// parse content of each page
func ParsePageContent(url string) MovieInfo {
	var (
		number 	 string = ""
		create   string = ""
		name     string = ""
		actress  string = ""
		link 	 string = ""
	)
	doc, err := goquery.NewDocument(url)
	if err != nil {
		logs.Trace("[Error] cannot parse page!")
		return MovieInfo{}
	}
	doc.Find("title").Each(func(i int, s *goquery.Selection) {
		title := ConvertByte2String([]byte(s.Text()), GB18030)
		number = strings.Trim(strings.Split(title, " ")[0], "[MP4]")
		// logs.Trace("[No.] ", ReplaceWithSymbols(number))
		// logs.Trace("[Short] ", ReplaceWithSymbols(typeName))
	})

	doc.Find(".tpc_detail").Each(func(i int, s *goquery.Selection) {
		text := ConvertByte2String([]byte(s.Text()), GB18030)
		reg := regexp.MustCompile(`\d{4}-\d{1,2}-\d{1,2}\s\d{1,2}:\d{1,2}`)
		res := reg.FindAllString(text, -1)
		if len(res) > 0 && i == 0 {
			create = res[0]
			// logs.Trace("[Create Time] ", res[0])
		}
	})
	// extract some base information
	doc.Find(".tpc_cont").Each(func(i int, s *goquery.Selection) {
		text := ConvertByte2String([]byte(s.Text()), GB18030)
		reg := regexp.MustCompile(`【.*?】`)
		content := reg.FindAllString(text, -1)
		if len(content) != 0 {
			// fmt.Println(content[NAME])
			start := strings.Index(text, content[NAME])
			end := strings.Index(text, content[NAME + 1])
			name = text[start + len(content[NAME]):end]
			name = strings.Trim(name, " ")
			// logs.Trace("[Movie Name] ", ReplaceWithSymbols(name))
			
			// fmt.Println(content[ACTOR])
			start = strings.Index(text, content[ACTOR])
			end = strings.Index(text, content[ACTOR + 1])
			actress = text[start + len(content[ACTOR]):end]
			actress = strings.Trim(actress, "：")
			// logs.Trace("[Actress] ", ReplaceWithSymbols(actress))
		}
	})

	doc.Find("a").Each(func(i int, s *goquery.Selection) {
		l, _ := s.Attr("href")
		if len(l) > 0 {
			reg := regexp.MustCompile(`(.*rmdown.*)`)
			res := reg.FindAllString(l, -1)
			if len(res) > 0 {
				link = s.Text()
				// logs.Trace("[Download Link] ", link)
				// ParseDownloadContent(downloadLink)
			}
		}
	})

	return MovieInfo{ Actress:actress, Name:name, Number:number, ProductTime:create, DownloadLink:link }
}

// parse content of download content
func ParseDownloadContent(url string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Println("cannot parse page!")
		log.Fatal(err)
	}
	link := DOWNLOAD_INDEX
	doc.Find("input").Each(func(i int, s *goquery.Selection) {
		input, _ := s.Attr("name")
		if len(input) > 0 {
			value, _ := s.Attr("value")
			if input == "reff" {
				link += "reff=" + value
			}
			if input == "ref" {
				link += "&ref=" + value
			}
		}
	})
	logs.Trace("[File url] ", link)
	//DownloadFile(link, "D:\\Downloads\\test.torrent")
}

func GetMagnetsPages(number string) *list.List {
	var pages = list.New()
	var totalPage = 1
	var schema = ""
	doc, err := RequestWithHeaders(M_SEARCH_URL + number)
	if nil != err {
		return pages
	}
	// get last page index
	doc.Find(".pagination").Each(func(i int, s *goquery.Selection) {
		page := s.Find("li")
		schema, _ = page.Last().Find("a").Attr("href")
		splits := strings.Split(schema, "/")
		totalPage, _ = strconv.Atoi(splits[len(splits) - 1])
		schema = strings.Trim(schema, splits[len(splits) - 1])[1:]
	})
	for i := 1; i <= totalPage; i++ {
		pages.PushBack(M_INDEX_URL + "/" + schema + strconv.Itoa(i))
	}
	return pages
}

func GetMagnetsOnPage(page string) *list.List {
	var magnets = list.New()
	doc, err := RequestWithHeaders(page)
	if nil != err {
		logs.Trace("[ERROR] cannot find the magnet ", err, "▶ ", page)
		return magnets
	}
	doc.Find(".panel").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Find("a").Attr("href")
		content, _  := RequestWithHeaders(M_INDEX_URL + link)
		/* parse page to get more information */
		// time & size
		var time = ""
		var size = ""
		content.Find(".col-xs-4").Each(func(i int, info *goquery.Selection) {
			html, _ := info.Html()
			reg := regexp.MustCompile(`创建时间`)
			res := reg.FindAllString(html, -1)
			if len(res) > 0 {
				time = info.Find("span").Text()
			}
			reg = regexp.MustCompile(`文件大小`)
			res = reg.FindAllString(html, -1)
			if len(res) > 0 {
				size = info.Find("span").Text()
			}
		})
		// hash
		var hash = ""
		content.Find(".col-md-12").Each(func(i int, info *goquery.Selection) {
			html, _ := info.Html()
			reg := regexp.MustCompile(`哈希`)
			res := reg.FindAllString(html, -1)
			if len(res) > 0 {
				hash = info.Find("span").Text()
			}
		})
		// magnet
		var magnet = ""
		content.Find("textarea").Each(func(i int, info *goquery.Selection) {
			magnet = info.Text()
		})
		magnets.PushBack(MagnetInfo{Time:time, Size:size, Magnet:magnet, Hash:hash})
	})
	return magnets
}

func GetImagesOnPage(page string, title string) map[string]string {
	var images = make(map[string]string)
	doc, err := RequestWithHeaders(page)
	if nil != err {
		logs.Trace("[ERROR] cannot find the movie image ", err, "▶ ", page)
		return images
	}

	// get directory name from title
	// GenerateDirectoryName(ConvertByte2String([]byte(title), GB18030))
	path := MOVIE_IMAGES_DIR + GenerateDirectoryName(ConvertByte2String([]byte(title), GB18030)) + "/"
	MakeDir(path)

	index := 1
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		src, _ := s.Attr("data-src")
		if len(src) != 0 && strings.Index(src, ".jpg") != -1 {
			src = strings.Replace(src, "th.jpg", "jpg", -1)
			
			_, ok := images[src]
			if !ok {
				images[src] = path + strconv.Itoa(index) + ".jpg"
				index++
			}
		}
	})
	return images
}

/***********************
 *  Movie Heaven Zone  *
 ***********************/

func GetHeavenList(url string) *list.List {
	var pages = list.New()
	doc, err := RequestWithHeaders(url)
	if nil != err {
		logs.Trace("[ERROR] cannot get the heaven list ", err, "▶ ", url)
		return pages
	}
	doc.Find(".co_content2").Each(func(i int, s *goquery.Selection) {
		img, _ := s.Find("img").Attr("src")
		if img == "" {
			s.Find("a").Each(func(j int, l *goquery.Selection) {
				// name := ConvertByte2String([]byte(l.Text()), GB18030)
				link, _ := l.Attr("href")
				pages.PushBack(MOIVE_HEAVEN_URL + link)
			})
		}
	})
	pages.Remove(pages.Front())
	return pages
}

func GetHeavenContent(page string) MovieHeaven {
	var splits []string
	var image string
	var poster string
	doc, err := RequestWithHeaders(page)
	if nil != err {
		logs.Trace("[ERROR] cannot parse the heaven content", err, "▶ ", page)
	}
	content := doc.Find("#Zoom")
	content.Find("p").Each(func(i int, s *goquery.Selection) {
		text := ConvertByte2String([]byte(s.Text()), GB18030)
		re := regexp.MustCompile(`◎`)
		if re.MatchString(text) {
			s.Find("img").Each(func(j int, img *goquery.Selection) {
				if j == 0 {
					image, _ = img.Attr("src")
				} 
				if j == 1 {
					poster, _ = img.Attr("src")
				}
			})
			splits = strings.Split(text, "◎")
			return
		}
	})

	if len(splits) == 0 {
		content.Find("span").Each(func(i int, s *goquery.Selection) {
			text := ConvertByte2String([]byte(s.Text()), GB18030)
			re := regexp.MustCompile(`◎`)
			if re.MatchString(text) {
				s.Find("img").Each(func(j int, img *goquery.Selection) {
					if j == 0 {
						image, _ = img.Attr("src")
					} 
					if j == 1 {
						poster, _ = img.Attr("src")
					}
				})
				splits = strings.Split(text, "◎")
				return
			}
		})
	}

	var thunder string
	content.Find("a").Each(func(i int, s *goquery.Selection) {
		html, _ := s.Html()
		// reg := regexp.MustCompile(`thunder://[\w]+`)
		reg := regexp.MustCompile(`ftp:\/\/(.*)\.[\w]{3,}$`)
		if reg.MatchString(ConvertByte2String([]byte(html), GB18030)) {
			thunder = reg.FindString(ConvertByte2String([]byte(html), GB18030))
		}
	})
	
	if len(splits) == 0 {
		return MovieHeaven{}
	}
	return MovieHeaven{ 
		TranslateName:splits[_TranslateName],
		MovieName:splits[_MovieName],
		Year:splits[_Year],
		Country:splits[_Country],
		Category:splits[_Category],
		Language:splits[_Language],
		Subtitle:splits[_Subtitle],
		Url: page,
		Poster:poster,
		Image:image,
		Magnet:thunder }
}

func AddMovieHeavenToDatabase(movie MovieHeaven) (int64, error) {

	if IsMovieHeavenVisit(movie.Magnet) {
		return 0, errors.New("Magnet is existed")
	}
	AddToMovieHeavenVisitedSet(movie.Magnet)

	poster, err := ResouceToBlob(movie.Poster)
	image, err := ResouceToBlob(movie.Image)
	if nil != err {
		return 0, err
	}
	result, err := db.Exec("INSERT INTO heaven" +
		"(tanslatename, moviename, year, country, category, language, subtitle, url, poster, image, magnet)" +
		"VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)", 
		movie.TranslateName, movie.MovieName, movie.Year, movie.Country, movie.Category, movie.Language, movie.Subtitle, movie.Url, poster, image, movie.Magnet)
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	return id, err
}

/**************** 
 *  Image Zone  *
 ****************/
 
func GetImageNextPage(url string) string {
	doc, err := RequestWithHeaders(url)
	if nil != err {
		logs.Trace("[ERROR] cannot get next page ▶ ", url)
	}
	s := doc.Find(".pg").Find(".nxt")
	if nil == s {
		return ""
	}
	next, _ := s.Attr("href")
	return FORUM_INDEX + next
}

func GetImageList(url string) *list.List {
	var pages = list.New()
	doc, _ := RequestWithHeaders(url)
	field := doc.Find("#threadlist")
	if nil == field {
		return pages
	}
	field.Find("tbody").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		if strings.Index(id, "normalthread") != -1 {
			link := s.Find(".s.xst")
			if nil != link {
				title := link.Text()
				page, _ := link.Attr("href")
				pages.PushBack(NormalImageInfo{ Title:title, Url:FORUM_INDEX + page })
			}
		}
	})
	return pages
}

func GetImageListBeforeDays(url string, days int) *list.List {
	var pages = list.New()
	doc, _ := RequestWithHeaders(url)
	field := doc.Find("#threadlist")
	if nil == field {
		return pages
	}
	field.Find("tbody").Each(func(i int, s *goquery.Selection) {
		id, _ := s.Attr("id")
		if strings.Index(id, "normalthread") != -1 {
			var element *list.Element
			link := s.Find(".s.xst")
			if nil != link {
				title := link.Text()
				page, _ := link.Attr("href")
				element = pages.PushBack(NormalImageInfo{ Title:title, Url:FORUM_INDEX + page })
			}
			date := s.Find("span").Text()
			if DateFilter(date) {
				if !IsWithinDays(date, days) && nil != element {
					pages.Remove(element)
				}
			}
		}
	})
	return pages
}


// parse page from unsuccess pages
func GetUnsuccessPages(indexType string) *list.List {
	unsuccessPages := list.New()
	isExisted, _ := client.Exists(indexType)
	if len(indexType) == 0 || !isExisted {
		return unsuccessPages
	}
	items, _ := client.Smembers(indexType)
	for item := range items {
		unsuccessPages.PushBack(string(item))
	}
	return unsuccessPages
}

func IsImageUrlValid(url string) bool {
	client := new(http.Client)
	reqest, _ := http.NewRequest("GET", url, nil)

    reqest.Header.Add("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:69.0) Gecko/20100101 Firefox/69.0")
	reqest.Header.Add("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	reqest.Header.Add("Accept-Language", "zh-CN,zh;q=0.8,zh-TW;q=0.7,zh-HK;q=0.5,en-US;q=0.3,en;q=0.2")
	reqest.Header.Add("Upgrade-Insecure-Requests", "1")

	response, err := client.Do(reqest)

	if  nil != err || 404 == response.StatusCode {
		return false
	}
	return true
}

func GetNormalImagePageContent(info NormalImageInfo, indexType string) map[string]string {
	var images = make(map[string]string)
	doc, err := RequestWithHeaders(info.Url)
	if nil != err {
		logs.Trace("[ERROR] cannot find any images on the page(", indexType, ") ▶ ", info.Url)
		return images
	}

	path := IMG_DIR + indexType + "/" + info.Title + "/"
	MakeDir(path)

	index := 1
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		img := s.Find(".zoom")
		if nil != img {
			src, _ := s.Attr("src")
			if len(src) != 0 && strings.Index(src, ".jpg") != -1 {
				_, ok := images[src]
				if !ok && IsImageUrlValid(src) {
					images[src] = path + strconv.Itoa(index) + ".jpg"
					index++
				}
			}
		}
	})
	return images
}

func GetNormalImagePageConentWithoutPath(info NormalImageInfo, indexType string) *list.List {
	// color.ColorLog.Log(fmt.Sprintf("%s [START PARSING PAGE]                                                                                  \n", GetDateTime()), color.GreenLight)
	var urls = list.New()
	doc, err := RequestWithHeaders(info.Url)
	if nil != err {
		logs.Trace("[ERROR] cannot parse content of image page ▶ ", info.Url)
		RemoveItemFromRedis("NEW_TIMES_VISIT_SET", info.Url)
		AddToContentUnsuccess(info.Url)
		return urls
	}
	path := IMG_DIR + indexType + "/" + info.Title + "/"
	index := 1
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		img := s.Find(".zoom")
		if nil != img {
			src, _ := s.Attr("src")
			if len(src) != 0 && strings.Index(src, ".jpg") != -1 {
				if IsImageUrlValid(src) {
					// fmt.Printf("[Src %s]\r", src)
					urls.PushBack(NormalImageDetail{ 
						Id:index, 
						Title:info.Title,
						Src:src, Path:path + strconv.Itoa(index) + ".jpg",
						Table:indexType,
					})
					index++
				}
			}
		}
	})
	// color.ColorLog.Log(fmt.Sprintf("%s [END PARSING PAGE]                                                                                  \n", GetDateTime()), color.RedLight)
	return urls
}

// add image to database
func AddNormalImageToDatabase(path, table string) (int64, error) {
	// logs.Trace("[Path] ", path)
	buf, err := ReadFileBinary(path)
	splits := strings.Split(path, "/")
	title := splits[len(splits) - 2]
	if nil != err {
		return 0, err
	}
	if table == "" {
		return 0, errors.New("table name is nil")
	}
	result, err := db.Exec("INSERT INTO " + table +
		"(title, path, data)" +
		"VALUES(?, ?, ?)", 
		title, path, buf)
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	return id, err
}

func AddNormalImageToDatabaseDirectly(count, chNum, requestId int, chName, title, url, path, table string) (int64, error) {
	// color.New(color.FgYellow, color.Bold).Printf("[Storing image...]\r")
	NewTimesTableCount++
	color.ColorLog.Logs(
		"           " + GetDateTime() + " ", 	 color.MagentaLight, 
		"#", 					 				 color.CyanBold, 
		"STORING IMAGES ", 					 	 color.NormalBold, 
		fmt.Sprintf("no.%3d", count),		 	 color.MagentaBold, 
		" | ", 					 				 color.RedBold, 
		"#", 					 				 color.CyanBold, 
		"TOTAL COUNT", 						 	 color.NormalBold, 
		fmt.Sprintf("%6d", NewTimesTableCount),  color.MagentaBold,
		" | ", 					 				 color.RedBold, 
		"#", 					 				 color.CyanBold, 
		"CHANNEL NUM",	 					 	 color.NormalBold,
		fmt.Sprintf(" %s %3d", chName, chNum),	 color.MagentaBold,
		" | ", 					 				 color.RedBold, 
		"#", 					 				 color.CyanBold, 
		"REQUEST ID ", 		 			 	 	 color.NormalBold,
		fmt.Sprintf("%2d\r", requestId),		 color.MagentaBold)
	buf, err := ResouceToBlob(url)
	if nil != err {
		return 0, err
	}
	if table == "" {
		return 0, errors.New("table name is nil")
	}
	result, err := db.Exec("INSERT INTO " + table +
		"(title, path, data)" +
		"VALUES(?, ?, ?)", 
		title, path, buf)
	if nil != err {
		AddToStoreFailedImageSet(url)
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	return id, err
}

func AddImageToDatabaseDirectlyConcurrently(image NormalImageDetail) (int64, error) {
	NewTimesTableCount++
	color.ColorLog.Logs(
		"           " + GetDateTime() + " ", 	color.MagentaLight,  
		"STORING IMAGES", 					 	color.NormalBold, 
		fmt.Sprintf("%3d", image.Id),		 	color.MagentaBold, 
		" | ", 					 				color.RedBold, 
		"TOTAL COUNT ", 						color.NormalBold, 
		fmt.Sprintf("%6d", NewTimesTableCount), color.MagentaBold, 
		" | \r", 					 			color.RedBold)
	buf, err := ResouceToBlob(image.Src)
	if nil != err {
		return 0, err
	}
	if image.Table == "" {
		return 0, errors.New("table name is nil")
	}
	result, err := db.Exec("INSERT INTO " + image.Table +
		"(title, path, data)" +
		"VALUES(?, ?, ?)", 
		image.Title, image.Path, buf)
	if nil != err {
		AddToStoreFailedImageSet(image.Src)
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	return id, err
}

func AddNormalImageToDatabaseWithoutPic(title, url, path, table string) (int64, error) {
	if table == "" {
		return 0, errors.New("table name is nil")
	}
	result, err := db.Exec("INSERT INTO " + table +
		"(title, path)" +
		"VALUES(?, ?)", 
		title, path)
	if nil != err {
		AddToStoreFailedImageSet(url)
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	return id, err
}

func AddNormalImageToDatabaseSplittly(title, url, path, table string) (int64, error) {
	var tableName = table
	if NewTimesTableCount > MaxItemCount {
		tableName = table + strconv.Itoa(NewTimesTableCount / MaxItemCount)
		if NewTimesTableCount % MaxItemCount == 1 {
			CreateImageTable(tableName)
		}
	}
	
	color.New(color.FgYellow, color.Bold).Printf("[Storing image...]\r")
	buf, err := ResouceToBlob(url)
	if nil != err {
		return 0, err
	}
	if tableName == "" {
		return 0, errors.New("table name is nil")
	}
	result, err := db.Exec("INSERT INTO " + tableName +
		"(title, path, data)" +
		"VALUES(?, ?, ?)", 
		title, path, buf)
	if nil != err {
		AddToStoreFailedImageSet(url)
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	NewTimesTableCount++
	return id, err
}

// add magnet to database
func AddMagnetToDatabase(no string, magnet MagnetInfo) (int64, error) {
	result, err := db.Exec("INSERT INTO magnet" +
		"(no, time, size, magnet, hash)" +
		"VALUES(?, ?, ?, ?, ?)", 
		no, magnet.Time, magnet.Size, magnet.Magnet, magnet.Hash)
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	return id, err
}

// add image to database
func AddImageToDatabase(info ImageInfo) (int64, error) {
	buf, err := ReadFileBinary(info.Path)
	if nil != err {
		return 0, err
	}
	result, err := db.Exec("INSERT INTO image" +
		"(no, path, data)" +
		"VALUES(?, ?, ?)", 
		info.Number, info.Path, buf)
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	id, err := result.LastInsertId()
	if nil != err {
		logs.Trace("[Error] ", err)
		return 0, err
	}
	return id, err
}

// load image from database
func LoadImageFromDatabase(table, path string) {
	if table == "newtimes" || table == "fire" {
		rows, _ := db.Query("SELECT * FROM " + table/* + " LIMIT 1"*/)
		defer rows.Close()
		for rows.Next() {
			var image NormalImageData
			rows.Scan(&image.Id, &image.Title, &image.Path, &image.Data)
			if path != "" {
				splits := strings.Split(image.Path, "/")
				path = path + splits[2] + "\\" + splits[3]
				MakeDir(path)
			} else {
				CreateLoadImagePath(image.Path)
			}
			ioutil.WriteFile(image.Path, image.Data, 0666)
		}
	}
}

// load data from local and save to database
func SaveDatabaseFromLocal(table string) {
	if table == "newtimesnopic" || table == "firenopic" {
		noPicInfos := GetAllDirectories("./static/img/" + table + "/")
		for _, info := range noPicInfos {
			result, err := db.Exec("INSERT INTO " + table +
			"(title, path)" +
			"VALUES(?, ?)", 
			info.Title, info.Path)
			if nil != err {
				logs.Trace("[Error] ", err)
				continue
			}
			_, err = result.LastInsertId()
			if nil != err {
				logs.Trace("[Error] ", err)
				continue
			}
		}
		
	}
}

func LoadItemsFromDatabase(table string, group int, capcity int) *list.List {
	var items = list.New()
	if table == "newtimes" {
		for i := group; i < group + capcity && i < len(NewTimesCountIncrease); i++ {
			sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(NewTimesCountIncrease[i + 1]) + ", 1"
			// logs.Trace("[SQL]", sql)
			rows, _ := db.Query(sql)
			defer rows.Close()
			for  rows.Next() {
				var image NormalImageData
				rows.Scan(&image.Id, &image.Title, &image.Path, &image.Data)
				items.PushBack(image)
			}
		}
	} else if table == "newtimesnopic" {
		for i := group; i < group + capcity && i < len(NewTimesCountIncreaseNoPic); i++ {
			sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(NewTimesCountIncreaseNoPic[i + 1]) + ", 1"
			// logs.Trace("[SQL]", sql)
			rows, _ := db.Query(sql)
			defer rows.Close()
			for  rows.Next() {
				var image NormalImageWithoutData
				rows.Scan(&image.Id, &image.Title, &image.Path)
				items.PushBack(image)
			}
		}
	} else if table == "fire" {
		for i := group; i < group + capcity && i < len(FireCountIncrease); i++ {
			sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(FireCountIncrease[i + 1]) + ", 1"
			// logs.Trace("[SQL]", sql)
			rows, _ := db.Query(sql)
			defer rows.Close()
			for  rows.Next() {
				var image NormalImageData
				rows.Scan(&image.Id, &image.Title, &image.Path, &image.Data)
				items.PushBack(image)
			}
		}
		
	}  else if table == "firenopic" {
		for i := group; i < group + capcity && i < len(NewTimesCountIncreaseNoPic); i++ {
			sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(NewTimesCountIncreaseNoPic[i + 1]) + ", 1"
			// logs.Trace("[SQL]", sql)
			rows, _ := db.Query(sql)
			defer rows.Close()
			for  rows.Next() {
				var image NormalImageWithoutData
				rows.Scan(&image.Id, &image.Title, &image.Path)
				items.PushBack(image)
			}
		}
	} else if table == "magnet" {
		for i := group; i < group + capcity && i < len(MagnetCountIncrease); i++ {
			sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(MagnetCountIncrease[i + 1]) + ", 1"
			// logs.Trace("[SQL]", sql)
			rows, _ := db.Query(sql)
			defer rows.Close()
			for  rows.Next() {
				var magnet MagnetData
				rows.Scan(&magnet.Id, &magnet.Number, &magnet.Time, &magnet.Size, &magnet.Magnet, &magnet.Hash)
				items.PushBack(magnet.Number)
			}
		}
	}
	return items
}

func LoadDetailsFromDatabase(table string, group int) *list.List {
	var images = list.New()
	if table == "newtimes" {
		sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(NewTimesCountIncrease[group + 1]) + ", " + strconv.Itoa(NewTimesCapcity[group + 1])
		// logs.Trace("[SQL]", sql)
		rows, _ := db.Query(sql)
		defer rows.Close()
		for  rows.Next() {
			var image NormalImageData
			rows.Scan(&image.Id, &image.Title, &image.Path, &image.Data)
			images.PushBack(image)
		}
	} else if table == "newtimesnopic" {
		sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(NewTimesCountIncreaseNoPic[group + 1]) + ", " + strconv.Itoa(NewTimesCapcity[group + 1])
		// logs.Trace("[SQL]", sql)
		rows, _ := db.Query(sql)
		defer rows.Close()
		for  rows.Next() {
			var image NormalImageWithoutData
			rows.Scan(&image.Id, &image.Title, &image.Path)
			images.PushBack(image)
		}
	} else if table == "fire" {
		sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(FireCountIncrease[group + 1]) + ", " + strconv.Itoa(FireCapcity[group + 1])
		// logs.Trace("[SQL]", sql)
		rows, _ := db.Query(sql)
		defer rows.Close()
		for  rows.Next() {
			var image NormalImageData
			rows.Scan(&image.Id, &image.Title, &image.Path, &image.Data)
			images.PushBack(image)
		}
	} else if table == "firenopic" {
		sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(FireCountIncreaseNoPic[group + 1]) + ", " + strconv.Itoa(FireCapcity[group + 1])
		// logs.Trace("[SQL]", sql)
		rows, _ := db.Query(sql)
		defer rows.Close()
		for  rows.Next() {
			var image NormalImageWithoutData
			rows.Scan(&image.Id, &image.Title, &image.Path)
			images.PushBack(image)
		}
	} else if table == "magnet" {
		sql := "SELECT * FROM " + table + " LIMIT " + strconv.Itoa(MagnetCountIncrease[group + 1]) + ", " + strconv.Itoa(MagnetCapcity[group + 1])
		// logs.Trace("[SQL]", sql)
		rows, _ := db.Query(sql)
		defer rows.Close()
		for  rows.Next() {
			var magnet MagnetData
			rows.Scan(&magnet.Id, &magnet.Number, &magnet.Time, &magnet.Size, &magnet.Magnet, &magnet.Hash)
			images.PushBack(magnet)
		}
	}
	return images
}

func GetItemsCount(table string, pageCount int) int {
	var num = 0
	if table == "newtimes" || table == "fire" || table == "newtimesnopic" || table == "firenopic" {
		rows, _ := db.Query("SELECT COUNT(DISTINCT title) FROM " + table)
		defer rows.Close()
		if rows.Next() {
			rows.Scan(&num)
		}
	} else if table == "magnet" {
		rows, _ := db.Query("SELECT COUNT(DISTINCT no) FROM " + table)
		defer rows.Close()
		if rows.Next() {
			rows.Scan(&num)
		}
	}
	return int(math.Ceil(float64(num) / float64(pageCount)))
}

func GetItemsTotalCount(table string) int {
	var num = 0
	rows, _ := db.Query("SELECT COUNT(*) FROM " + table)
		defer rows.Close()
		if rows.Next() {
			rows.Scan(&num)
	}
	return num
}

func GetCapcity(table string) {
	type Count struct {
		id int
		count int
	}
	if table == "magnet" {
		// group by title
		group := "SELECT (@i:=@i+1) as id, COUNT(*) as count FROM " + table + ", (SELECT @i:=0) as it GROUP BY no"
		counts, _ := db.Query(group)
		defer counts.Close()
		total := 0
		for counts.Next() {
			var c Count
			counts.Scan(&c.id, &c.count)
			// logs.Trace("[id]", c.id, "[count]", c.count, "[total]", total)
			MagnetCapcity[c.id] = c.count
			MagnetCountIncrease[c.id] = total
			total += c.count
		}
	} else {
		// group by title
		group := "SELECT (@i:=@i+1) as id, COUNT(*) as count FROM " + table + ", (SELECT @i:=0) as it GROUP BY title"
		counts, _ := db.Query(group)
		defer counts.Close()
		total := 0
		for counts.Next() {
			var c Count
			counts.Scan(&c.id, &c.count)
			// logs.Trace("[id]", c.id, "[count]", c.count, "[total]", total)
			if table == "newtimes" {
				NewTimesCapcity[c.id] = c.count
				NewTimesCountIncrease[c.id] = total
				total += c.count
			} else if table == "newtimesnopic" {
				NewTimesCapcityNoPic[c.id] = c.count
				NewTimesCountIncreaseNoPic[c.id] = total
				total += c.count
			} else if table == "fire" {
				FireCapcity[c.id] = c.count
				FireCountIncrease[c.id] = total
				total += c.count
			} else if table == "firenopic" {
				FireCapcityNoPic[c.id] = c.count
				FireCountIncreaseNoPic[c.id] = total
				total += c.count
			} 
		}
	}
}
