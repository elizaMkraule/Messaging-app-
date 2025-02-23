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

// testing the writer output when the collection contains no documents
func TestColGet200Empty(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/doc%2Fcol/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 200, "empty_col.json")

}

// test the writer request was sucessfully putting a collection
func TestColGet200(t *testing.T) {
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
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	colURL := "http://localhost:4318/v1/db/doc%2Fcol/"
	doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL = "http://localhost:4318/v1/db/doc%2Fcol%2Fd"
	requestBody =
		`{
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)
	docURL = "http://localhost:4318/v1/db/doc%2Fcol%2Fdoc"
	requestBody =
		`{
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	w := doGetRequest(t, colURL, token, &owlDB, tokenMap, subscribers, schema)
	CheckResponseGet(t, w, 200, "get_col_200.json")

}

// getting collection but the user is not authroized
func TestColGet401(t *testing.T) {
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
	colURL := "http://localhost:4318/v1/db/doc%2Fcol/"
	doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	token = "token"
	getURL := "http://localhost:4318/v1/db/doc%2Fcol/"
	w := doGetRequest(t, getURL, token, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 401, "missingToken.json")
}

// testing putting a collection
func TestColPut201(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/db/doc"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	colURL := "http://localhost:4318/v1/db/doc%2Fcol/"
	w := doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 201, "put_col_201.json")
}

// testing putting a collection with unauthroized user
func TestColPut401(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:4318/v1/db/dpc"

	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
	doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	// Perform PUT request and get the recorder
	docURL := "http://localhost:3318/v1/db/doc"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	token = "token"
	colURL := "http://localhost:4318/v1/db/doc%2Fcol"
	w := doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 401, "missingToken.json")

}

// testing putting a document in a collection
func TestDocPut201Nested(t *testing.T) {
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
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	colURL := "http://localhost:4318/v1/db/doc%2Fcol/"
	doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL = "http://localhost:4318/v1/db/doc%2Fcol%2Fd"
	requestBody =
		`{
		"hope": "passing",
		"this": "tests",
		"works": "check"
	  }`
	w := doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 201, "put_doc_nested.json")
}
