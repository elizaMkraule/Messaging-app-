// This package defines databases and http methods that require their use.
package database

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/docAndColl"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/validator"
	"github.com/santhosh-tekuri/jsonschema"
)

// Defines a database struct that contains a skiplist of documents
type Database struct {
	Mu          sync.Mutex
	Name        string
	URI         []byte
	DocumentMap map[string]*docAndColl.Document // Map of document IDs to document instances
	DocSkipList skiplist.List[string, *docAndColl.Document]
}

// Defines a struct that helps with formatting when returning a database
type Format struct {
	Path string               `json:"path"`
	Doc  interface{}          `json:"doc"`
	Meta *docAndColl.Metadata `json:"meta"`
}

// Constructs a new database
func NewDatabase(name string) Database {
	return Database{
		Name:        name,
		DocSkipList: skiplist.NewList[string, *docAndColl.Document]("", "zzz"),
	}
}

// Takes in a document name and attempts to get that document from the database
// Returns the document and true on success, or an empty value and false on failure
func (db *Database) GetDocumentFromDatabase(docName string) (*docAndColl.Document, bool) {
	db.Mu.Lock()
	defer db.Mu.Unlock()

	doc, exist := db.DocSkipList.Find(docName)

	return doc, exist
}

// Formats the database for printing purposes
func (db *Database) DatabaseFormat(w http.ResponseWriter, r *http.Request) {
	slog.Info("DatabaseFormat: " + db.Name)
	queryParams := r.URL.Query()
	interval := queryParams.Get("interval")
	slog.Info("interval", interval)
	start := ""
	end := ""
	if interval != "" {
		trimmedData := strings.Trim(interval, "[]")
		parts := strings.Split(trimmedData, ",")
		if len(parts) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("unable to parse interval request"))
		}
		start = parts[0]
		end = parts[1]

	}

	var dbFormat []Format
	dbFormat = make([]Format, 0)
	results := db.DocSkipList.Query(start, end)

	var docName string
	var document *docAndColl.Document
	for i, _ := range results {
		docName = results[i].Key
		document = results[i].Value
		var data any

		if err := json.Unmarshal(document.Data, &data); err != nil {
			slog.Error("unable to unmarshal data", "error", err)
		}

		output := Format{
			Path: "/" + docName,
			Doc:  data,
			Meta: document.Metadata,
		}

		dbFormat = append(dbFormat, output)

	}
	jsonData, err := json.MarshalIndent(dbFormat, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to marshal document " + db.Name))
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Takes in a writer and a document name and attempts to delete that document from the database
// Writes the appropriate message to the header based on success/failure
func (db *Database) DeleteDocument(w http.ResponseWriter, docName string) {
	slog.Info("DeleteDocument (from Database): " + docName)

	// SKIPLISTS:
	doc, removed := db.DocSkipList.Remove(docName)

	if !removed {
		slog.Info("did not remove document successfully")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"not found"`))
	} else {
		// updating subs after removing
		docAndColl.Update_subscribers(string(doc.URI), doc.Subscribers, "delete", nil)
		slog.Info("removed document successfully, docname: " + doc.Name)
		w.WriteHeader(http.StatusNoContent)
	}
}

// Takes in information on a document and attempts to put the document into the database
// Writes the appropriate header based on success/failure and whether a patch, update, or insertion occurred
func (db *Database) PutDocIntoDatabase(w http.ResponseWriter, r *http.Request, desc []byte, name string, schema *jsonschema.Schema, username string, patch bool) {
	slog.Info("PutDocIntoDatabase: " + db.Name)

	valid, err := validator.Validate(schema, desc)
	if !valid {
		slog.Error("document does not conform to schema")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid document:" + err.Error()))
		return

	}

	newDocument := docAndColl.NewDocument(name, desc)
	meta := docAndColl.NewMetadata(username)

	newDocument.ColSkipList = skiplist.NewList[string, *docAndColl.Collection]("", "zzz")

	// Check if document already exists
	// TODO: wouldnt it make more sense to check this first and then we only need to change the prev_doc.body and metadata without having to make a new document?
	// that way we can ensure that the subscribers stay the same and its overall less error prone?
	slog.Info("checking if doc already exists")
	prev_doc, exists := db.DocSkipList.Find(newDocument.Name)

	if exists {
		slog.Info("prev_doc exists already")
	} else {
		slog.Info("prev_doc does not exist")
	}

	var uri map[string]string
	if r.Method == http.MethodPost {
		uri = map[string]string{
			"uri": r.URL.Path + name,
		}
	} else {
		uri = map[string]string{
			"uri": r.URL.Path,
		}
	}
	jsonData, err := json.MarshalIndent(uri, "", "  ")
	newDocument.URI = jsonData
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to create database " + newDocument.Name + ": url cannot be marshalled"))
	}

	if exists {
		// check timestamp
		queryParams := r.URL.Query()
		timestamp := queryParams.Get("timestamp")
		if timestamp != "" {
			slog.Info("timestamp found")
			// Convert string to int64
			timestamp_num, err := strconv.ParseInt(timestamp, 10, 64)
			if err != nil {
				slog.Error(err.Error())
			}
			if timestamp_num != prev_doc.Metadata.LastModifiedAt {
				slog.Error("timestamps dont match")
				str := fmt.Sprintf("unable to create/replace document: pre-condition timestamp %d doesn't match current timestamp %d ", timestamp_num, prev_doc.Metadata.LastModifiedAt)
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(str))

				// call return as we dont need to change anything
				return
			}

		}

		meta.CreatedAt = prev_doc.Metadata.CreatedAt
		meta.CreatedBy = prev_doc.Metadata.CreatedBy
		newDocument.Subscribers = prev_doc.Subscribers

	}

	newDocument.Metadata = meta

	// SKIPLISTS:

	newDocument.ColSkipList = skiplist.NewList[string, *docAndColl.Collection]("", "zzz")

	// first do the check for updating
	c := func(name string, doc *docAndColl.Document, exists bool) (newValue *docAndColl.Document, err error) {
		// if the node alrady exists (exists == true), then we want to update.
		// if the node does not exist, return the new empty document

		if exists {
			return &newDocument, nil
		} else {
			return &newDocument, nil
		}
	}

	slog.Info("running Upsert")

	updating, err := db.DocSkipList.Upsert(newDocument.Name, c)
	if err != nil {
		slog.Error("error after upsert in PutDocIntoCollection:", err)
	}

	if !patch {
		if updating {
			slog.Info("PutDocIntoDatabase: replacing Document", "Name", newDocument.Name)
			docAndColl.Update_subscribers(r.URL.Path, prev_doc.Subscribers, "update", &newDocument)
			w.WriteHeader(http.StatusOK)
			w.Write(newDocument.URI)
		} else {
			slog.Info("PutDocIntoDatabase: creating Document", "Name", newDocument.Name)
			w.WriteHeader(http.StatusCreated)
			w.Write(newDocument.URI)
		}
	}
}
