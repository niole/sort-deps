package main

import (
    "fmt"
    "os"
    "io/ioutil"
    "regexp"
    "bufio"
)

var GET_REL_PATH_START_PATTERN = "[/.]"
var GET_SINGLE_QUOTE_PATH_PATTERN = "'(.*?)'"
var GET_DOUBLE_QUOTE_PATH_PATTERN = "\"(.*?)\""
var GET_PATH_PATTERN = "/([a-zA-Z0-9]*)"
var REQUIRE_PATTERN = "(= require)"
var IMPORT_PATTERN = "import (.)* from"


func main() {
  getFiles("./", ".js")
}

func getFiles(dirName string, suffix string) {
  //get files of certain suffix in certain directory
  //processes in n parallel channels
  fChannel := allFiles(dirName)
  parralelizer(2, fChannel, dirName, suffix)
}

func parralelizer(totalThreads int, fChannel <-chan os.FileInfo, dirName string, suffix string) {
  for nthThread := 0; nthThread < totalThreads; nthThread++ {
    p := fileProcessor(fChannel, dirName, suffix)

    for fileName := range p {
      fmt.Println(fileName)
    }
  }
}

func sortDeps(dirName string, file os.FileInfo, suffix string) string {
  //gets contents of specified file in specified directory
  //and sorts the contents according to airbnb styleguid
  //returns a file

  fileName := file.Name()
  contents, err := os.Open(dirName+fileName)
  if err != nil {
    throwErr(err)
  }

  scanner := bufio.NewScanner(contents)

  var sortedFile []string
  for scanner.Scan() {
    sortedFile = sort(sortedFile, scanner.Text(), 0)
  }

  formattedFileName := rgx(suffix).ReplaceAllString(fileName, "")
  fileHandle, _ := os.Create("./processed/" + formattedFileName + "_sorted.txt");

  writer := bufio.NewWriter(fileHandle)
  defer fileHandle.Close()

  for _, line := range sortedFile {
    fmt.Fprintln(writer, line)
  }
  writer.Flush()

  return fileName
}

func sort(soFar []string, curr string, index int) []string {
  //uses compareLines to determine where current string should go
  //index is the index you're comparing with
  if len(soFar) == 0 {
    return []string{curr}
  }

  if index == len(soFar) {
    //append curr to end and return
    return append(soFar, curr)
  }

  compared := compareLines(soFar[index], curr)

  if compared == -2 {
    //this is a code line
    //append to end
    return append(soFar, curr)
  }

  if compared <= 0 {
    //these two are in order already
    //keep going until string should go before
    //what it's being compared with
    return sort(soFar, curr, index+1)
  }

  if compared > 0 {
    //the curr line in soFar must go right
    //so curr must go left
    if index == 0 {
      return prepend(curr, soFar)
    }
    return sandwich(soFar[:index], soFar[index:], curr)
  }

  return soFar
}

func prepend(curr string, soFar []string) []string {
  return append([]string{curr}, soFar...)
}

func sandwich(slice1 []string, slice2 []string, curr string) []string {
  //sandwhich curr between slices
  //for string slices
  secondHalf := make([]string, len(slice2))
  copy(secondHalf, slice2)
  firstHalf := append(slice1, curr)
  return append(firstHalf, secondHalf...)
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

  fstIsCode := !fstIsImport && !fstIsRequire
  sndIsCode := !sndIsImport && !sndIsRequire

  if fstIsCode&&sndIsCode || sndIsCode {
    //all is code or comparing is code
    return -2
  }

  if fstIsCode {
    //comparing against code, should go before
    return 1
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
    return -1
  }

  return 1
}

func fileProcessor(files <-chan os.FileInfo, dirName string, suffix string) <-chan string {
  //takes a stream of files and outputs as stream
  //the file names of files successfully processed
  //while writing to an output file
  out := make(chan string)
  go func() {
      for file := range files {
          isMatch := matchesSuffix(file, suffix)
          if isMatch {
            sortedFile := sortDeps(dirName, file, suffix)
            out <- sortedFile
          }
      }
      close(out)
  }()
  return out
}

func getPathContent(line string, isRelative bool) string {
  //returns value of path excluding any periods or slashes
  //this makes it easier to compare paths alphabetically
  content := rgx(GET_SINGLE_QUOTE_PATH_PATTERN).FindString(line)
  if content == "" {
    content = rgx(GET_DOUBLE_QUOTE_PATH_PATTERN).FindString(line)
  }

  if isRelative {
    return rgx(GET_REL_PATH_START_PATTERN).ReplaceAllString(content, "")
  } else {
    return content
  }
  return ""
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
