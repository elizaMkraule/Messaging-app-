package Testing

import (
	"sync"
	"testing"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/santhosh-tekuri/jsonschema"
)

// testing with fuzz testing
func FuzzDBPut(f *testing.F) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:3318/v1/db"

	f.Fuzz(func(t *testing.T, data []byte) {
		// Assuming your getBearerToken also takes a *testing.T
		token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
		doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)
		docURL := "http://localhost:3318/v1/db/doc"
		requestBody := string(data)
		w := doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)
		CheckResponse(t, w, 201, "put_doc.json")

	})
}

// tests deleting a database
func TestDbDelete204(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db"

	// Execute the GET request and receive the response.
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	if w.Code != 204 {
		t.Errorf("Expected status code %d, got %d", 204, w.Code)
	}

}

// throw error if the database to be deleted is incorrect
func TestDbDelete404(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db"

	// Execute the GET request and receive the response.
	doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 404, "get_db_404.json")

}

// if user cannot be allowed to delete
func TestDbDelete401(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	schema := new(jsonschema.Schema)
	URL := "http://localhost:4318/v1/db/dpc"

	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
	doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	// Perform PUT request and get the recorder
	docURL := "http://localhost:3318/v1/db/doc"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)
	token = "token"
	getURL := "http://localhost:3318/v1/db"
	w := doDeleteRequest(t, getURL, token, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 401, "missingToken.json")
}

// deleteion is sucesfful on documents
func TestDocDelete204(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/doc"

	// Execute the GET request and receive the response.
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	if w.Code != 204 {
		t.Errorf("Expected status code %d, got %d", 204, w.Code)
	}

}

// throws an error if the delete function cannot find document
func TestDocDelete404db(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/doc"
	// Execute the GET request and receive the response.
	doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)
	getURL = "http://localhost:3318/v1/d/doc"
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 404, "get_db_404.json")

}

// cannot find document to delete but database in path exists
func TestDocDelete404doc(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/doc"

	// Execute the GET request and receive the response.
	doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 400, "get_db_404.json")

}
