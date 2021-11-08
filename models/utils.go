package models


import (
    "fmt"
    "io"
    "io/ioutil"
    "os"
    "regexp"
    "strconv"
    "strings"
    "syscall"
    "time"

    "github.com/astaxie/beego/logs"
)

type Status string

const (
    One   = Status("\\")
    Two   = Status("-")
    Three = Status("/")
)

func ReplaceWithSymbols(str string) string {
    return string("□□□□□□□")
}

func ReplaceWithSymbolsEqual(str string) string {
    var res = ""
    for i := 0; i < len(str); i++ {
        res += "□"
    }
    return res
}

func GetLastElementOfUrl(url string) string {
    splits := strings.Split(url, "/")
    return splits[len(splits) - 1]
}

func UrlFilter(url string) bool {
    re := regexp.MustCompile(`read\.php\?tid=[\d]*$`)
    return re.MatchString(url)
}

func DateFilter(date string) bool {
    re := regexp.MustCompile(`\d{4}-\d{1,2}-\d{1,2}`)
    return re.MatchString(date)
}
func GetToday() (s string, i int64) {
    return time.Now().Format("2006-01-02 15:04"), time.Now().Unix()
}

func GetTodayDate() string {
    return time.Now().Format("2006-01-02")
}

func GetDateTime() string {
    return time.Now().Format("2006-01-02 15:04:05")
}

func ConvertStringToDate(date string) int64 {
    _date, err := time.ParseInLocation("2006-01-02 15:04", date, time.Local)
    if err == nil {
        return _date.Unix()
    }
    return 0
}

func CompareDate(day1, day2 string) bool {
    a, _ := time.ParseInLocation("2006-01-02 15:04", day1, time.Local)
    b, _ := time.ParseInLocation("2006-01-02 15:04", day2, time.Local)
    return a.After(b)
}

func IsWithinDays(date string, days int) bool {
    timeLayout  := "2016-01-02"
	loc, _ := time.LoadLocation("Local")
	
	startUnix, _ := time.ParseInLocation(timeLayout, GetTodayDate(), loc) 
	endUnix, _ := time.ParseInLocation(timeLayout, date, loc)
	startTime := startUnix.Unix()
	endTime := endUnix.Unix()

	return (endTime - startTime) / 86400 <= int64(days)
}

func ConvertTimeToString(t int64) string {
    return time.Unix(t,0).Format("2006-01-02 15:04")
}

func DaysAgo(date string, days int) string {
    t := ConvertStringToDate(date)
    if t != 0 {
        return ConvertTimeToString(t - int64(days * 24 * 3600))
    }
    return ""
}

func CountDown(start int64) string {
    end := time.Now().UnixNano()
    consume := end - start
    return strconv.FormatInt(consume / 1e9, 10) + "." + strconv.FormatInt(consume % 1e6, 10) + "s"
}

func PathExists(path string) bool {
    _, err := os.Stat(path)
    if err == nil {
        return true
    }
    if os.IsNotExist(err) {
        return false
    }
    return false
}

func FileExisted(filename string) bool {
    var exist = true
    if _, err := os.Stat(filename); os.IsNotExist(err) {
        exist = false
    }
    return exist
}

func MakeDir(dir string) {
    exist := PathExists(dir)
    if !exist {
        err := os.Mkdir(dir, os.ModePerm)
        if err != nil {
            logs.Trace("[Error] ", err)
        }
    }
}

func CreateLoadImagePath(path string) {
    splits := strings.Split(path, "/")
    dir := path[:(len(path) - len(splits[len(splits) - 1]))]
    MakeDir(dir)
}

func InitDir() {
    MakeDir(MOVIE_IMAGES_DIR)
    MakeDir(IMG_DIR + TABLE_NEW_TIMES)
    MakeDir(IMG_DIR + TABLE_FIRE)
}

func ReadFileBinary(filePath string) ([]byte, error){
    file, err := os.Open(filePath)
    if err != nil {
        logs.Trace("[Error] ", err)
    }
    defer file.Close()

    stats, err := file.Stat()
    if err != nil {
        return nil, err
    }

    var size int64 = stats.Size()
    buff := make([]byte, size)

    for {
        len, err := file.Read(buff)
        if err == io.EOF || len < 0 {
            break
        }
    }
    return buff, err
}

func GenerateDirectoryName(title string) string {
    r := []rune(title)
	s := []string{}
	cnstr := ""
	for i := 0; i < len(r); i++ {
		if r[i] <= 40869 && r[i] >= 2048 && r[i] != 12304 && r[i] != 12305  {
			cnstr = cnstr + string(r[i])
			s = append(s, cnstr)

		}
	}
	return cnstr
}

func GetAllFiles(dir string) []string {
	files := make([]string, 0)
    dirs, err := ioutil.ReadDir(dir)
    if err != nil {
        return files
    }

    for _, fi := range dirs {
        if !fi.IsDir() { 
            ok := strings.HasSuffix(fi.Name(), ".jpg")
            if ok {
                files = append(files, dir + "/" + fi.Name())
            }
        }
    }

    return files
}

func GetAllDirectories(directory string) []NormalInfoNoPic {
	infos := make([]NormalInfoNoPic, 0)
	dirs, _ := ioutil.ReadDir(directory)
	for _, dir := range dirs {
		if dir.IsDir() {
			files := GetAllFiles(directory + dir.Name())
			for _, file := range files {
				infos = append(infos, NormalInfoNoPic{ Title:dir.Name(), Path:file })
			}
		}
	}
	return infos
}

func GetStatus(count *int64) Status {
    (*count)++
    if *count % 3 == 0 {
        return One
    } else if *count % 3 == 1 {
        return Two
    } else if *count % 3 == 2 {
        return Three
    }
    return ""
}

func ColorPrint(status Status, operation string, time string, color int) { 
    kernel32 := syscall.NewLazyDLL("kernel32.dll")
    proc := kernel32.NewProc("SetConsoleTextAttribute")
    handle, _, _ := proc.Call(uintptr(syscall.Stdout), uintptr(color))
    fmt.Printf("%s %s %s\r", status, operation, time)
    handle, _, _ = proc.Call(uintptr(syscall.Stdout), uintptr(7))
    CloseHandle := kernel32.NewProc("CloseHandle")
    CloseHandle.Call(handle)
}

func CreateImageTable(tableName string) {
    sqlTable := fmt.Sprintf("%s %s %s %s %s %s %s %s", 
    "CREATE TABLE", 
    tableName, 
    "(`id` int(10) unsigned NOT NULL AUTO_INCREMENT,", 
    "`title` varchar(100) NULL COMMENT '名称',",
    "`path` varchar(200) NULL COMMENT '路径',",
    "`data` MEDIUMBLOB COMMENT '图片',",
    "PRIMARY KEY(`id`)",
    ") ENGINE=InnoDB DEFAULT CHARSET=utf8 COMMENT='图片表'");
	db.Exec(sqlTable)
}

func CountToTable(tableName string, count int) string {
    if count > MaxItemCount {
		return tableName + strconv.Itoa(NewTimesTableCount / MaxItemCount)
    }
    return tableName
}

func SizeToString(size int) string {
    var (
        integer int
        decimal int
    )
    if size < 1000 {
        return fmt.Sprintf("%7dB ", size)
    } else if size < 1000 * 1000 {
        integer = size / 1024
        decimal = int((float64(size) / 1024 - float64(integer)) * 1000)
        return fmt.Sprintf("%3d.%3dKB", integer, decimal)
    } else {
        integer = size / (1024 * 1024)
        decimal = int((float64(size) / (1024 * 1024) - float64(integer)) * 1000)
        return fmt.Sprintf("%3d.%3dMB", integer, decimal)
    }
}
