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

// seeing if the writer response works for documents
func TestDocGet200(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL1 := "http://localhost:3318/v1/db/doc"
	requestBody := `{
		"hope": "passing",
		"this": "tests",
		"works": "check"
	  }`
	doPutDocRequest(t, docURL1, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	getURL := "http://localhost:3318/v1/db/doc"
	w := doGetRequest(t, getURL, token, &owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 200, "get_doc_20.json")

}

// test getting a document where the document does not exist
func TestDocGet404(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/do"
	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 404, "get_doc_404.json")
}

// test a bad url request for getting a document
func TestDocGet400(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/doc/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 400, "get_doc_400.json")
}

// testing putting a dcoument in a database
func TestDocPut201(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, err := compiler.Compile("document-schema.json")
	if err != nil {
		print("schema error")
	}

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
	w := doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 201, "put_doc.json")
}

// testing replacing a document
func TestDocPut200(t *testing.T) {
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
	w := doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 200, "put_doc.json")

}

// testing a put request with an unauthroized user
func TestDocPut401(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:4318/v1/db/dpc"

	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
	doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)
	token = "token"

	// Perform PUT request and get the recorder
	docURL := "http://localhost:3318/v1/db/doc"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	w := doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 401, "missingToken.json")
}
