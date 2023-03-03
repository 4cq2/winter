package main

import (
   "database/sql"
   "embed"
   "fmt"
   "html/template"
   "net/http"
   "strings"
   "time"
)

func new_music_artist(db *sql.DB, like string) (*music_artist, error) {
   var artist music_artist
   err := db.QueryRow(`
   SELECT * FROM artist_t WHERE artist_s LIKE ?
   `, like).Scan(&artist.ID, &artist.Name, &artist.Check, &artist.MB)
   if err != nil {
      return nil, err
   }
   recs, err := records(db, like)
   if err != nil {
      return nil, err
   }
   var (
      album music_album
      prev = -1
      track_name = make(map[string]bool)
   )
   for _, rec := range recs {
      var track music_track
      track.ID = rec.song_ID
      track.Name = rec.song
      if track_name[strings.ToUpper(rec.song)] {
         track.Note = "duplicate"
      } else {
         track.Note = rec.note
      }
      if rec.album_ID != prev {
         if prev >= 0 {
            artist.Album = append(artist.Album, album)
         }
         album.ID = rec.album_ID
         album.Name = rec.album
         album.Date = rec.date
         if strings.HasPrefix(rec.url, "http") {
            album.URL = rec.url
         } else {
            album.URL = "http://" + rec.url
         }
         album.Track = []music_track{track}
         prev = rec.album_ID
      } else {
         album.Track = append(album.Track, track)
      }
      track_name[strings.ToUpper(rec.song)] = true
   }
   // append last album
   artist.Album = append(artist.Album, album)
   return &artist, nil
}

func records(db *sql.DB, like string) ([]record, error) {
   rows, err := db.Query(`
   SELECT
      album_n,
      album_s,
      date_s,
      note_s,
      song_n,
      song_s,
      url_s
   FROM album_t
   NATURAL JOIN song_t
   NATURAL JOIN song_artist_t
   NATURAL JOIN artist_t
   WHERE artist_s LIKE ?
   ORDER BY date_s
   `, like)
   if err != nil {
      return nil, err
   }
   defer rows.Close()
   var recs []record
   for rows.Next() {
      var rec record
      err := rows.Scan(
         &rec.album_ID,
         &rec.album,
         &rec.date,
         &rec.note,
         &rec.song_ID,
         &rec.song,
         &rec.url,
      )
      if err != nil {
         return nil, err
      }
      recs = append(recs, rec)
   }
   return recs, nil
}

func select_one(db *sql.DB, like string) error {
   http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
      w.Header().Set("Content-Type", "text/html")
      artist, err := new_music_artist(db, like)
      if err != nil {
         fmt.Println(err)
      } else {
         tem, err := template.ParseFS(content, "select-one.html")
         if err != nil {
            fmt.Println(err)
         } else {
            tem.Execute(w, artist)
         }
      }
   })
   fmt.Println("localhost:99")
   return http.ListenAndServe(":99", nil)
}

//go:embed select-one.html
var content embed.FS

type music_artist struct {
   ID int
   Name string
   Check string
   MB string
   Album []music_album
}

type music_album struct {
   ID int
   Name string
   Date string
   URL string
   Track []music_track
}

type music_track struct {
   ID int
   Name string
   Note string
}

type record struct {
   album string
   album_ID int
   date string
   note string
   song string
   song_ID int
   url string
}

func select_all(db *sql.DB) error {
   then := time.Now().AddDate(-1, 0, 0)
   rows, err := db.Query(`
   SELECT
      count(1) filter (WHERE note_s = 'good') AS good,
      artist_s
   FROM artist_t
   NATURAL JOIN song_artist_t
   NATURAL JOIN song_t
   WHERE check_s < ?
   GROUP BY artist_n
   ORDER BY good
   `, then)
   if err != nil {
      return err
   }
   defer rows.Close()
   for rows.Next() {
      var (
         artist string
         count int
      )
      err := rows.Scan(&count, &artist)
      if err != nil {
         return err
      }
      fmt.Println(count, "|", artist)
   }
   return nil
}

func delete_album(db *sql.DB, album string) error {
   tx, err := db.Begin()
   if err != nil {
      return err
   }
   defer tx.Commit()
   rows, err := tx.Query("SELECT song_n FROM song_t WHERE album_n = ?", album)
   if err != nil {
      return err
   }
   defer rows.Close()
   var songs []int
   for rows.Next() {
      var song int
      err := rows.Scan(&song)
      if err != nil {
         return err
      }
      songs = append(songs, song)
   }
   for _, song := range songs {
      _, err := tx.Exec("DELETE FROM song_t WHERE song_n = ?", song)
      if err != nil {
         return err
      }
      if _, err := tx.Exec("DELETE FROM song_artist_t WHERE song_n = ?", song)
      err != nil {
         return err
      }
   }
   if _, err := tx.Exec(`DELETE FROM album_t WHERE album_n = ?`, album)
   err != nil {
      return err
   }
   return nil
}

func copy_album(db *sql.DB, source, dest string) error {
   tx, err := db.Begin()
   if err != nil {
      return err
   }
   defer tx.Commit()
   var ref string
   // COPY URL
   if err := tx.QueryRow(`
   SELECT url_s FROM album_t WHERE album_n = ?
   `, source).Scan(&ref); err != nil {
      return err
   }
   // PASTE URL
   if _, err := tx.Exec(`
   UPDATE album_t SET url_s = ? WHERE album_n = ?
   `, ref, dest); err != nil {
      return err
   }
   // COPY NOTES
   rows, err := tx.Query(`
   SELECT song_s, note_s FROM song_t WHERE album_n = ?
   `, source)
   if err != nil {
      return err
   }
   defer rows.Close()
   songs := make(map[string]string)
   for rows.Next() {
      var note, song string
      err := rows.Scan(&song, &note)
      if err != nil {
         return err
      }
      songs[song] = note
   }
   // PASTE NOTES
   for song, note := range songs {
      _, err := tx.Exec(`
      UPDATE song_t SET note_s = ?
      WHERE album_n = ? AND song_s = ? COLLATE NOCASE
      `, note, dest, song)
      if err != nil {
         return err
      }
   }
   return nil
}
