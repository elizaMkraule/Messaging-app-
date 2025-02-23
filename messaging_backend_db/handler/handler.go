// Package handler encapsulates functions handling GET, POST, PUT, DELETE, PATCH and OPTIONS requests.
package handler

import (
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/authorize"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/docAndColl"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/parser"
	"github.com/santhosh-tekuri/jsonschema"
)

// This is the main fucntion for the handler requests, it determines which method is called, calles differnet functions to parse the body and path
// it also calls helper methods to change the database, docoments and collumns as well as some general error handling
func HndlRequest(w http.ResponseWriter, r *http.Request, owlDB *database_host.Database_host, tokenmap *sync.Map, schema *jsonschema.Schema) {
	// Handle incoming HTTP requests here

	// Athorize all incoming requests
	slog.Info("authorize")
	flag, username := authorize.Authorize(w, r, tokenmap)
	if !flag {
		// if flag is false the request couldnt be authorized/authentificated therefore we cant do anything
		return
	}
	slog.Info("after authroize")
	switch r.Method {

	case http.MethodOptions:
		slog.Info("IN OPTIONS")
		w.Header().Set("Allow", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Origin", "*") // TODO: should patch be added?
		w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.WriteHeader(http.StatusOK)

	case http.MethodGet:

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		slog.Info("IN GET")

		// call parser to get fields
		segments, stopPoint := parser.ParseURL(r.URL.Path, false)
		parse := GetValid(segments, owlDB, stopPoint)

		// CHECK IF MODE IS SUBSCRIBE
		queryParams := r.URL.Query()
		mode := queryParams.Get("mode")

		slog.Info("mode is ", mode)

		// finding obj in system to return
		// should I convert to JSON obj? and how do you return the value
		if parse.Exist {
			switch parse.ObjType {
			case "database":
				slog.Info("case database")
				if !hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"bad resource path"`))
				} else {
					parse.Database.DatabaseFormat(w, r)
				}
			case "document":
				slog.Info("case doc")
				if string(mode) == "subscribe" {
					slog.Info("mode is subcribe")
					docAndColl.CreateSubscriber(r.URL.Path, w, r, parse.Document.Subscribers)
				}
				if hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"bad resource path"`))
				} else {
					parse.Document.DocumentFormat(w)
				}
			case "collection":
				slog.Info("case col")
				if mode == "subscribe" {
					docAndColl.CreateSubscriber(r.URL.Path, w, r, &parse.Collection.Subscribers)
				}
				if !hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"bad resource path"`))
				} else {
					parse.Collection.CollectionFormat(w)
				}
			}
		} else {
			slog.Error("url given does not exist in system")

			if parse.ObjType == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`"bad resource path"`))
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`"` + parse.ObjType + `"`))
			}
		}

	case http.MethodPut:
		slog.Info("IN PUT")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// read the body
		desc, err := io.ReadAll(r.Body)
		//desc1, err1 := io.ReadAll(r.URL.Path)

		defer r.Body.Close()
		if err != nil {
			slog.Error("PUT database: error reading database request", "error", err)
			http.Error(w, `"invalid database format"`, http.StatusBadRequest)
			return
		}

		// call parser for putting
		segments, stopPoint := parser.ParseURL(r.URL.Path, true)
		parse := PutValid(segments, owlDB, stopPoint)
		w.Header().Set("Content-Type", "application/json")
		if parse.Exist {
			switch parse.ObjType {
			case "server":
				slog.Info("validate the database")
				if hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to create collection: bad resource path"`))
				} else {
					owlDB.PutDatabaseIntoServer(owlDB, w, r, parse.Name)
				}
			case "database":
				// this puts doc into db
				slog.Info("case db")
				if hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to create document: bad resource path"`))
				} else {
					parse.Database.PutDocIntoDatabase(w, r, desc, parse.Name, schema, username, false)
				}
			case "document":
				if !hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to create collection: bad resource path"`))
				} else {
					parse.Document.PutColIntoDocument(w, r, parse.Name)
				}

			case "collection":
				slog.Info("case coll")
				if hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to create document: bad resource path"`))
				} else {
					parse.Collection.PutDocIntoCollection(w, r, desc, parse.Name, schema, username, false)
				}

			}
		} else {
			slog.Error("url given does not exist in system")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`"` + parse.ObjType + `"`))
		}
	case http.MethodPost:
		slog.Info("IN POST")
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		// read the body
		desc, err := io.ReadAll(r.Body)
		//desc1, err1 := io.ReadAll(r.URL.Path)

		defer r.Body.Close()
		if err != nil {
			slog.Error("PUT database: error reading database request", "error", err)
			http.Error(w, `"invalid database format"`, http.StatusBadRequest)
			return
		}

		// IF POST IS USED FOT AUTHENTIFICATION
		if r.URL.Path == "/auth" {
			slog.Info("POST auth")
			authorize.Authenticate(w, r, desc, tokenmap)
			return
		}

		// IF POST IS USED FOT NORMAL OPS --> add logic
		// Changing to Put because I want the parent. This differentiates between whether the post is happening on a database or a collection
		slog.Info("POST before parse")
		segments, stopPoint := parser.ParseURL(r.URL.Path, false)
		parse := GetValid(segments, owlDB, stopPoint)
		slog.Info("POST after parse")
		// is setting the header necessary?
		w.Header().Set("Content-Type", "application/json")

		// generate random string for post name
		const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
		seed := rand.NewSource(time.Now().UnixNano()) // TODO: should i use a random integer as a seed instead of the current time
		random := rand.New(seed)
		randString := make([]byte, 14)
		for i := range randString {
			randString[i] = charset[random.Intn(len(charset))]
		}
		slog.Info("POST exist, ", parse.Exist)
		// finding obj in system to return
		if parse.Exist {
			slog.Info("POST objtype, ", parse.ObjType)
			// doc has two POST's. One for posting into a database and one for posting into a collection
			switch parse.ObjType {
			case "database":
				// post is essentially a put without a name, NEED TO GENERATE A RANDOM UNIQUE STRING
				parse.Database.PutDocIntoDatabase(w, r, desc, string(randString), schema, username, false)
			case "collection":
				parse.Collection.PutDocIntoCollection(w, r, desc, string(randString), schema, username, false)
			default:
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`"document cannot be posted: bad resource path"`))
			}

		} else {
			slog.Error("url given does not exist in system")

			if parse.ObjType == "" {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(`"bad resource path"`))
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`"` + parse.ObjType + `"`))
			}
		}

	case http.MethodDelete:
		slog.Info("IN DELETE")

		// IF DELETE IS USED FOR DELETING AUTHENTIFICATION
		if r.URL.Path == "/auth" {
			authorize.Delete(w, r, tokenmap)
			return
		}

		// IF DELETE IS USED FOR DELETING MAP DOC COL

		// call parser to get fields
		segments, stopPoint := parser.ParseURL(r.URL.Path, true)
		parse := PutValid(segments, owlDB, stopPoint)

		// finding obj in system to Delete and Delete it
		if parse.Exist {
			switch parse.ObjType {
			case "server":
				// delete the database from the server
				if hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to delete database: bad resource path"`))
				} else {
					slog.Info("Starting delete database from server")
					owlDB.DeleteDatabase(w, parse.Name)
				}
			case "database":
				// delete the document from the database
				// check if the docuemnt has a subs
				if hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to delete document: bad resource path"`))
				} else {
					slog.Info("Starting delete document from database")
					parse.Database.DeleteDocument(w, parse.Name)
				}
			case "document":
				// delete collection from the document
				// check if the collection has a subs
				if !hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to delete collection: bad resource path"`))
				} else {
					slog.Info("Starting delete collection from doc")
					parse.Document.DeleteCollection(w, parse.Name)
				}
			case "collection":
				slog.Info("Starting delete doc from collection")
				// delete document from the collection
				if hasEndSlash(r.URL.Path) {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte(`"unable to delete document: bad resource path"`))
				} else {
					doc := parse.Collection.DeleteDocument(w, parse.Name)
					// check if the docuemnt has a subs
					slog.Info("updating collection subscribers about delete event")
					docAndColl.Update_subscribers(r.URL.Path, &parse.Collection.Subscribers, "delete", doc)
					slog.Info("updating docuemnt subscribers about delete event")
					docAndColl.Update_subscribers(r.URL.Path, doc.Subscribers, "delete", doc)
				}
			}

		} else {
			slog.Error("url given does not exist in system")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`"` + parse.ObjType + `"`))
		}
	case http.MethodPatch:
		// call parser to get fields
		slog.Info("patch")
		segments, stopPoint := parser.ParseURL(r.URL.Path, false)
		parse := GetValid(segments, owlDB, stopPoint)
		desc, _ := io.ReadAll(r.Body)

		if parse.Exist {
			if parse.ObjType == "document" {
				patched_document, _ := parse.Document.Patch(w, r, desc, schema)
				// call parser for putting
				segments, stopPoint := parser.ParseURL(r.URL.Path, true)
				parse := PutValid(segments, owlDB, stopPoint)
				if parse.Exist {
					switch parse.ObjType {

					case "database":
						// this puts doc into db
						parse.Database.PutDocIntoDatabase(w, r, patched_document, parse.Name, schema, username, true)
					// updating happens when its put
					// docAndColl.Update_subscribers(r.URL.Path, parse.Document.Subscribers, "update", parse.Document)

					case "collection":
						slog.Info("case coll")
						parse.Collection.PutDocIntoCollection(w, r, patched_document, parse.Name, schema, username, true)
						// updating happens when its put
						// docAndColl.Update_subscribers(r.URL.Path, &parse.Collection.Subscribers, "update", parse.Document)

					}
				} else {
					slog.Error("url given does not exist in system")
					w.WriteHeader(http.StatusNotFound)
					w.Write([]byte(`""document not found:` + parse.ObjType + `"`))
				}

			} else {
				slog.Error("url given does not exist in system")
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(`""document not found:` + parse.ObjType + `"`))
			}

		} else {
			slog.Error("url given does not exist in system")
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`"document not found:` + parse.ObjType + `"`))
		}

	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("unable to parse url"))
	}
}

// helper function that helps with some general error handeling of the path
func hasEndSlash(path string) bool {
	r := []rune(path)
	lastRune := r[len(r)-1]
	b := false
	if lastRune == '/' {
		b = true
	}
	return b
}

// Called to see if the parsed URL is valid for a get the database, collection or document
// this is called when we want to create a get request or generally need an object and want to use the entire get path
func GetValid(segments []string, server *database_host.Database_host, stopPoint int) *Parsed {

	var prev string
	var database *database.Database
	var document *docAndColl.Document
	var collection *docAndColl.Collection

	//checking to see if there is a slash at the end

	expecting := "not found"

	for i, word := range segments {
		slog.Info("word ", word)
		if i >= stopPoint {
			break
		}
		slog.Info(word)
		if i == 0 || i == 1 {
			continue
		} else if i == 2 {

			db, exist := server.GetDatabase(word)
			database = db
			prev = "database"
			if !exist {
				slog.Info("returning, ", expecting)
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse
			}
		} else if word == "" {
			parse := NewParsed(false, "", "", nil, nil, nil)
			return &parse

		} else if prev == "database" {

			// following function maybe has error because its a map of ID->doc not Name->doc ??
			doc, exist := database.GetDocumentFromDatabase(word)
			document = doc
			prev = "document"
			if !exist {
				expecting = "unable to retrieve document " + word + ": not found"
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse
			}
		} else if prev == "document" {
			col, exist := document.GetCollection(word)
			collection = col
			prev = "collection"
			if !exist {
				expecting = "unable to retrieve collection " + word + ": not found"
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse

			}
		} else if prev == "collection" {
			doc, exist := collection.GetDocumentFromCollection(word)
			document = doc
			prev = "document"
			if !exist {
				expecting = "unable to retrieve document " + word + ": not found"
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse

			}
		} else {
			parse := NewParsed(false, expecting, "", nil, nil, nil)
			return &parse
		}
	}
	if prev == "database" {
		parse := NewParsed(true, "database", "", database, nil, nil)
		return &parse
	}
	if prev == "document" {
		parse := NewParsed(true, "document", "", nil, document, nil)
		return &parse
	}
	if prev == "collection" {
		parse := NewParsed(true, "collection", "", nil, nil, collection)
		return &parse
	}
	parse := NewParsed(false, "", "", nil, nil, nil)
	return &parse

}

// Checks if the parsed URL is valid when you want to insert a document, collection or database
// called when you want to get an object to put a request in, there are different error messages seperate from get and
// process the path in a differnet way such that it returns the pointer to the value that needs to be returned
func PutValid(segments []string, owlsDB *database_host.Database_host, stopPoint int) *Parsed {

	var prev string
	var database *database.Database
	var document *docAndColl.Document
	var collection *docAndColl.Collection

	if stopPoint == 2 {
		parse := NewParsed(true, "server", segments[stopPoint], nil, nil, nil)
		return &parse
	}
	expecting := ""
	if stopPoint%2 == 0 {
		expecting = "unable to create collection: "

	} else {
		expecting = "unable to create/replace document: "
	}

	for i, word := range segments {
		if i >= stopPoint {
			break
		}
		if i == 0 || i == 1 {
			continue
		} else if i == 2 {
			tempdb, exist := owlsDB.GetDatabase(word)
			//database = *tempdb
			database = tempdb
			prev = "database"
			if !exist {
				slog.Info("database DNE")
				expecting += "not found"
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse
			}
		} else if prev == "database" {
			doc, exist := database.GetDocumentFromDatabase(word)
			//document = *doc
			document = doc
			prev = "document"
			if !exist {
				expecting += "unable to retrieve document " + word + ": not found"
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse
			}
		} else if prev == "document" {
			col, exist := document.GetCollection(word)
			//collection = *col
			collection = col
			prev = "collection"
			if !exist {
				expecting += "unable to retrieve collection " + word + ": not found"
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse
			}
		} else if prev == "collection" {
			doc, exist := collection.GetDocumentFromCollection(word)
			//document = *doc
			document = doc
			prev = "document"
			if !exist {
				expecting += "unable to retrieve document " + word + ": not found"
				parse := NewParsed(false, expecting, "", nil, nil, nil)
				return &parse
			}
		} else {
			parse := NewParsed(false, "", "", nil, nil, nil)
			return &parse
		}
	}
	if prev == "database" {
		slog.Info("returning a database from parseURLPut")
		parse := NewParsed(true, "database", segments[stopPoint], database, nil, nil)
		//parse := NewParsed(true, "database", segments[b], &database, nil, nil)
		return &parse
	}
	if prev == "document" {
		slog.Info("returning a document from parseURLPut")
		parse := NewParsed(true, "document", segments[stopPoint], nil, document, nil)
		//parse := NewParsed(true, "document", segments[b], nil, &document, nil)
		return &parse
	}
	if prev == "collection" {
		slog.Info("returning a collection from parseURLPut")
		parse := NewParsed(true, "collection", segments[stopPoint], nil, nil, collection)
		//parse := NewParsed(true, "collection", segments[b], nil, nil, &collection)
		return &parse
	}
	parse := NewParsed(false, "", "", nil, nil, nil)
	return &parse

}

// struct that contains all of the information the handler needs from the parse
type Parsed struct {
	Exist      bool
	ObjType    string
	Name       string
	Database   *database.Database
	Document   *docAndColl.Document
	Collection *docAndColl.Collection
}

// called when a Parsed struct needs to be created
func NewParsed(
	exist bool,
	ObjType string,
	Name string,
	Db *database.Database,
	Doc *docAndColl.Document,
	Collection *docAndColl.Collection,

) Parsed {
	return Parsed{Exist: exist, ObjType: ObjType, Name: Name, Database: Db, Document: Doc, Collection: Collection}
}
