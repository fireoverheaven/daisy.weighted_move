package main

import (
  "bufio"
  "log"
  "os"
  "strings"
  "strconv"
  "io"
  "fmt"
  "path/filepath"

  "github.com/jmcvetta/randutil"
)

func main() {
  weightedRoots := parse_weightdir_file(os.Args[1])
  weighted_move(os.Args[2], weightedRoots)
}

func weighted_move(src string, weightedDests []randutil.Choice) {
  START, err := filepath.Abs(src)
  if err != nil {
    log.Fatalf("Couldn't get absolute path of %s", src)
  }
  err = filepath.Walk(START,
    func(current string, info os.FileInfo, err error) error {
      if err != nil {
        log.Fatalf("%s", err)
      }
      log.Printf("Walking %s", current)
  
      if info.IsDir() {
        log.Println("  -> Directory, skipping")
        return nil
      }

      filename := filepath.Base(current)
      choice, err := randutil.WeightedChoice(weightedDests)
      if err != nil {
        log.Fatalf("Couldn't get weighted choice: %s", err)
      }
      dest := fmt.Sprintf("%s/%s", choice.Item, filename)
      log.Printf("  -> destination choice: %s", dest)

      infile, err := os.Open(current)
      if err != nil {
        log.Fatalf("Couldn't open source file: %s", err)
      }
      log.Println("  -> Opened source file")

      outfile, err := os.Create(dest)
      if err != nil {
        infile.Close()
        log.Fatalf("Couldn't open dest file: %s", err)
      }
      log.Println("  -> Created dest file")

      defer outfile.Close()
      _, err = io.Copy(outfile, infile)
      infile.Close()
      if err != nil {
        log.Fatalf("Writing to output file failed: %s", err)
      }
      log.Println("  -> Copied to output file")

      err = os.Remove(current)
      if err != nil {
        log.Fatalf("Failed removing original file: %s", err)
      }
      log.Println("  -> Removed original file")

      return nil
    })
  if err != nil {
    log.Fatalf("Couldn't walk %s", START)
  }
}

func parse_weightdir_file(weightfile_path string) []randutil.Choice {
  readFile, err := os.Open(weightfile_path)
  if err != nil {
    log.Fatalf("Couldn't open %s", weightfile_path)
  }

  fileScanner := bufio.NewScanner(readFile)
  fileScanner.Split(bufio.ScanLines)

  var weightedRoots []randutil.Choice
  var totalWeight int
  totalWeight = 0

  for fileScanner.Scan() {
    var weight int
    info := strings.Split(fileScanner.Text(), ":")
    weight, err := strconv.Atoi(info[0])
    if err != nil {
      log.Fatalf("Couldn't convert %s to integer", info[0])
    }
    path, err := filepath.Abs(info[1])
    if err != nil {
      log.Fatalf("Couldn't get absolute path of %s", info[1])
    }

    totalWeight = totalWeight + weight
    root := randutil.Choice{
      Item: path,
      Weight: weight,
    }
    weightedRoots = append(weightedRoots, root)
  }
  readFile.Close()

  return weightedRoots
}
