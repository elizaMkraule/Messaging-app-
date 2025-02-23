// this is a Testing suite that compares different https writer responses against a json file each test
// is given different handler requests and we store the response in a response writter to compare
package Testing

import (
	"encoding/json"
	"io"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/santhosh-tekuri/jsonschema"
)

// these tests focus on get requests. This is the main point for end to end testing
func CheckResponseGet(t *testing.T, resp *httptest.ResponseRecorder, expectedStatusCode int, expectedOutputFile string) {
	// Ensure we don't carry on after an assertion failure
	t.Helper()

	// Check status code
	if resp.Code != expectedStatusCode {
		t.Errorf("Expected status code %d, got %d", expectedStatusCode, resp.Code)
	}

	// Read expected output from the file
	expectedOutput, err := os.ReadFile(expectedOutputFile)
	if err != nil {
		t.Fatalf("Failed to read expected output file: %v", err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("Failed to read response body: %v", err)
	}

	// Try to unmarshal into maps first
	var expectedMap, resultMap []map[string]interface{}
	errExpected := json.Unmarshal(expectedOutput, &expectedMap)
	errResult := json.Unmarshal(body, &resultMap)
	if errExpected != nil || errResult != nil {
		var expectedMap, resultMap map[string]interface{}
		errExpected = json.Unmarshal(expectedOutput, &expectedMap)
		errResult = json.Unmarshal(body, &resultMap)
	}
	// Case 1: Both are JSON objects
	if errExpected == nil && errResult == nil {
		for i, refObj := range resultMap {
			testObj := expectedMap[i]
			if !AlmostDeepEqual(refObj, testObj) {
				e := string(expectedOutput)
				r := string(body)
				t.Errorf("Case 1: Response body does not match the expected output:\nExpected: %v\nGot: %v", e, r)
			}
		}
		return
	}

	expectedString := ""
	resultString := ""
	if errExpected != nil {
		expectedString = string(expectedOutput)
	}
	if errResult != nil {
		resultString = string(body)
	}
	// Convert both to strings for further comparison

	// Case 2: Both are strings
	if errExpected != nil && errResult != nil {
		if expectedString != resultString {
			t.Errorf("Case 2: Response body does not match the expected output as strings:\nExpected: %v\nGot: %v", expectedString, resultString)
		}
		return
	}

	// Case 3: One is a JSON object and the other is a string
	if errExpected == nil && errResult != nil {
		if expectedString != resultString {
			t.Errorf("Case 3: Response body does not match the expected output:\nExpected: %v\nGot: %v", expectedMap, resultString)
		}
		return
	}
	if errExpected != nil && errResult == nil {
		if expectedString != resultString {
			t.Errorf("Case 4: Response body does not match the expected output:\nExpected: %v\nGot: %v", expectedString, resultMap)
		}
		return
	}

}

// this is used to compare get requests as the meta data coantined will have different numbers so those keys are checked but values ignored
func AlmostDeepEqual(a, b interface{}) bool {
	ignoreKeys := []string{"createdAt", "lastModifiedAt"}

	// Check type info and compare basic types
	if reflect.TypeOf(a) != reflect.TypeOf(b) || !reflect.DeepEqual(a, b) {
		if aMap, aOk := a.(map[string]interface{}); aOk {
			if bMap, bOk := b.(map[string]interface{}); bOk {
				// Both a and b are maps, need to check keys and values
				for k, vA := range aMap {
					if ContainsString(ignoreKeys, k) {
						continue
					}
					vB, exists := bMap[k]
					if !exists || !AlmostDeepEqual(vA, vB) {
						return false
					}
				}
				// All keys in a exist in b and their values are equal
				return true
			}
		}
		// Types don’t match, or DeepEqual failed for non-map type
		return false
	}
	// Types match and DeepEqual passed
	return true
}
func ContainsString(slice []string, value string) bool {
	for _, item := range slice {
		if item == value {
			return true
		}
	}
	return false
}

// tests return results of getting a database when there are 2 documents in a database
func TestDbGet200doc2(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL1 := "http://localhost:3318/v1/db/doc"
	docURL2 := "http://localhost:3318/v1/db/d"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	doPutDocRequest(t, docURL1, token, requestBody, &owlDB, tokenMap, subscribers, schema)
	doPutDocRequest(t, docURL2, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	getURL := "http://localhost:3318/v1/db/"
	w := doGetRequest(t, getURL, token, &owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 200, "get_2doc.json")

}

// checks that an error is thrown when in the url path and an error should be thrown
func TestDbGet400extraslash(t *testing.T) {

	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db%2F/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 400, "get_doc_400.json")

}

// check getting a document from a collection and not a database since they are different functions
func TestDocGet200Nested(t *testing.T) {
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
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)
	w := doGetRequest(t, docURL, token, &owlDB, tokenMap, subscribers, schema)
	CheckResponseGet(t, w, 200, "get_doc_nested.json")
}

// throwing an error if there are too many slashes
func TestDocGet400Double(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:3318/v1/db/doc//"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 400, "get_doc_400.json")
}

// getting collection from a bad url request
func TestColGet404doc(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/do%2Fcol/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 404, "get_doc_404.json")

}

// getting a collection where the collection does not exist
func TestColGet404col(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/doc%2Fco/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 404, "get_col_404.json")

}

// getting collection but the url path is bad
func TestColGet400slash(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/doc%2Fcol%2F/"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 400, "get_doc_400.json")

}

// sucessfuly deletes a collection
func TestColDelete204(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/doc%2Fcol/"

	// Execute the GET request and receive the response.
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	if w.Code != 204 {
		t.Errorf("Expected status code %d, got %d", 204, w.Code)
	}

}

// testing interval query with multiple documents to parse
func TestDbGet200IntervalComplex(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)
	docURL := "http://localhost:3318/v1/db/aa"
	requestBody :=
		`{
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	}`
	doPutDocRequest(t, docURL, token, requestBody, owlDB, tokenMap, subscribers, schema)
	docURL = "http://localhost:3318/v1/db/ab"
	doPutDocRequest(t, docURL, token, requestBody, owlDB, tokenMap, subscribers, schema)
	docURL = "http://localhost:3318/v1/db/bb"
	doPutDocRequest(t, docURL, token, requestBody, owlDB, tokenMap, subscribers, schema)
	docURL = "http://localhost:3318/v1/db/a"
	doPutDocRequest(t, docURL, token, requestBody, owlDB, tokenMap, subscribers, schema)
	docURL = "http://localhost:3318/v1/db/bc"
	doPutDocRequest(t, docURL, token, requestBody, owlDB, tokenMap, subscribers, schema)
	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/?interval=%5Baa%2Cbb%5D"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 200, "interval_complex.json")

}

// testing putting mutiple databases
func TestDBPut201twoDB(t *testing.T) {
	// initialzie the owlDB database and token map
	// owlDB := database_host.Database_host{DatabaseMap: make(map[string]*database.Database)}
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL1 := "http://localhost:4318/v1/database"
	URL2 := "http://localhost:4318/v1/db"
	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	// Perform PUT request and get the recorder
	doPutRequest(t, URL1, &owlDB, tokenMap, subscribers, schema, token)
	w := doPutRequest(t, URL2, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 201, "put_db_201.json")

}

// testing if an interval query is called
func TestDbGet200Interval(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/?interval=%5Baa%2Cbb%5D"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 200, "get_doc_200.json")

}

// testing post of a document in a database
func TestDbPost201(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/db/"
	requestBody :=
		`{
		"additionalProp1": "string",
		"additionalProp2": "string",
		"additionalProp3": "string"
	  }`
	w := doPostRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 201, "patch_doc_200.json")
}
