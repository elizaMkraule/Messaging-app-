package docAndColl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/jsonPatch"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/jsonvisit"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/validator"

	"github.com/santhosh-tekuri/jsonschema"
)

// writeFlusher is used to nofity the subscribers about changes.
type writeFlusher interface {
	http.ResponseWriter
	http.Flusher
}

// Document struct represents a document: it contains information of the metadata, path, skip list containing child collections, and current subscribers
type Document struct {
	Name          string
	Metadata      *Metadata
	Mu            sync.Mutex
	ID            string
	Data          []byte
	URI           []byte
	CollectionMap map[string]*Collection // Map of collection names to collection instances
	ColSkipList   skiplist.List[string, *Collection]
	Subscribers   *sync.Map
}

// Metadata type structure represents the metadata this struct is used in database, colleciton and document to hold their respective metadata
type Metadata struct {
	CreatedAt      int64  `json:"createdAt"`
	CreatedBy      string `json:"createdBy"`
	LastModifiedAt int64  `json:"lastModifiedAt"`
	LastModifiedBy string `json:"lastModifiedBy"`
}

// creates a new metadata
func NewMetadata(name string) *Metadata {
	return &Metadata{time.Now().UnixMilli(), name, time.Now().UnixMilli(), name}
}

// this is a struct that hold all the information to be marshalled for the response writer
type Format struct {
	Path string      `json:"path"`
	Doc  interface{} `json:"doc"`
	Meta *Metadata   `json:"meta"`
}

// PatchResponse type struct is used to format a repsonse for the client upon a Patch request.
type PatchResponse struct {
	Uri         string `json:"uri"`
	PatchFailed bool   `json:"patchFailed"`
	Message     string `json:"message"`
}

// NewPatchRespons constructs a new patch response.
func NewPatchResponse(uri string, PatchFailed bool, Message string) PatchResponse {
	return PatchResponse{uri, PatchFailed, Message}
}

// Constructs a new document
func NewDocument(name string, data []byte) Document {
	return Document{
		Name:        name,
		Data:        data,
		ColSkipList: skiplist.NewList[string, *Collection]("", "zzz"),
		Subscribers: new(sync.Map),
	}
}

// this retruns a pointer to the colleciton that we are looking for
func (doc *Document) GetCollection(colName string) (*Collection, bool) {
	doc.Mu.Lock()
	defer doc.Mu.Unlock()

	// SKIPLISTS:
	cols, exist := doc.ColSkipList.Find(colName)

	return cols, exist

}

// This gets the inputted document. Essentially the GET function for Documents.
func (doc *Document) DocumentFormat(w http.ResponseWriter) {
	slog.Info("success")
	slog.Info(doc.Name)
	w.WriteHeader(http.StatusOK)

	var data any
	if err := json.Unmarshal(doc.Data, &data); err != nil {
		slog.Error("unable to unmarshal data", "error", err)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to unmarshal document " + doc.Name))
	}
	var jsonMap map[string]string
	json.Unmarshal(doc.URI, &jsonMap)
	uri, _ := jsonMap["uri"]
	parts := strings.Split(uri, "/")
	substr := strings.Join(parts[3:], "/")
	output := Format{
		Path: "/" + substr,
		Doc:  data,
		Meta: doc.Metadata,
	}
	slog.Info("output ", output)
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to marshal document " + doc.Name))
	}
	w.WriteHeader(http.StatusOK)
	w.Write(jsonData)
}

// Delete the collection from the document
func (doc *Document) DeleteCollection(w http.ResponseWriter, colName string) {
	slog.Info("Delete Collection: ")
	slog.Info(colName)

	// SKIPLISTS:
	col, removed := doc.ColSkipList.Remove(colName)

	if !removed {
		slog.Info("did not remove collection successfully")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`"not found"`))
	} else {
		slog.Info("removed collection successfully, colname: " + col.Name)
		slog.Info("updating collection subscribers about delete event")
		Update_subscribers(string(col.URI), &col.Subscribers, "delete", doc)
		w.WriteHeader(http.StatusNoContent)
		w.Write([]byte(`"bad resource path"`))
	}
}

// puts a collection given a document and creates the collection struct
func (doc *Document) PutColIntoDocument(w http.ResponseWriter, r *http.Request, name string) {

	slog.Info("collection success")
	slog.Info(name)
	// Convert request into a database
	newCollection := NewCollection(name)
	metadata := NewMetadata("username")

	slog.Info("before")
	uri := map[string]string{
		"uri": r.URL.Path,
	}
	jsonData, err := json.MarshalIndent(uri, "", "  ")
	newCollection.URI = jsonData
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to create collection " + newCollection.Name + ": url cannot be marshalled"))
	}

	newCollection.Metadata = metadata
	slog.Info("after")

	//Lock database before writting to it
	doc.Mu.Lock()
	defer doc.Mu.Unlock()

	// SKIPLISTS:

	newCollection.DocSkipList = skiplist.NewList[string, *Document]("", "zzz")

	// first do the check for updating
	c := func(name string, col *Collection, exists bool) (newValue *Collection, err error) {
		// if the node already exists, we cannot update it. return an error
		// else, if the node does not exists, insert the collection

		if exists {
			return nil, fmt.Errorf("error in check of PutColIntoDocument, collection already exists")
			//return &newCollection, nil
		} else {
			return &newCollection, nil
		}
	}

	slog.Info("running Upsert")

	updating, err := doc.ColSkipList.Upsert(newCollection.Name, c)
	if err != nil {
		slog.Error("error after upsert in PutDocIntoCollection:", err)
	}

	if updating {
		slog.Info("PUT Collection: replacing collection", "Name", newCollection.Name)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`"unable to create collection: exists"`))
	} else {
		slog.Info("PUT Collection: creating collection", "Name", newCollection.Name)
		w.WriteHeader(http.StatusCreated)
		w.Write(newCollection.URI)
	}
}

// This is called when the handler request detects a patch method. This will create a new partch response
// and try and execture the patch, throwing an error if it does not work
func (doc *Document) Patch(w http.ResponseWriter, r *http.Request, data []byte, schema *jsonschema.Schema) ([]byte, error) {

	// set to default values
	newdoc := doc.Data
	message := "patch applied"

	// get all the patches
	slog.Info("unmarshal patch")
	var patches []map[string]interface{}
	if err := json.Unmarshal(data, &patches); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Invalid patches: " + err.Error()))
		return doc.Data, err
	}

	for _, patch := range patches {

		op, opExists := patch["op"].(string)
		val, valueExists := patch["value"]
		path, pathExists := patch["path"].(string)

		// if missing fields for one patch
		if !opExists || !valueExists || !pathExists {
			slog.Info("Patch did not have all necessary values")
			sendPatchResponse(w, http.StatusOK, NewPatchResponse(r.URL.Path, true, "Each patch object must have 'op', 'value', and 'path' fields."))
			return doc.Data, nil
		}
		var err error
		if err != nil {
			sendPatchResponse(w, http.StatusOK, NewPatchResponse(r.URL.Path, true, fmt.Sprintf("Validation failed for patch operation: %v", err)))
			return doc.Data, nil
		}
		// if no errors applyPatch
		newdoc, err = applyPatch(path, val, newdoc, op)
		if err != nil {
			slog.Error(err.Error())
			sendPatchResponse(w, http.StatusOK, NewPatchResponse(r.URL.Path, true, err.Error()))
			return doc.Data, nil
		}

	}

	// validate
	valid, err := validator.Validate(schema, newdoc)

	if !valid {
		slog.Error("document does not conform to schema")
		sendPatchResponse(w, http.StatusOK, NewPatchResponse(r.URL.Path, true, fmt.Sprintf("patched document is invalid: %v", err.Error())))
		return doc.Data, nil
	}

	sendPatchResponse(w, http.StatusOK, NewPatchResponse(r.URL.Path, false, message))

	return newdoc, nil
}

// This is a helper fucntion for patch the trys to apply the diven patchs given by the body of the handler request
func applyPatch(path string, value interface{}, data []byte, op string) ([]byte, error) {

	slog.Info("op code:", op)

	var docMap map[string]interface{}
	if err := json.Unmarshal(data, &docMap); err != nil {
		// Handle error, return data as is
		return data, err
	}
	var jsonpatch jsonPatch.JsonPatchVisitor

	components := strings.Split(path, "/")[1:]
	lastIndex := len(components) - 1
	newkey := ""
	var new_components []string
	if op == "ObjectAdd" {
		newkey = components[lastIndex]
		slog.Info("newkey:", newkey)
		new_components = components[:len(components)-1]
		for _, component := range new_components {
			slog.Info("new array:", component)
		}
		jsonpatch = jsonPatch.New(new_components, value, op, newkey)
	} else {

		jsonpatch = jsonPatch.New(components, value, op, newkey)
	}

	// Validate the patch docMap
	res, err := jsonvisit.Accept(docMap, jsonpatch)

	slog.Info(fmt.Sprintf("patched%t", res))

	if err != nil {
		slog.Error(err.Error())
		slog.Info("Result:", res)
		return nil, err
	}

	// Marshal the modified map back to JSON
	resultData, err := json.Marshal(res)
	if err != nil {
		return data, err
	}

	slog.Info("Document after patch:", string(resultData))
	return resultData, nil
}

// this formats the patch response after the patch has been attempted and sends the marshalled data to the response writer
func sendPatchResponse(w http.ResponseWriter, statusCode int, output PatchResponse) {
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal server error: unable to marshal patch response"))
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(jsonData)
}

// this updates the substribers when a new subscriber is added to a specific document or collection
func Update_subscribers(path string, subscribers *sync.Map, event string, doc *Document) {

	// Check if the map is empty
	isEmpty := true
	subscribers.Range(func(key, value interface{}) bool {
		isEmpty = false
		return true
	})

	if isEmpty {
		return
	}

	var eventData string

	switch event {
	case "delete":

		// Delete event format
		eventID := time.Now().UnixNano()
		eventData = fmt.Sprintf("event: delete\ndata: %s\nid: %d\n\n", path, eventID)

		slog.Info("Updating each subscriber about delete event for", "path", path, "eventID", eventID)

	case "update":

		var data any
		if err := json.Unmarshal(doc.Data, &data); err != nil {
			slog.Error("unable to unmarshal data", "error", err)
		}

		var jsonMap map[string]string
		json.Unmarshal(doc.URI, &jsonMap)
		uri := jsonMap["uri"]
		parts := strings.Split(uri, "/")
		substr := strings.Join(parts[3:], "/")
		substr = "/" + substr

		output := Format{
			Path: substr,
			Doc:  data,
			Meta: doc.Metadata,
		}

		jsonData, err := json.MarshalIndent(output, "", "  ")
		if err != nil {
			slog.Error(err.Error())
		}

		//event format
		eventID := time.Now().UnixNano()
		eventData = fmt.Sprintf("event: update\ndata: %s\nid: %d\n\n", jsonData, eventID)

		slog.Info("Updating each subscriber about update event for", "path", path, "eventID", eventID)
	}

	// going through the subscribers
	subscribers.Range(func(key, value interface{}) bool {
		if subscriber, ok := key.(writeFlusher); ok {
			subscriber.Write([]byte(eventData))
			subscriber.Flush()
			var evt bytes.Buffer
			evt.WriteString("\n")
			slog.Info("Now Sending the empty line", "msg", evt.String())
			// Send event
			subscriber.Write(evt.Bytes())
			subscriber.Flush()
		}
		return true
	})
}

// this creates a new subscriber and only occurs once per subsriber
func CreateSubscriber(path string, w http.ResponseWriter, r *http.Request, subscribers *sync.Map) {

	slog.Info("Mode = subscribe. Processing server-sent events")
	// Handle server-sent events logic here

	// ResponseWriter ==> writeFlusher
	wf, ok := w.(writeFlusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	slog.Info("Converted to writeFlusher")

	// Set up event stream connection
	wf.Header().Set("Content-Type", "text/event-stream")
	wf.Header().Set("Cache-Control", "no-cache")
	wf.Header().Set("Connection", "keep-alive")
	wf.Header().Set("Access-Control-Allow-Origin", "*")
	wf.WriteHeader(http.StatusOK)
	wf.Flush()

	slog.Info("Sent headers")

	slog.Info("store new subscriber to:", path)

	// store subscribers in a mapping of name to slice of subscribers
	// check if a name to subscriber mapping already exists
	if _, ok := subscribers.Load(wf); !ok {
		// a new subscriber
		subscribers.Store(wf, true)

		for {
			select {
			case <-r.Context().Done():
				// Client closed connection
				slog.Info("Client closed connection")
				return
			case <-time.After(15 * time.Second): // Send a comment line every 15 seconds to prevent connection timeout
				var evt bytes.Buffer
				evt.WriteString("\n")
				slog.Info("Sending", "msg", evt.String())
				// Send event
				wf.Write(evt.Bytes())
				wf.Flush()
			}
		}
	}
	// if ok the subscriber already exists so we do not need to do anything

}
