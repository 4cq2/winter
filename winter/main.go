package main

import (
   "database/sql"
   "flag"
   "fmt"
   "os"
   _ "github.com/mattn/go-sqlite3"
)

func main() {
   database := `D:\Music\Backblaze\winter.db`
   flag.StringVar(&database, "d", database, "database")
   flag.Parse()
   db, err := sql.Open("sqlite3", database)
   if err != nil {
      panic(err)
   }
   defer db.Close()
   argc := len(os.Args)
   if argc >= 2 {
      switch {
      case argc == 2 && os.Args[1] != "artist":
         err = select_one(db, os.Args[1]) // winter radiohead
      case argc == 4 && os.Args[1] == "check":
         _, err = db.Exec(`
         UPDATE artist_t SET check_s = ? WHERE artist_n = ?
         `, os.Args[3], os.Args[2]) // winter check 9 2020-12-31
      case argc == 2 && os.Args[1] == "artist":
         err = select_all(db) // winter artist
      case argc == 3 && os.Args[1] == "artist":
         _, err = db.Exec(`
         INSERT INTO artist_t (artist_s, check_s, mb_s) VALUES (?, '', '')
         `, os.Args[2]) // winter artist radiohead
      case argc == 4 && os.Args[1] == "date":
         _, err = db.Exec(`
         UPDATE album_t SET date_s = ? WHERE album_n = ?
         `, os.Args[3], os.Args[2]) // winter date 9 2020-12-31
      case argc == 4 && os.Args[1] == "mb":
         _, err = db.Exec(`
         UPDATE artist_t SET mb_s = ? WHERE artist_n = ?
         `, os.Args[3], os.Args[2]) // winter mb 9 a74b1b7f-71a5-4011-9441-d0b5e
      case argc == 4 && os.Args[1] == "note":
         _, err = db.Exec(`
         UPDATE song_t SET note_s = ? WHERE song_n = ?
         `, os.Args[3], os.Args[2]) // winter note 9 good
      case argc == 4 && os.Args[1] == "url":
         _, err = db.Exec(`
         UPDATE album_t SET url_s = ? WHERE album_n = ?
         `, os.Args[3], os.Args[2]) // winter url 9 youtube.com/watch?v=XFkzRNyy
      case argc == 3 && os.Args[1] == "album":
         err = delete_album(db, os.Args[2]) // winter album 9
      case argc == 4 && os.Args[1] == "album":
         err = copy_album(db, os.Args[2], os.Args[3]) // winter album 9 9
      }
      if err != nil {
         panic(err)
      }
   } else {
      fmt.Println(`Copy album:
   winter album 999 1000

Delete album:
   winter album 999

Select all artist:
   winter artist

Select one artist:
   winter 'Kate Bush'

Insert artist:
   winter artist 'Kate Bush'

Update artist date:
   winter check 999 2020-12-31

Update artist id:
   winter mb 999 3f5be744-e867-42fb-8913-5fd69e4099b5

Update album date:
   winter date 999 2020-12-31

Update album URL:
   winter url 999 youtube.com/watch?v=HQmmM_qwG4k

Update song note:
   winter note 999 good`)
   }
}
