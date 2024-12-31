package main

import (
	"encoding/json"
	"fmt"
	"github.com/hashicorp/go-memdb"
	"io/ioutil"
	"log"
	"net/http"
)

const PORT = ":8080"

type Article struct {
	Id      uint   `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content`
	Author  string `json:"author"`
}

type Author struct {
	Id       uint     `json:"id"`
	Name     string   `json:"name"`
	Subjects []string `json:"subjects"`
}

func (a *Author) SaveDB(db *memdb.MemDB) (err error) {
	txn := *db.Txn(true)
	if err := txn.Insert("author", a); err != nil {
		return err
	}
	txn.Commit()
	return nil
}

func main() {
	schema := buildSchema()
	db, err := memdb.NewMemDB(schema)
	if err != nil {
		log.Panicln("Error on starting schema db...", err)
	}

	mux := http.NewServeMux()

	fmt.Println("Server starting on port 8080...")

	mux.HandleFunc("GET /healthcheck", aliveChecker)

	mux.HandleFunc("POST /authors", func(w http.ResponseWriter, r *http.Request) {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "Error reading body", http.StatusBadRequest)
			return
		}

		var author Author
		if err := json.Unmarshal(body, &author); err != nil {
			http.Error(w, "Error unmarshaling body", http.StatusBadRequest)
			return
		}

		fmt.Println(author)
		author.SaveDB(db)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(author)
	})

	if err := http.ListenAndServe(PORT, mux); err != nil {
		log.Panicln("Error on starting server...", err)
	}
}

func aliveChecker(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Current I'm alive\n")
}

func buildSchema() *memdb.DBSchema {
	return &memdb.DBSchema{
		Tables: map[string]*memdb.TableSchema{
			"author": &memdb.TableSchema{
				Name: "author",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "Id"},
					},
					"subjects": &memdb.IndexSchema{
						Name:    "subjects",
						Unique:  false,
						Indexer: &memdb.StringSliceFieldIndex{Field: "Subjects", Lowercase: false},
					},
				},
			},
			"article": &memdb.TableSchema{
				Name: "article",
				Indexes: map[string]*memdb.IndexSchema{
					"id": &memdb.IndexSchema{
						Name:    "id",
						Unique:  true,
						Indexer: &memdb.UintFieldIndex{Field: "Id"},
					},
					"title": &memdb.IndexSchema{
						Name:    "title",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Title", Lowercase: false},
					},
					"content": &memdb.IndexSchema{
						Name:    "content",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Content", Lowercase: false},
					},
					"author": &memdb.IndexSchema{
						Name:    "author",
						Unique:  false,
						Indexer: &memdb.StringFieldIndex{Field: "Author", Lowercase: false},
					},
				},
			},
		},
	}
}
