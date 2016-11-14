/*
Package main executes a parralel dependency sorter that looks at JavaScript files in current directory
and sorts their dependencies according to the aibnb style guide
*/

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

//getFiles gets files of certain suffix in certain directory
//processes in n parallel channels
func getFiles(dirName string, suffix string) {
  fChannel := allFiles(dirName)
  parralelizer(2, fChannel, dirName, suffix)
}

//parralelizer is a utility that multiplexes the work done by this code processor
func parralelizer(totalThreads int, fChannel <-chan os.FileInfo, dirName string, suffix string) {
  for nthThread := 0; nthThread < totalThreads; nthThread++ {
    p := fileProcessor(fChannel, dirName, suffix)

    for fileName := range p {
      fmt.Println(fileName)
    }
  }
}

//sortDeps gets contents of specified file in specified directory
//and sorts the contents according to airbnb style guide
//returns name of file just sorted
func sortDeps(dirName string, file os.FileInfo, suffix string) string {
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

//sort puts curr in next appropriate spot by scanning soFar and using compareLines
//to determine in which part of soFar curr should go
func sort(soFar []string, curr string, index int) []string {
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

//prepend is a utility for prepending a string
//to a slice
func prepend(curr string, soFar []string) []string {
  return append([]string{curr}, soFar...)
}

//sandwich sticks curr between slice1
//and slice2 and returns the result
func sandwich(slice1 []string, slice2 []string, curr string) []string {
  secondHalf := make([]string, len(slice2))
  copy(secondHalf, slice2)
  firstHalf := append(slice1, curr)
  return append(firstHalf, secondHalf...)
}

//compareLines returns 1 if fst should come after snd,
//-1 or 0 if fst and snd are in proper order or -2 if we're
//looking at a code block.
//This is where the logic from the airbnb style guide comes
//into play. The sorting is implemented with the notion that
//imports come before requires
//absolute paths come before relative paths
//and paths are sorted alphabetically
func compareLines(fst string, snd string) int {
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

//fileProcessor takes a stream of files and outputs as stream
//the file names of files successfully processed
//while writing the sorted results to output files
func fileProcessor(files <-chan os.FileInfo, dirName string, suffix string) <-chan string {
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

//getPathContent returns value of path excluding periods or slashes.
//This makes it easier to compare paths alphabetically
func getPathContent(line string, isRelative bool) string {
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

//allFiles returns a stream of all files in a specified
//directory via a channel
func allFiles(dirName string) <-chan os.FileInfo {
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

//test returns a boolean which indicates if the input string
//contains input pattern
func test(pattern string, s string) (isMatch bool) {
  match, err := regexp.MatchString(pattern, s)
  if err != nil {
    return false
  }
  return match
}

//matchesSuffix tests if string occurs at end of file name
func matchesSuffix(file os.FileInfo, suffix string) bool {
  return test(suffix+"$", file.Name())
}

//rgx returns regexp version of pattern
func rgx(pattern string) *regexp.Regexp {
  return regexp.MustCompile(pattern)
}

//throwErr is a wrapper for panic
func throwErr(e error) {
  panic(e)
}
