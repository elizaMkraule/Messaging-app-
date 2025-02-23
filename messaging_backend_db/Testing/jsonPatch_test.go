// this is a Testing suite that compares different https writer responses against a json file each test
// is given different handler requests and we store the response in a response writter to compare
package Testing

import (
	"bytes"
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

// helper function that creates a patch request
func doPatchRequest(t *testing.T, url, token, requestBody string, owlDB *database_host.Database_host, tokenMap *sync.Map, subscribers *sync.Map, schema *jsonschema.Schema) *httptest.ResponseRecorder {
	t.Helper() // Marks the calling function as a test helper function.

	req, err := http.NewRequest("PATCH", url, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		t.Fatalf("Could not create request: %v", err)
	}
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			// handler.HndlRequest would be your actual handler that you want to test.
			handler.HndlRequest(w, r, owlDB, tokenMap, schema)
		})
	handler(w, req)

	return w
}
func TestDocPatch200false(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/db/doc"
	requestBody :=
		`[
		{
		  "op": "ArrayAdd",
		  "path": "/db/doc",
		  "value": {
			"user": "another_user"
		  }
		}
	  ]`
	w := doPatchRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 404, "put_doc_404.json")
}

// tesitng patch where the return is true
func TestPatch200true(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	docURL := "http://localhost:3318/v1/db/doc"
	requestBody :=
		`{
		"person": {
			"first_name": ["John", "greg", "fin"],
			"last_name": ["Doe", "kennedy"],
			"age": 30,
			"email": "john@example.com",
			"address": "123"
		}
	}`
	doPutDocRequest(t, docURL, token, requestBody, owlDB, tokenMap, subscribers, schema)
	requestBody =
		`[
		{
		  "op": "ArrayAdd",
		  "path": "/person/last_name",
		  "value":  "Sigrid"
		},
		{
		  "op": "ArrayRemove",
		  "path": "/person/first_name",
		  "value": "John"    
		},
		{
		  "op": "ObjectAdd",
		  "path": "/person/new",
		  "value": "new"
		}                                       
	  ]`
	w := doPatchRequest(t, docURL, token, requestBody, owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 200, "patch_200.json")
}

// testing incorrect request on a patch
func TestDocPatch404(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/db/doc"
	requestBody :=
		`{
		"person": {
			"first_name": ["John", "greg", "fin"],
			"last_name": ["Doe", "kennedy"],
			"age": 30,
			"email": "john@example.com",
			"address": "123"
		}
	}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	docURL = "http://localhost:3318/v1/db/document"
	requestBody =
		`[
		{
		  "op": "ArrayAdd",
		  "path": "/db/doc",
		  "value": {
			"user": "another_user"
		  }
		}
	  ]`
	w := doPatchRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 404, "put_doc_404.json")
}
