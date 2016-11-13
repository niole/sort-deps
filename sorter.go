package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "regexp"
)

func main() {
  getFiles("./../../../../../../gigster-sourceror/client/containers", "js")
}

func throwErr(e error) {
  panic(e)
}

func getFiles(dirName string, suffix string) (selectFiles []string) {
  //get files of certain suffix in certain directory
  files, err := ioutil.ReadDir(dirName)
  if err != nil {
    throwErr(err)
  }

  fChannel := allFiles(files)

  p1 := fileProcessor(fChannel, dirName, suffix)
  p2 := fileProcessor(fChannel, dirName, suffix)

  for f1 := range p1 {
    fmt.Println(f1)
  }

  for f2 := range p2 {
    fmt.Println(f2)
  }

  return
}

func matchesSuffix(file os.FileInfo, suffix string) (isMatch bool) {
  match, err := regexp.MatchString(suffix+"$", file.Name())
  isMatch = false
  if err != nil {
    isMatch = false
  } else if (match) {
    isMatch = true
  }
  return
}

func fileProcessor(files <-chan os.FileInfo, dirName string, suffix string) <-chan os.FileInfo {
    //TODO output will be js files with sorted dependencies
    out := make(chan os.FileInfo)
    go func() {
        for file := range files {
            isMatch := matchesSuffix(file, suffix)
            fmt.Println(isMatch)
            out <- file
        }
        close(out)
    }()
    return out
}

func allFiles(files []os.FileInfo) <-chan os.FileInfo {
    out := make(chan os.FileInfo)
    go func() {
        for _, file := range files {
          out <- file
        }
        close(out)
    }()
    return out
}
