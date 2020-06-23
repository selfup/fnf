package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"
	"sync"
)

func main() {
	var paths string
	flag.StringVar(&paths, "p", "", "any absolute path - can be comma delimited: Example: $HOME or '/tmp,/usr'")

	var keywords string
	flag.StringVar(&keywords, "k", "", "keywords the filename(s) contain(s) - can be comma delimited: Example 'wow' or 'wow,omg,lol'")

	flag.Parse()

	scanKeywords := strings.Split(keywords, ",")
	scanPaths := strings.Split(paths, ",")

	nfnf := NewFileNameFinder(scanKeywords)

	for _, path := range scanPaths {
		nfnf.Scan(path)
	}

	for _, file := range nfnf.Files {
		fmt.Println(file)
	}
}

// FileNameFinder struct contains needed data to perform concurrent operations
type FileNameFinder struct {
	mutex     sync.Mutex
	Direction string
	Files     []string
	Keywords  []string
}

// NewFileNameFinder creates a pointer to FileNameFinder with default values
func NewFileNameFinder(keywords []string) *FileNameFinder {
	fnf := new(FileNameFinder)

	if runtime.GOOS == "windows" {
		fnf.Direction = "\\"
	} else {
		fnf.Direction = "/"
	}

	fnf.Keywords = keywords

	return fnf
}

// Scan is a concurrent/parallel directory walker
func (f *FileNameFinder) Scan(directory string) {
	_, err := ioutil.ReadDir(directory)
	if err != nil {
		panic(err)
	}

	f.findFiles(directory, "")
}

func (f *FileNameFinder) findFiles(directory string, prefix string) {
	paths, _ := ioutil.ReadDir(directory)

	var dirs []os.FileInfo
	var files []os.FileInfo

	for _, path := range paths {
		if path.IsDir() {
			dirs = append(dirs, path)
		} else {
			files = append(files, path)
		}
	}

	for _, file := range files {
		for _, keyword := range f.Keywords {
			if strings.Contains(file.Name(), keyword) {
				f.mutex.Lock()
				f.Files = append(f.Files, directory+f.Direction+file.Name())
				f.mutex.Unlock()
			}
		}
	}

	dirLen := len(dirs)
	if dirLen > 0 {
		var dirGroup sync.WaitGroup
		dirGroup.Add(dirLen)

		for _, dir := range dirs {
			go func(diR os.FileInfo, direcTory string, direcTion string) {
				f.findFiles(direcTory+direcTion+diR.Name(), direcTory)
				dirGroup.Done()
			}(dir, directory, f.Direction)
		}

		dirGroup.Wait()
	}
}
