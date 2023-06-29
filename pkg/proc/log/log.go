package log

import (
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/gookit/color"
	"go.uber.org/atomic"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
)

// The LogLevel settings used in proc
const (
	Off LogLevel = iota
	Fatal
	Error
	Check
	Warn
	Info
	Debug
	Trace
)

var (
	// LevelSpecs specifies the id, string name and color-printing function
	LevelSpecs = map[LogLevel]LevelSpec{
		Off:   gLS(Off, 0, 0, 0),
		Fatal: gLS(Fatal, 255, 0, 0),
		Error: gLS(Error, 255, 128, 0),
		Check: gLS(Check, 255, 255, 0),
		Warn:  gLS(Warn, 255, 255, 0),
		Info:  gLS(Info, 0, 255, 0),
		Debug: gLS(Debug, 0, 128, 255),
		Trace: gLS(Trace, 128, 0, 255),
	}
	// LocTimeStampFormat is a custom time format that provides millisecond precision.
	LocTimeStampFormat = "2006-01-02T15:04:05.000000"
	// LvlStr is a map that provides the uniform width strings that are printed
	// to identify the LogLevel of a log entry.
	LvlStr = LevelMap{
		Off:   "off",
		Fatal: "fatal",
		Error: "error",
		Warn:  "warn",
		Info:  "info",
		Check: "check",
		Debug: "debug",
		Trace: "trace",
	}
	// log is your generic Logger creation invocation that uses the version data
	// in version.go that provides the current compilation path prefix for making
	// relative paths for log printing code locations.
	log     = GetLogger()
	lvlStrs = map[string]LogLevel{
		"off":   Off,
		"fatal": Fatal,
		"error": Error,
		"check": Check,
		"warn":  Warn,
		"info":  Info,
		"debug": Debug,
		"trace": Trace,
	}
	startTime                 = time.Now()
	timeStampFormat           = "15:04:05.000000"
	tty             io.Writer = os.Stderr
	file            *os.File
	path            string
	writer          = tty
	writerMx        sync.Mutex
	logLevel        = Info
	// App is the name of the application. Change this at the beginning of
	// an application main.
	App     atomic.String
	Longest atomic.Uint32
	// allSubsystems stores all package subsystem names found in the current
	// application.
	allSubsystems []string
)

func init() {
	Longest.Store(maxLen)
}

type (
	LevelMap map[LogLevel]string
	// LogLevel is a code representing a scale of importance and context for log
	// entries.
	LogLevel int32
	// Println prints lists of interfaces with spaces in between
	Println func(a ...interface{})
	// Printf prints like fmt.Println surrounded by log details
	Printf func(format string, a ...interface{})
	// Prints  prints a spew.Sdump for an interface slice
	Prints func(a ...interface{})
	// Printc accepts a function so that the extra computation can be avoided if
	// it is not being viewed
	Printc func(closure func() string)
	// Chk is a shortcut for printing if there is an error, or returning true
	Chk func(e error) bool
	// LevelPrinter defines a set of terminal printing primitives that output
	// with extra data, time, level, and code location
	LevelPrinter struct {
		Ln Println
		// F prints like fmt.Println surrounded by log details
		F Printf
		// S uses spew.dump to show the content of a variable
		S Prints
		// C accepts a function so that the extra computation can be avoided if
		// it is not being viewed
		C Printc
		// Chk is a shortcut for printing if there is an error, or returning
		// true
		Chk Chk
	}
	// LevelSpec is a key pair of log level and the text colorizer used
	// for it.
	LevelSpec struct {
		Name      string
		Colorizer func(format string, a ...interface{}) string
	}
	// Logger is a set of log printers for the various LogLevel items.
	Logger struct {
		F, E, W, I, D, T LevelPrinter
	}
)

// Add adds a subsystem to the list of known subsystems and returns the
// string.
func Add() (subsystem string) {
	writerMx.Lock()
	defer writerMx.Unlock()
	_, f, _, _ := runtime.Caller(1)
	subsystemPath := filepath.Dir(f)
	allSubsystems = append(allSubsystems, subsystemPath)
	sortSubsystemsList()
	return
}

func GetAllSubsystems() (o []string) {
	writerMx.Lock()
	defer writerMx.Unlock()
	o = make([]string, len(allSubsystems))
	for i := range allSubsystems {
		o[i] = allSubsystems[i]
	}
	return
}

func GetLevelByString(lvl string, def LogLevel) (ll LogLevel) {
	var exists bool
	if ll, exists = lvlStrs[lvl]; !exists {
		return def
	}
	return ll
}

func GetLevelName(ll LogLevel) string {
	return strings.TrimSpace(LvlStr[ll])
}

// GetLoc calls runtime.Caller and formats as expected by source code editors
// for terminal hyperlinks
//
// Regular expressions and the substitution texts to make these clickable in
// Tilix and other RE hyperlink configurable terminal emulators:
//
// This matches the shortened paths generated in this command and printed at
// the very beginning of the line as this logger prints:
/*
	^((([\/a-zA-Z@0-9-_.]+/)+([a-zA-Z@0-9-_.]+)):([0-9]+))
	/usr/local/bin/goland --line $5 $GOPATH/src/github.com/p9c/matrjoska/$2
*/
// I have used a shell variable there but tilix doesn't expand them,
// so put your GOPATH in manually, and obviously change the repo subpath.
func GetLoc(skip int, subsystem string) (output string) {
	_, file, line, _ := runtime.Caller(skip)
	defer func() {
		if r := recover(); r != nil {
			fmt.Fprintln(
				os.Stderr, "getloc panic on subsystem",
				subsystem, file,
			)
		}
	}()
	//split := strings.Split(file, subsystem)
	//if len(split) < 2 {
	//	output = fmt.Sprint(
	//		color.White.Sprint(subsystem),
	//		color.Gray.Sprint(
	//			file, ":", line,
	//		),
	//	)
	//} else {
	output = fmt.Sprint(
		subsystem,
		file, ":", line,
	)
	//}
	return
}

func GetLogLevel() (l LogLevel) {
	writerMx.Lock()
	defer writerMx.Unlock()
	l = logLevel
	return
}

// GetLogger returns a set of LevelPrinter with their subsystem preloaded
func GetLogger() (l *Logger) {
	ss := Add()
	return &Logger{
		getOnePrinter(Fatal, ss),
		getOnePrinter(Error, ss),
		getOnePrinter(Warn, ss),
		getOnePrinter(Info, ss),
		getOnePrinter(Debug, ss),
		getOnePrinter(Trace, ss),
	}
}

func SetLogFilePath(p string) (err error) {
	writerMx.Lock()
	defer writerMx.Unlock()
	if file != nil {
		err = StopLogToFile()
		if err != nil {
			return err
		}
		path = p
		err = StartLogToFile()
	} else {
		path = p
	}
	return
}

func SetLogLevel(l LogLevel) {
	writerMx.Lock()
	defer writerMx.Unlock()
	logLevel = l
}

// SetTimeStampFormat sets a custom timeStampFormat for the logger
func SetTimeStampFormat(format string) {
	timeStampFormat = format
}

func StartLogToFile() (err error) {
	writerMx.Lock()
	defer writerMx.Unlock()
	file, err = os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", path)
		return err
	}
	writer = io.MultiWriter(tty, file)
	return
}

func StopLogToFile() (err error) {
	writerMx.Lock()
	defer writerMx.Unlock()
	writer = tty
	if file != nil {
		// err = file.Close()
		file = nil
	}
	return
}

func (l LevelMap) String() (s string) {
	ss := make([]string, len(l))
	for i := range l {
		ss[i] = strings.TrimSpace(l[i])
	}
	return strings.Join(ss, " ")
}

func _c(level LogLevel, subsystem string) Printc {
	return func(closure func() string) {
		logPrint(level, subsystem, closure)()
	}
}
func _chk(level LogLevel, subsystem string) Chk {
	return func(e error) (is bool) {
		if e != nil {
			logPrint(level,
				subsystem,
				joinStrings(
					" ",
					e))()
			is = true
		}
		return
	}
}

func _f(level LogLevel, subsystem string) Printf {
	return func(format string, a ...interface{}) {
		logPrint(
			level, subsystem, func() string {
				return fmt.Sprintf(format, a...)
			},
		)()
	}
}

// The collection of the different types of log print functions,
// includes spew.Dump, closure and error check printers.

func _ln(l LogLevel, ss string) Println {
	return func(a ...interface{}) {
		logPrint(l, ss, joinStrings(" ", a...))()
	}
}
func _s(level LogLevel, subsystem string) Prints {
	return func(a ...interface{}) {
		text := "spew:\n"
		if s, ok := a[0].(string); ok {
			text = strings.TrimSpace(s) + "\n"
			a = a[1:]
		}
		logPrint(
			level, subsystem, func() string {
				return text + spew.Sdump(a...)
			},
		)()
	}
}

// gLS is a helper to make more compact declarations of LevelSpec names and
// colors by using the LogLevel LvlStr map.
func gLS(lvl LogLevel, r, g, b byte) LevelSpec {
	return LevelSpec{
		Name:      LvlStr[lvl],
		Colorizer: color.Bit24(r, g, b, false).Sprintf,
	}
}

func getOnePrinter(level LogLevel, subsystem string) LevelPrinter {
	return LevelPrinter{
		Ln:  _ln(level, subsystem),
		F:   _f(level, subsystem),
		S:   _s(level, subsystem),
		C:   _c(level, subsystem),
		Chk: _chk(level, subsystem),
	}
}

// getTimeText is a helper that returns the current time with the
// timeStampFormat that is configured.
func getTimeText(tsf string) string { return time.Now().Format(tsf) }

// joinStrings constructs a string from a slice of interface same as Println but
// without the terminal newline
func joinStrings(sep string, a ...interface{}) func() (o string) {
	return func() (o string) {
		for i := range a {
			o += fmt.Sprint(a[i])
			if i < len(a)-1 {
				o += sep
			}
		}
		return
	}
}

// logPrint is the generic log printing function that provides the base
// format for log entries.
func logPrint(
	level LogLevel,
	subsystem string,
	printFunc func() string,
) func() {
	return func() {
		writerMx.Lock()
		defer writerMx.Unlock()
		if level > logLevel {
			return
		}
		tsf := "15:04:05"
		timeText := getTimeText(tsf)
		var loc string
		if int(Longest.Load()) < len(loc) {
			Longest.Store(uint32(len(loc)))
		}
		formatString := "%s%s %s%s %s"
		//timeText = time.Now().Format("2006-01-02 15:04:05.999999999 UTC+0700")
		if logLevel > Info {
			loc = GetLoc(3, subsystem)
			count := int(Longest.Load()) - len(loc) + 1
			if count < 0 {
				count = 0
			}
			loc = loc + strings.Repeat(" ", count)
			tsf = timeStampFormat
		}
		var app string
		if len(App.Load()) > 0 {
			//	app = fmt.Sprint(App.Load(), "")
			app = App.Load()
		}
		s := fmt.Sprintf(
			formatString,
			loc,
			timeText,
			strings.ToUpper(app)+" ",
			LevelSpecs[level].Colorizer(
				LvlStr[level],
			),
			printFunc(),
		)
		s = strings.TrimSuffix(s, "\n")
		fmt.Fprintln(writer, s)
	}
}

// sortSubsystemsList sorts the list of subsystems, to keep the data read-only,
// call this function right at the top of the main, which runs after
// declarations and main/init. Really this is just here to alert the reader.
func sortSubsystemsList() { sort.Strings(allSubsystems) }
