package main

import (
   "2a.pages.dev/nursery/musicbrainz"
   "database/sql"
   "flag"
   "os"
   "time"
   _ "github.com/mattn/go-sqlite3"
)

func main() {
   // d
   database := `D:\Music\Backblaze\winter.db`
   flag.StringVar(&database, "d", database, "database")
   // g
   var group string
   flag.StringVar(&group, "g", "", "MusicBrainz release group ID")
   // r
   var release string
   flag.StringVar(&release, "r", "", "MusicBrainz release ID")
   flag.Parse()
   db, err := sql.Open("sqlite3", database)
   if err != nil {
      panic(err)
   }
   defer db.Close()
   tx, err := db.Begin()
   if err != nil {
      panic(err)
   }
   defer tx.Commit()
   if len(os.Args) == 3 {
      rel, err := new_release(release, group)
      if err != nil {
         panic(err)
      }
      if err := insert(rel, tx); err != nil {
         panic(err)
      }
   } else {
      flag.Usage()
   }
}

type credit_error struct {
   name string
   err error
}

func (c credit_error) Error() string {
   return c.name + " " + c.err.Error()
}

func insert(album *musicbrainz.Release, tx *sql.Tx) error {
   // ALBUM
   result, err := tx.Exec(`
   INSERT INTO album_t (album_s, date_s, url_s) VALUES (?, ?, '')
   `, album.Title, album.Date)
   if err != nil {
      return err
   }
   album_ID, err := result.LastInsertId()
   if err != nil {
      return err
   }
   // CREATE ARTIST ARRAY
   var artists []int
   for _, credit := range album.Artist_Credit {
      var artist int
      err := tx.QueryRow(`
      SELECT artist_n FROM artist_t WHERE mb_s = ?
      `, credit.Artist.ID).Scan(&artist)
      if err != nil {
         return credit_error{credit.Name, err}
      }
      artists = append(artists, artist)
   }
   // CREATE SONG ARRAY
   var tns []title_note
   for _, media := range album.Media {
      for _, track := range media.Tracks {
         tns = append(tns, title_note{
            track.Title, note(track),
         })
      }
   }
   // ITERATE SONG ARRAY
   for _, tn := range tns {
      result, err := tx.Exec(`
      INSERT INTO song_t (song_s, note_s, album_n) VALUES (?, ?, ?)
      `, tn.title, tn.note, album_ID)
      if err != nil {
         return err
      }
      song, err := result.LastInsertId()
      if err != nil {
         return err
      }
      // ITERATE ARTIST ARRAY
      for _, artist := range artists {
         _, err := tx.Exec(`
         INSERT INTO song_artist_t VALUES (?, ?)
         `, song, artist)
         if err != nil {
            return err
         }
      }
   }
   return nil
}

func note(t musicbrainz.Track) string {
   dur := t.Duration()
   if dur == 0 {
      return "?:??"
   }
   if dur < 179_500 * time.Millisecond {
      return "short"
   }
   if dur > 15 * time.Minute {
      return "long"
   }
   return ""
}

type title_note struct {
   title string
   note string
}

func new_release(release, raw_group string) (*musicbrainz.Release, error) {
   if release != "" {
      return musicbrainz.New_Release(release)
   }
   group, err := musicbrainz.New_Release_Group(raw_group)
   if err != nil {
      return nil, err
   }
   group.Sort()
   return group.Releases[0], nil
}
