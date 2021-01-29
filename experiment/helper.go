package main

import (
  "os"
  configparser "github.com/bigkevmcd/go-configparser"
  _ "github.com/mattn/go-sqlite3" // Import go-sqlite3 library
)

func getConfig(path string) (*configparser.ConfigParser, error){
  p, err := configparser.NewConfigParserFromFile(path)
  if err != nil {
    return nil,err
  }

  return p,nil
}

// fileExists checks if a file exists and is not a directory before we
// try using it to prevent further errors.
func DBFileExists(filename string) bool {
    info, err := os.Stat(filename)
    if os.IsNotExist(err) {
        return false
    }
    return !info.IsDir()
}
