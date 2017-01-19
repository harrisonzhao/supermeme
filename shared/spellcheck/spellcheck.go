package spellcheckutil

import (
  "github.com/sajari/fuzzy"
  "os"
  "fmt"
  "bufio"
  "regexp"
  "strings"
  "runtime"
  "path"
  "github.com/golang/glog"
)

var spellChecker *fuzzy.Model

func SampleEnglish() []string {
  var out []string
  _, filename, _, ok := runtime.Caller(1)
  if !ok {
    glog.Fatalf("Could not obtain runtime information.")
  }
  filepath := path.Join(path.Dir(filename), "data/big.txt")
  file, err := os.Open(filepath)
  if err != nil {
    fmt.Println(err)
    return out
  }
  reader := bufio.NewReader(file)
  scanner := bufio.NewScanner(reader)
  scanner.Split(bufio.ScanLines)

  // Count the words.
  count := 0
  for scanner.Scan() {
    exp, _ := regexp.Compile("[a-zA-Z]+")
    words := exp.FindAll([]byte(scanner.Text()), -1)
    for _, word := range words {
      if len(word) > 1 {
        out = append(out, strings.ToLower(string(word)))
        count++
      }
    }
  }
  if err := scanner.Err(); err != nil {
    fmt.Fprintln(os.Stderr, "reading input:", err)
  }

  return out
}

func InitSpellChecker() {
  spellChecker = fuzzy.NewModel()
  spellChecker.SetThreshold(4)
  spellChecker.SetDepth(2)
  spellChecker.Train(SampleEnglish())
}

func GetSpellChecker() *fuzzy.Model {
  if (spellChecker == nil) {
    InitSpellChecker()
  }
  return spellChecker
}