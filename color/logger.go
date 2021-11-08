package color

import (
	"errors"
	"io"
	"os"
	"sync"
	"text/template"

	BeeColor "github.com/beego/bee/logger/colors"
)

const (
	NormalBold = iota
	BlackLight
	WhiteLight
	RedLight
	BlueLight
	CyanLight
	YellowLight
	GreenLight
	GrayLight
	MagentaLight
	BlackBold
	WhiteBold
	RedBold
	BlueBold
	CyanBold
	YellowBold
	GreenBold
	GrayBold
	MagentaBold
)

type Logger struct {
	mutex  sync.Mutex
	output io.Writer
}

type LogRecord struct {
	Message string
}

var ColorLog = GetLogger(os.Stdout)
var once sync.Once
var instance *Logger
var logRecordTemplate *template.Template

func GetLogger(w io.Writer) *Logger {
	once.Do(func() {
		var (
			err error
			simpleLogFormat = `{{.Message}}`
		)
		logRecordTemplate, err = template.New("simpleLogFormat").Parse(simpleLogFormat)
		if err != nil {
			panic(err)
		}
		instance = &Logger{output: BeeColor.NewColorWriter(w)}
	})
	return instance
}

func (log *Logger) SetOutput(w io.Writer) {
	log.mutex.Lock()
	defer log.mutex.Unlock()
	log.output = BeeColor.NewColorWriter(w) 
}

func (log *Logger) SetColor(message string, color int) string{
	switch color {
	case NormalBold:
		return BeeColor.Bold(message)
	case BlackLight:
		return BeeColor.Black(message)
	case WhiteLight:
		return BeeColor.White(message)
	case RedLight:
		return BeeColor.Red(message)
	case BlueLight:
		return BeeColor.Blue(message)
	case CyanLight:
		return BeeColor.Cyan(message)
	case YellowLight:
		return BeeColor.Yellow(message)
	case GreenLight:
		return BeeColor.Green(message)
	case GrayLight:
		return BeeColor.Gray(message)
	case MagentaLight:
		return BeeColor.Magenta(message)
	case BlackBold:
		return BeeColor.BlackBold(message)
	case WhiteBold:
		return BeeColor.WhiteBold(message)
	case RedBold:
		return BeeColor.RedBold(message)
	case BlueBold:
		return BeeColor.BlueBold(message)
	case CyanBold:
		return BeeColor.CyanBold(message)
	case YellowBold:
		return BeeColor.YellowBold(message)
	case GreenBold:
		return BeeColor.GreenBold(message)
	case GrayBold:
		return BeeColor.GrayBold(message)
	case MagentaBold:
		return BeeColor.MagentaBold(message)
	default:
		return ""
	}
}

func (log *Logger) Log(message string, color int, args ...interface{}) {
	log.mutex.Lock()
	defer log.mutex.Unlock()

	// Create the logging record and pass into the output
	record := LogRecord {
		Message: log.SetColor(message, color),
	}
	err := logRecordTemplate.Execute(log.output, record)
	if err != nil {
		panic(err)
	}
}

func (log *Logger) Logs(args ...interface{}) {
	var messages = make([]string, 10)
	var colors = make([]int, 10)
	for _, arg := range args {
		switch value := arg.(type) {
		case string:
			messages = append(messages, value)
		case int:
			colors = append(colors, value)
		}
	}
	if len(messages) != len(colors) {
		panic(errors.New("The tnputing message and color are not matched!"))
	}
	for i := 0; i < len(messages); i++ {
		log.Log(messages[i], colors[i])
	}
}