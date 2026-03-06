package env

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

/**
 *	Logger for debug output
 */
var LogDbg *log.Logger = log.New(os.Stderr, "", log.Lshortfile|log.LstdFlags)

/**
 *	No-op Writer
 */
type NullWriter struct {
	size int64
}

/**
 *	Discard log content
 */
func (w *NullWriter) Write(p []byte) (n int, err error) {
	size := int64(len(p))
	w.size += size

	return len(p), nil
}

/**
 *	Initialize logger to record log to file
 */
func InitLogger(debug, logfile string) error {
	nullWriter := &NullWriter{}
	logDbg := false
	if strings.EqualFold(debug, "Off") { //No output
		log.SetOutput(nullWriter)
		LogDbg.SetOutput(nullWriter)
		return nil
	} else if strings.EqualFold(debug, "Dbg") {
		logDbg = true
	}
	var logStderr bool = false
	var writer io.Writer = os.Stderr

	LogDbg.SetFlags(log.Lshortfile | log.LstdFlags)
	log.SetFlags(log.Lshortfile | log.LstdFlags)

	if logfile == "" || logfile[0:1] == "+" { //Output to console
		if logfile != "" {
			logfile = logfile[1:]
		}
		logStderr = true
	}
	if logfile != "" { //Output to log file
		dir := filepath.Dir(logfile)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
		logFp, err := os.OpenFile(logfile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Println("open log file failed, err:", err)
			return err
		}
		if logStderr {
			writer = io.MultiWriter(os.Stderr, logFp)
		} else {
			writer = logFp
		}
	}
	log.SetOutput(writer)
	if logDbg {
		LogDbg.SetOutput(writer)
	} else {
		LogDbg.SetOutput(nullWriter)
	}
	return nil
}
