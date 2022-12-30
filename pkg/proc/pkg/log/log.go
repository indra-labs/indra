package log

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/cybriq/proc"
	"github.com/davecgh/go-spew/spew"
	"github.com/gookit/color"
)

// log is your generic Logger creation invocation that uses the version data
// in version.go that provides the current compilation path prefix for making
// relative paths for log printing code locations.
var log = GetLogger(proc.PathBase)

// LogLevel is a code representing a scale of importance and context for log
// entries.
type LogLevel int32

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

type LevelMap map[LogLevel]string

func (l LevelMap) String() (s string) {
	ss := make([]string, len(l))
	for i := range l {
		ss[i] = strings.TrimSpace(l[i])
	}
	return strings.Join(ss, " ")
}

// LvlStr is a map that provides the uniform width strings that are printed
// to identify the LogLevel of a log entry.
var LvlStr = LevelMap{
	Off:   "off  ",
	Fatal: "fatal",
	Error: "error",
	Warn:  "warn ",
	Info:  "info ",
	Check: "check",
	Debug: "debug",
	Trace: "trace",
}

func GetLevelName(ll LogLevel) string {
	return strings.TrimSpace(LvlStr[ll])
}

type (
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

// gLS is a helper to make more compact declarations of LevelSpec names and
// colors by using the LogLevel LvlStr map.
func gLS(lvl LogLevel, r, g, b byte) LevelSpec {
	return LevelSpec{
		Name:      LvlStr[lvl],
		Colorizer: color.Bit24(r, g, b, false).Sprintf,
	}
}

// LevelSpecs specifies the id, string name and color-printing function
var LevelSpecs = map[LogLevel]LevelSpec{
	Off:   gLS(Off, 0, 0, 0),
	Fatal: gLS(Fatal, 255, 0, 0),
	Error: gLS(Error, 255, 128, 0),
	Check: gLS(Check, 255, 255, 0),
	Warn:  gLS(Warn, 255, 255, 0),
	Info:  gLS(Info, 0, 255, 0),
	Debug: gLS(Debug, 0, 128, 255),
	Trace: gLS(Trace, 128, 0, 255),
}

var (
	tty      io.Writer = os.Stderr
	file     *os.File
	path     string
	writer   = tty
	writerMx sync.Mutex
	logLevel = Info
	// App is the name of the application. Change this at the beginning of
	// an application main.
	App = "  main"
	// allSubsystems stores all package subsystem names found in the current
	// application.
	allSubsystems []string
	CodeLoc       = true
)

func GetAllSubsystems() (o []string) {
	writerMx.Lock()
	defer writerMx.Unlock()
	o = make([]string, len(allSubsystems))
	for i := range allSubsystems {
		o[i] = allSubsystems[i]
	}
	return
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
		err = file.Close()
		file = nil
	}
	return
}

func SetLogLevel(l LogLevel) {
	writerMx.Lock()
	defer writerMx.Unlock()
	logLevel = l
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
	split := strings.Split(file, subsystem)
	if len(split) < 2 {
		output = fmt.Sprint(
			color.White.Sprint(subsystem),
			color.Gray.Sprint(
				file, ":", line,
			),
		)
	} else {
		output = fmt.Sprint(
			color.White.Sprint(subsystem),
			color.Gray.Sprint(
				split[1], ":", line,
			),
		)
	}
	return
}

// LocTimeStampFormat is a custom time format that provides millisecond precision.
var LocTimeStampFormat = "2006-01-02T15:04:05.000000Z07:00"
var timeStampFormat = time.Stamp

// SetTimeStampFormat sets a custom timeStampFormat for the logger
func SetTimeStampFormat(format string) {
	timeStampFormat = format
}

// getTimeText is a helper that returns the current time with the
// timeStampFormat that is configured.
func getTimeText(tsf string) string {
	return time.Now().Format(tsf)
}

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
		if level > Off && level <= logLevel {
			formatString := "%v%s%s%-6v %s\n"
			loc := ""
			tsf := timeStampFormat
			if CodeLoc {
				formatString = "%-58v%s%s%-6v %s\n"
				loc = GetLoc(3, subsystem)
				tsf = LocTimeStampFormat
			}
			var app string
			if len(App) > 0 {
				fmt.Sprint(" [" + App + "]")
			}
			s := fmt.Sprintf(
				formatString,
				loc,
				color.Gray.Sprint(getTimeText(tsf)),
				app,
				LevelSpecs[level].Colorizer(
					" "+LevelSpecs[level].Name+" ",
				),
				printFunc(),
			)
			writerMx.Lock()
			defer writerMx.Unlock()
			fmt.Fprintf(writer, s)
		}
	}
}

// sortSubsystemsList sorts the list of subsystems, to keep the data read-only,
// call this function right at the top of the main, which runs after
// declarations and main/init. Really this is just here to alert the reader.
func sortSubsystemsList() {
	sort.Strings(allSubsystems)
}

// Add adds a subsystem to the list of known subsystems and returns the
// string so it is nice and neat in the package logg.go file
func Add(pathBase string) (subsystem string) {
	var ok bool
	var file string
	_, file, _, ok = runtime.Caller(2)
	if ok {
		r := strings.Split(file, pathBase)
		fromRoot := filepath.Base(file)
		if len(r) > 1 {
			fromRoot = r[1]
		}
		split := strings.Split(fromRoot, "/")
		subsystem = strings.Join(split[:len(split)-1], "/")
		writerMx.Lock()
		defer writerMx.Unlock()
		allSubsystems = append(allSubsystems, subsystem)
		sortSubsystemsList()
	}
	return
}

// GetLogger returns a set of LevelPrinter with their subsystem preloaded
func GetLogger(pathBase string) (l *Logger) {
	ss := Add(pathBase)
	// fmt.Println("subsystems:", allSubsystems)
	return &Logger{
		getOnePrinter(Fatal, ss),
		getOnePrinter(Error, ss),
		getOnePrinter(Warn, ss),
		getOnePrinter(Info, ss),
		getOnePrinter(Debug, ss),
		getOnePrinter(Trace, ss),
	}
}

// The collection of the different types of log print functions,
// includes spew.Dump, closure and error check printers.

func _ln(l LogLevel, ss string) Println {
	return func(a ...interface{}) {
		logPrint(l, ss, joinStrings(" ", a...))()
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
func _s(level LogLevel, subsystem string) Prints {
	return func(a ...interface{}) {
		logPrint(
			level, subsystem, func() string {
				return fmt.Sprint("spew:\n\n" + spew.Sdump(a...))
			},
		)()
	}
}
func _c(level LogLevel, subsystem string) Printc {
	return func(closure func() string) {
		logPrint(level, subsystem, closure)()
	}
}
func _chk(level LogLevel, subsystem string) Chk {
	return func(e error) (is bool) {
		if e != nil {
			logPrint(level, subsystem,
				joinStrings(" ", e.Error()))()
			is = true
		}
		return
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
