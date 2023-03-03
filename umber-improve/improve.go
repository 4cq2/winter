package main

import (
   "database/sql"
   "embed"
   "encoding/json"
   "fmt"
   "os"
   "strings"
   _ "github.com/mattn/go-sqlite3"
)

//go:embed improve.json
var content embed.FS

func main() {
   buf, err := content.ReadFile("improve.json")
   if err != nil {
      panic(err)
   }
   var config struct {
      Umber string
      Winter string
   }
   if err := json.Unmarshal(buf, &config); err != nil {
      panic(err)
   }
   db, err := sql.Open("sqlite3", config.Winter)
   if err != nil {
      panic(err)
   }
   defer db.Close()
   rows, err := db.Query("SELECT artist_s FROM artist_t")
   if err != nil {
      panic(err)
   }
   defer rows.Close()
   table := make(map[string]bool)
   for rows.Next() {
      var row string
      if err := rows.Scan(&row); err != nil {
         panic(err)
      }
      table[strings.ToUpper(row)] = true
   }
   songs, err := songs(config.Umber)
   if err != nil {
      panic(err)
   }
   for _, item := range songs {
      artists := strings.Split(item.S, " - ")[0]
      artist := strings.Split(artists, ", ")[0]
      if !table[strings.ToUpper(artist)] {
         fmt.Println(artist)
         table[strings.ToUpper(artist)] = true
      }
   }
}

type song struct {
   S string
}

func songs(umber string) ([]song, error) {
   file, err := os.Open(umber)
   if err != nil {
      return nil, err
   }
   defer file.Close()
   var songs []song
   if err := json.NewDecoder(file).Decode(&songs); err != nil {
      return nil, err
   }
   return songs, nil
}

