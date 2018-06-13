package ltlog

import (
	"sync"
	"os"
	"time"
	"log"
	"fmt"
	"runtime"
	"path/filepath"
)

type LogFile struct {
	mu   sync.Mutex
	name string
	file *os.File
}

var (
	LogLevel int = 0
	tags         = map[int]string{
		Ldebug:   "[debug]",
		Linfo:    "[info]",
		Lnotice:  "[notice]",
		Lwarning: "[warning]",
		Lerror:   "[error]",
		Lfatal:   "[fatal]",
	}
)

const (
	Ldebug   = iota
	Linfo
	Lnotice
	Lwarning
	Lerror
	Lfatal
)

// NewLogFile creates a new LogFile. The file is optional - it will be created if needed.
func NewLogFile(name string, file *os.File) (*LogFile, error) {
	rw := &LogFile{
		file: file,
		name: name,
	}
	if file == nil {
		if err := rw.Rotate(); err != nil {
			return nil, err
		}
	}

	return rw, nil
}

func (l *LogFile) Write(b []byte) (n int, err error) {
	l.mu.Lock()
	n, err = l.file.Write(b)
	l.mu.Unlock()
	return
}

func (l *LogFile) Rotate() error {
	// rename dest file if it already exists
	if _, err := os.Stat(l.name); err == nil {
		name := l.name + "." + time.Now().Format(time.RFC3339)
		if err = os.Rename(l.name, name); err != nil {
			return err
		}
	}

	// create new file
	file, err := os.Create(l.name)
	if err != nil {
		return err
	}

	// switch dest file safely
	// avoid inode same
	l.mu.Lock()
	file, l.file = l.file, file
	l.mu.Unlock()

	// close old file if open
	if file != nil {
		if err := file.Close(); err != nil {
			return err
		}
	}

	return nil
}

func Println(level int, v ...interface{}) {
	if level >= LogLevel {
		defaultPrefix(levelTag(level))
		log.Print(endPointFile(), fmt.Sprintln(v...))
		defaultPrefix("")
	}
}

func Printf(level int, format string, v ...interface{}) {
	if level >= LogLevel {
		defaultPrefix(levelTag(level))
		log.Print(endPointFile(), fmt.Sprintf(format, v...))
		defaultPrefix("")
	}
}

func Fatalf(format string, v ...interface{}) {
	defaultPrefix(levelTag(Lfatal))
	log.Fatal(endPointFile(), fmt.Sprintf(format, v...))
	defaultPrefix("")
}

func Fatalln(v ...interface{}) {
	defaultPrefix(levelTag(Lfatal))
	log.Fatal(endPointFile(), fmt.Sprintln(v...))
	defaultPrefix("")
}

func levelTag(level int) (ret string) {
	if ret, ok := tags[level]; ok {
		return ret
	}

	return "[undefined]"
}

func endPointFile() string {
	_, file, line, ok := runtime.Caller(2)
	if ok {
		return fmt.Sprintf("%s:%d: ", filepath.Base(file), line)
	}

	return "undefined"
}

func defaultPrefix(prefix string) {
	if prefix != "" {
		log.SetPrefix(prefix + " ")
	} else {
		log.SetPrefix("[default] ")
	}
}
