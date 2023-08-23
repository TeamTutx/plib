package ally

import (
	"io"
	"os"
	"path/filepath"

	"gitlab.com/g-harshit/plib/perror"
)

//MoveFile : move file from one drive to other
func MoveFile(source, target string) (err error) {
	err = MakeDir(target)
	if err == nil {
		if err = os.Rename(source, target); err != nil {
			err = perror.MiscError(err)
		}
	}
	return
}

//CopyFile : copy file from one path to other
func CopyFile(source, target string) (err error) {
	var (
		from *os.File
		to   *os.File
	)
	err = MakeDir(target)
	if err == nil {
		if from, err = os.Open(source); err == nil {
			defer from.Close()
			if to, err = os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0666); err == nil {
				defer to.Close()
				_, err = io.Copy(to, from)
			}
		}
	}
	return
}

//RemoveFile will return file
func RemoveFile(file string) (err error) {
	if _, err = os.Stat(file); os.IsNotExist(err) == false {
		if err = os.Remove(file); err != nil {
			err = perror.MiscError(err)
		}
	} else {
		err = perror.MiscError(err)
	}
	return
}

//FileExists : File & Directory Exist
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return true, err
}

//FileExtension : Return File Extension
func FileExtension(fileName string) string {
	ext := filepath.Ext(fileName)
	return ext
}

//MakeDir will make dir if not exists
func MakeDir(file string) (err error) {
	dir := filepath.Dir(file)
	if _, err = os.Stat(dir); os.IsNotExist(err) {
		if err = os.MkdirAll(dir, 0755); err != nil {
			err = perror.MiscError(err, "Dir Create Error")
		}
	}
	return
}
