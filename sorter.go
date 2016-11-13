package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "regexp"
    "bufio"
)

var GET_REL_PATH_START_PATTERN = "[/.]"
var GET_SINGLE_QUOTE_PATH_PATTERN = "'(.*)'"
var GET_DOUBLE_QUOTE_PATH_PATTERN = "\"(.*)\""
var GET_PATH_PATTERN = "/([a-zA-Z0-9]*)"
var REQUIRE_PATTERN = "(= require)"
var IMPORT_PATTERN = "import (.)* from"


func main() {
  getFiles("./../../../../../../gigster-sourceror/client/containers/", "js")
}

func getFiles(dirName string, suffix string) (selectFiles []string) {
  //get files of certain suffix in certain directory
  //processes in two parallel channels
  fChannel := allFiles(dirName)

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

func sortDeps(dirName string, file os.FileInfo) []string {
  //gets contents of specified file in specified directory
  //and sorts the contents according to airbnb styleguid
  //returns a file

  contents, err := os.Open(dirName+file.Name())
  if err != nil {
    throwErr(err)
  }

  scanner := bufio.NewScanner(contents)

  var sortedFile []string
  for scanner.Scan() {
    sortedFile = sort(sortedFile, scanner.Text(), 0)
  }

//    fileHandle, _ := os.Create("output.txt")
//    writer := bufio.NewWriter(fileHandle)
//    defer fileHandle.Close()
//    for l := range f1 {
//      fmt.Fprintln(writer, "String I want to write")
//      writer.Flush()
//    }

  return sortedFile
}

func sort(soFar []string, curr string, index int) []string {
  //uses compareLines to determine where current string should go
  //index is the index you're comparing with
  if len(soFar) == 0 {
    return []string{curr}
  }

  compared := compareLines(soFar[index], curr)
  if compared < 0 {
    //go left
    if index == 0 {
      nextStart := []string{curr}
      return append(nextStart, soFar...)
    }
    return sort(soFar, curr, index-1)
  }

  if compared > 0 {
    //go right
    if index == len(soFar)-1 {
      return append(soFar, curr)
    }
    return sort(soFar, curr, index+1)
  }

  if compared == 0 {
    return append(append(soFar[:index], curr), soFar[index:]...)
  }
  return soFar
}

func compareLines(fst string, snd string) int {
  //returns 1 if l1 comes after l2, -1 if comes before
  //0 if neither should move
  //imports before requires
  //absolute paths before relative paths
  //sort alphabetically by paths

  fstIsRequire := test(REQUIRE_PATTERN, fst)
  sndIsRequire := test(REQUIRE_PATTERN, snd)

  fstIsImport := test(IMPORT_PATTERN, fst)
  sndIsImport := test(IMPORT_PATTERN, snd)

  if !fstIsImport && !sndIsImport && !sndIsRequire && !fstIsRequire {
    //if not import or require, do nothing
    return 0
  }

  if fstIsImport && sndIsImport || !fstIsImport && !sndIsImport {

    fstRelative := test(GET_REL_PATH_START_PATTERN, fst)
    sndRelative := test(GET_REL_PATH_START_PATTERN, snd)

    if fstRelative && sndRelative || !fstRelative && !sndRelative {
      fstContent := getPathContent(fst, fstRelative)
      sndContent := getPathContent(snd, sndRelative)

      if (fstContent < sndContent) {
        return -1
      }

      if (fstContent > sndContent) {
        return 1
      }

      return 0
    } else if (fstRelative) {
      return 1
    } else {
      return -1
    }
  }

  if fstIsImport {
    return 1
  }

  return -1
}

func fileProcessor(files <-chan os.FileInfo, dirName string, suffix string) <-chan os.FileInfo {
  //takes a stream of files and outputs as stream
  //of their sorted versions
  out := make(chan os.FileInfo)
  go func() {
      for file := range files {
          isMatch := matchesSuffix(file, suffix)
          if isMatch {
            sortedFile := sortDeps(dirName, file)
            out <- sortedFile
          }
      }
      close(out)
  }()
  return out
}

func getPathContent(line string, isRelative bool) (content string) {
  //returns value of path excluding any periods or slashes
  //this makes it easier to compare paths alphabetically
  if isRelative {
    loc := rgx(GET_REL_PATH_START_PATTERN).FindStringIndex(line)
    rest := line[loc[0]:]
    return rgx(GET_REL_PATH_START_PATTERN).ReplaceAllString(rest, "")
  } else {
    content := rgx(GET_SINGLE_QUOTE_PATH_PATTERN).FindString(line)
    if content == "" {
      content = rgx(GET_DOUBLE_QUOTE_PATH_PATTERN).FindString(line)
    }
  }

  return
}

func allFiles(dirName string) <-chan os.FileInfo {
    //returns a stream of all files in a specified
    //directory via a channel

    files, err := ioutil.ReadDir(dirName)
    if err != nil {
      throwErr(err)
    }

    out := make(chan os.FileInfo)
    go func() {
        for _, file := range files {
          out <- file
        }
        close(out)
    }()
    return out
}

func test(pattern string, s string) (isMatch bool) {
  //returns a boolean indicated if the input string
  //contains input pattern
  match, err := regexp.MatchString(pattern, s)
  if err != nil {
    return false
  }
  return match
}

func matchesSuffix(file os.FileInfo, suffix string) bool {
  //tests if string occurs at end of file name
  return test(suffix+"$", file.Name())
}

func rgx(pattern string) *regexp.Regexp {
  //returns regexp version of pattern
  return regexp.MustCompile(pattern)
}

func throwErr(e error) {
  //wrapper for panic
  panic(e)
}
