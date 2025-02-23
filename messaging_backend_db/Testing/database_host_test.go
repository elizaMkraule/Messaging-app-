// this is a Testing suite that compares different https writer responses against a json file each test
// is given different handler requests and we store the response in a response writter to compare
package Testing

import (
	"sync"
	"testing"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/santhosh-tekuri/jsonschema"
)

// simple test with one database put
func TestDbGet200(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 200, "get_doc_200.json")

}

// tests the return results if the database contains no documents
func TestDbGet200Empty(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 200, "empty_col.json")

}

// throwing error if the database does not exist
func TestDbGet404(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/"
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 404, "get_db_404.json")
}

// throw an error if their is a unathorized user
func TestDbGet401(t *testing.T) {
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
	getURL := "http://localhost:3318/v1/db/"
	w := doGetRequest(t, getURL, token, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 401, "missingToken.json")
}

// testing a standard database put request
func TestDBPut201(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:3318/v1/db"

	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	// Perform PUT request and get the recorder
	w := doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 201, "put_db_201.json")

}

// testing putting a duplicate database and that an error is thrown
func TestDBPut400dupDB(t *testing.T) {
	// initialzie the owlDB database and token map
	// owlDB := database_host.Database_host{DatabaseMap: make(map[string]*database.Database)}
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:4318/v1/db"
	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	// Perform PUT request and get the recorder
	doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)
	w := doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 400, "put_db_dup.json")

}

// Testing when there is a missing token and an anothrized user is trying to do a put request
func TestDBPut401MissingToken(t *testing.T) {
	// initialzie the owlDB database and token map
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:4318/v1/db/dpc"

	// Get the Bearer Token
	token := "token"

	// Perform PUT request and get the recorder
	w := doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 401, "missingToken.json")
}
