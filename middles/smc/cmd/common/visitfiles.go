/*
Copyright © 2022 zbc <zbc@sangfor.com.cn>
*/
package common

import (
	"log"
	"os"
	"path/filepath"
)

/**
 *	VisitFiles callback function, called once for each visited file
 */
type VisitFilesCallback func(fpath, relPath string) error

/**
 *	Add all files in a directory to local dataset
 */
func VisitDir(dir string, cb VisitFilesCallback) error {
	filepath.Walk(dir, func(path string, f os.FileInfo, err error) error {
		if f == nil {
			return err
		}
		if f.IsDir() {
			return nil
		}
		relpath, _ := filepath.Rel(dir, path)
		return cb(path, relpath)
	})
	return nil
}

/**
 *	Visit all files in the list, apply callback function cb to each file and all files in each directory
 */
func VisitFiles(files []string, cb VisitFilesCallback) error {
	if len(files) == 0 {
		return nil
	}
	for _, file := range files {
		finfo, err := os.Stat(file)
		if err != nil {
			log.Println(err)
			return err
		}
		if finfo.IsDir() {
			err = VisitDir(file, cb)
		} else {
			err = cb(file, filepath.Base(file))
		}
		if err != nil {
			log.Println(err)
		}
	}
	return nil
}
