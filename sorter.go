package main

import (
    "fmt"
    "os"
    "io/ioutil"
)

func main() {
  getFiles("./../../gigster-sourceror/client/", "js")
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

  p1 := fileProcessor(fChannel)
  p2 := fileProcessor(fChannel)

  for f1 := range p1 {
    fmt.Println(f1)
  }

  for f2 := range p2 {
    fmt.Println(f2)
  }

  return
}

func fileProcessor(files <-chan os.FileInfo) <-chan os.FileInfo {
    //TODO output will be js files with sorted dependencies
    out := make(chan os.FileInfo)
    go func() {
        for file := range files {
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

