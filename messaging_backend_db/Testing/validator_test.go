// this is a Testing suite that compares different https writer responses against a json file each test
// is given different handler requests and we store the response in a response writter to compare
package Testing

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/handler"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/santhosh-tekuri/jsonschema"
)

// helper fuction that tests a post request
func doPostRequest(t *testing.T, url, token, requestBody string, owlDB *database_host.Database_host, tokenMap *sync.Map, subscribers *sync.Map, schema *jsonschema.Schema) *httptest.ResponseRecorder {
	t.Helper() // Marks the calling function as a test helper function.

	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	fmt.Print("everything set up")
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			fmt.Print("entering into handler")
			// handler.HndlRequest would be your actual handler that you want to test.
			handler.HndlRequest(w, r, owlDB, tokenMap, schema)
		})
	handler(w, req)

	return w
}

// these tests focus on delte, post and patch requests
func doDeleteRequest(t *testing.T, url, token string, owlDB *database_host.Database_host, tokenMap *sync.Map, subscribers *sync.Map, schema *jsonschema.Schema) *httptest.ResponseRecorder {
	t.Helper()

	// Create a new HTTP GET request.
	req := httptest.NewRequest("DELETE", url, nil)
	req.Header.Set("accept", "*/*")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}

	// Create a ResponseRecorder to record the response.
	w := httptest.NewRecorder()

	// Define the handler function.
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {

			handler.HndlRequest(w, r, owlDB, tokenMap, schema)
		})

	// Serve the request with our handler.
	handler.ServeHTTP(w, req)

	return w
}

// collection delete is not sucessful as database does not exists
func TestColDelete404db(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.

	getURL := "http://localhost:4318/v1/d/dc%2Fcol/"
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 404, "get_db_404.json")

}

// collection delete does not work as document does not exists
func TestColDelete404doc(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/do%2Fcol/"

	// Execute the GET request and receive the response.
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 404, "get_doc_404.json")

}

// collection delete does not work as the collection does not exist
func TestColDelete404col(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/doc%2Fcol/"

	// Execute the GET request and receive the response.
	doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 404, "get_db_404.json")

}

// testing patch on a document where the return is false

// testing incorrect url path for a post request
func TestDbPost404(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/d/"
	requestBody :=
		`{
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	  }`
	w := doPostRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 404, "get_db_404.json")
}

// unaurthorized user tries to post in the database
func TestDbPost401(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	schema := new(jsonschema.Schema)
	URL := "http://localhost:4318/v1/db"

	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
	doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	token = "token"
	docURL := "http://localhost:3318/v1/db/"
	requestBody :=
		`{
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	  }`
	w := doPostRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 401, "missingToken.json")
}
