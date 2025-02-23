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

// putting a collection with a database that does not exist
func TestColPut404dbDNE(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/d/doc"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	colURL := "http://localhost:4318/v1/d/doc%2Fcol/"
	w := doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 404, "put_col_404db.json")

}

// putting a colleciton where the document does not exist
func TestColPut404docDNE(t *testing.T) {
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

	colURL := "http://localhost:4318/v1/db/dc%2Fcol/"
	w := doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 404, "put_col_404doc.json")

}

// testing a document post in a collection
func TestColPost200(t *testing.T) {
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

	w := doPostRequest(t, colURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)
	CheckResponse(t, w, 200, "get_col_200.json")

}

// test a collection post when there is already a document in the collection
func TestColPost2002(t *testing.T) {
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

	w := doPostRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)
	CheckResponse(t, w, 400, "get_col_200.json")

}
