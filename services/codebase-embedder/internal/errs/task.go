package errs

import "errors"

var FileNotFound = errors.New("file or directory not found")
var ReadTimeout = errors.New("read timeout")
var RunTimeout = errors.New("run timeout")
