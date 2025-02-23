package docAndColl

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/validator"
	"github.com/santhosh-tekuri/jsonschema"
)

// this a struct that represents a collection: it contains information of the metadata, path, skip list containing child documents, and current substribbers
type Collection struct {
	Name        string
	Mu          sync.Mutex
	Metadata    *Metadata
	URI         []byte
	DocumentMap map[string]*Document // Map of document IDs to document instances
	Subscribers sync.Map
	DocSkipList skiplist.List[string, *Document]
}

// Constructs a new collection
func NewCollection(name string) Collection {
	return Collection{
		Name:        name,
		DocSkipList: skiplist.NewList[string, *Document]("", "zzz"),
	}
}

// given a pointer to a document finds a collection in that document
func (col *Collection) GetDocumentFromCollection(docName string) (*Document, bool) {
	col.Mu.Lock()
	defer col.Mu.Unlock()

	// SKIPLISTS:
	doc, exist := col.DocSkipList.Find(docName)

	return doc, exist

}

// formats the collection to be written to the response writer in a json format
func (col *Collection) CollectionFormat(w http.ResponseWriter) {
	slog.Info("success")
	slog.Info(col.Name)
	var dbFormat []Format
	dbFormat = make([]Format, 0)
	results := col.DocSkipList.Query("", "")

	var document *Document
	for i, _ := range results {
		document = results[i].Value
		var data any

		if err := json.Unmarshal(document.Data, &data); err != nil {
			slog.Error("unable to unmarshal data", "error", err)
		}
		var jsonMap map[string]string
		json.Unmarshal(document.URI, &jsonMap)
		uri, _ := jsonMap["uri"]
		parts := strings.Split(uri, "/")
		substr := strings.Join(parts[3:], "/")
		substr = "/" + substr
		output := Format{
			Path: substr,
			Doc:  data,
			Meta: document.Metadata,
		}

		dbFormat = append(dbFormat, output)

	}

	jsonData, err := json.MarshalIndent(dbFormat, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to marshal document " + col.Name))
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Delete the given docuemnt
func (col *Collection) DeleteDocument(w http.ResponseWriter, docName string) *Document {
	slog.Info("Delete Document (from Collection): ")
	slog.Info(docName)

	// SKIPLISTS:
	doc, removed := col.DocSkipList.Remove(docName)

	if !removed {
		slog.Info("did not remove document successfully")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"not found"`))
	} else {
		slog.Info("removed document successfully, docname: " + doc.Name)
		w.WriteHeader(http.StatusNoContent)
	}
	return doc
}

// given a collection pointer creates a new document and meta and inserts the document
func (col *Collection) PutDocIntoCollection(w http.ResponseWriter, r *http.Request, desc []byte, name string, schema *jsonschema.Schema, username string, patch bool) {

	slog.Info("database success")
	slog.Info(col.Name)
	// Convert request into a database

	valid, err := validator.Validate(schema, desc)

	if !valid {
		slog.Error("document does not conform to schema")
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("invalid document:" + err.Error()))
		return

	}

	newDocument := NewDocument(name, desc)
	metadata := NewMetadata(username)

	slog.Info("before")
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

	newDocument.Metadata = metadata
	slog.Info("after")

	prev_doc, exists := col.DocSkipList.Find(newDocument.Name)
	if exists {
		slog.Info("document exits in collection")
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
	}

	// SKIPLISTS:

	newDocument.ColSkipList = skiplist.NewList[string, *Collection]("", "zzz")

	// first do the check for updating
	c := func(name string, doc *Document, exists bool) (newValue *Document, err error) {
		// if the node alrady exists (exists == true), then we want to update.
		// if the node does not exist, return the new empty document

		// for documents, do we want to return the newDocument no matter what?
		if exists {
			return &newDocument, nil
		} else {
			return &newDocument, nil
		}
	}

	slog.Info("running Upsert")

	updating, err := col.DocSkipList.Upsert(newDocument.Name, c)

	slog.Info("Updating collection subscribers if they exist")
	Update_subscribers(r.URL.Path, &col.Subscribers, "update", &newDocument)

	if err != nil {
		slog.Error("error after upsert in PutDocIntoCollection:", err)
	}
	if !patch {
		if updating {
			slog.Info("PUT document: replacing Document", "Name", newDocument.Name)
			w.WriteHeader(http.StatusOK)
			w.Write(newDocument.URI)
		} else {
			slog.Info("PUT document: creating Document", "Name", newDocument.Name)
			w.WriteHeader(http.StatusCreated)
			w.Write(newDocument.URI)
		}
	}
}
