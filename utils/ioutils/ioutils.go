package ioutils

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
)

var(
	ErrFileIsExists = errors.New("the file had been exists")
)

type MutexIO interface{
	Lock()
	Unlock()
	Set(writer interface{})
	Write(p []byte) (n int, err error)
}

type FileMutexIO struct {
	*os.File
	*sync.Mutex
	autLock bool
	path string
}

func NewFileMutexIO(autoLock bool) *FileMutexIO  {
	return &FileMutexIO{
		Mutex: &sync.Mutex{},
		autLock:autoLock,
	}
}

func (fm *FileMutexIO) Set(writer interface{}) {
	if fm.autLock{
		fm.Lock()
		defer fm.Unlock()
	}

	if fd, isOk := writer.(*os.File); isOk{
		fm.File = fd
	}
}

func (fm *FileMutexIO)Write(p []byte) (n int, err error){
	if fm.autLock{
		fm.Lock()
		fm.Unlock()
	}

	return fm.File.Write(p)
}

func (fm *FileMutexIO)SetPath(path string)  {
	if fm.autLock{
		fm.Lock()
		defer fm.Unlock()
	}

	fm.path = path
}

func (fm *FileMutexIO)Path() string {
	if fm.autLock{
		fm.Lock()
		defer fm.Unlock()
	}
	return  fm.path
}

// FileExists checks whether the given file exists.
// If the file exists, this method also returns the size of the file.
func FileExists(filePath string) (bool, int64, error) {
	fileInfo, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false, 0, nil
	}
	return true, fileInfo.Size(), err
}

// If the file exists , open it else create it.
// is Permission not enough , create file in other dir
func OpenFile(filePath string, other string)(*os.File, error){

	fd, err := os.OpenFile(filePath, os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
	if err != nil && os.IsPermission(err) && other != ""{
		_, filePath = filepath.Split(filePath)
		filePath = filepath.Join(other, filePath)
		return os.OpenFile(filePath, os.O_APPEND | os.O_CREATE | os.O_RDWR, 0666)
	}

	return fd, err
}

func logDirStatus(msg string, dirPath string) {
	exists, _, err := FileExists(dirPath)
	if err != nil {
		log.Errorf("Error while checking for dir existence")
	}
	if exists {
		log.Debugf("%s - [%s] exists", msg, dirPath)
	} else {
		log.Debugf("%s - [%s] does not exist", msg, dirPath)
	}
}


// DirEmpty returns true if the dir at dirPath is empty
func DirEmpty(dirPath string) (bool, error) {
	f, err := os.Open(dirPath)
	if err != nil {
		log.Debugf("Error while opening dir [%s]: %s", dirPath, err)
		return false, err
	}
	defer f.Close()

	_, err = f.Readdir(1)
	if err == io.EOF {
		return true, nil
	}
	return false, err
}

func CreateDirIfMissing(dirPath string) (bool, error) {
	// if dirPath does not end with a path separator, it leaves out the last segment while creating directories
	if !strings.HasSuffix(dirPath, "/") {
		dirPath = dirPath + "/"
	}
	log.Debugf("CreateDirIfMissing [%s]", dirPath)
	logDirStatus("Before creating dir", dirPath)
	err := os.MkdirAll(path.Dir(dirPath), 0755)
	if err != nil {
		log.Debugf("Error while creating dir [%s]", dirPath)
		return false, err
	}
	logDirStatus("After creating dir", dirPath)
	return DirEmpty(dirPath)
}

func GetFileData(fName string)([]byte, error) {
	var path string // 保存音频文件的地址
	if path = os.Getenv("SPEAK_IN_DIR"); path ==""{
		path = os.Getenv("TEMP")
	}

	path = filepath.Join(path, "voice" , fName)
	if fd, err := os.Open(path); err != nil{
		return nil, err
	}else {
		defer fd.Close()

		return ioutil.ReadAll(fd)
	}

	return nil, nil
}

func SetFileData(fName string, data []byte, recover bool)(string, error)  {
	// TODO: 生成唯一的 key

	// 保存文件到本地
	var path string // 保存音频文件的地址
	if path = os.Getenv("SPEAK_IN_DIR"); path ==""{
		path = os.Getenv("TEMP")
	}

	path = filepath.Join(path, "voice")
	if _, err := CreateDirIfMissing(path); err != nil{
		return "", err
	}

	if isExists, _, err := FileExists(path); err != nil{
		return "", err
	}else if isExists && !recover{
		return "", ErrFileIsExists
	}

	path = filepath.Join(path, fName)
	if fd, err := os.Create(path); err != nil{
		return "", err
	}else {
		defer fd.Close()
		if _, err := fd.Write(data); err !=nil{
			return "", err
		}
	}

	return fName, nil
}

func TempDir() string {
	if temp := os.Getenv("TEMP"); temp != ""{
		return temp
	}

	return "/tmp"
}