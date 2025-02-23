// this is a Testing suite that compares different https writer responses against a json file each test
// is given different handler requests and we store the response in a response writter to compare
package Testing

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"sync"
	"testing"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/handler"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/skiplist"
	"github.com/santhosh-tekuri/jsonschema"
)

// these tests focus on put requests
type Token struct {
	Token string `json:"token"`
}

// this is a helper function for Check Response
func CheckResponse(t *testing.T, resp *httptest.ResponseRecorder, expectedStatusCode int, expectedOutputFile string) {
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
	var expectedMap, resultMap map[string]interface{}
	errExpected := json.Unmarshal(expectedOutput, &expectedMap)
	errResult := json.Unmarshal(body, &resultMap)

	// Case 1: Both are JSON objects
	if errExpected == nil && errResult == nil {
		if !reflect.DeepEqual(resultMap, expectedMap) {
			t.Errorf("Case 1: Response body does not match the expected output:\nExpected: %v\nGot: %v", expectedMap, resultMap)
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

// helper function that authorizes a user to make requests
func getBearerToken(t *testing.T, owlDB *database_host.Database_host, tokenMap *sync.Map, subscribers *sync.Map, schema *jsonschema.Schema) string {
	t.Helper()
	username := []byte(`{
		"username": "a_user"
	}`)
	req := httptest.NewRequest("POST", "http://localhost:3318/auth", bytes.NewBuffer(username))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.HndlRequest(w, r, owlDB, tokenMap, schema)
		})
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	var response Token
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("Failed to unmarshal token response: %v", err)
	}

	return response.Token
}

// helper function that does put requests for databases and collections
func doPutRequest(t *testing.T, url string, owlDB *database_host.Database_host, tokenMap *sync.Map, subscribers *sync.Map, schema *jsonschema.Schema, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest("PUT", url, nil)
	req.Header.Set("accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.HndlRequest(w, r, owlDB, tokenMap, schema)
		})
	handler(w, req)

	return w
}

func doPutDocRequest(t *testing.T, url, token, requestBody string, owlDB *database_host.Database_host, tokenMap *sync.Map, subscribers *sync.Map, schema *jsonschema.Schema) *httptest.ResponseRecorder {
	t.Helper() // Marks the calling function as a test helper function.
	req, err := http.NewRequest("PUT", url, bytes.NewBufferString(requestBody))
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

// does put requests for documents
func doGetRequest(t *testing.T, url, token string, owlDB *database_host.Database_host, tokenMap *sync.Map, subscribers *sync.Map, schema *jsonschema.Schema) *httptest.ResponseRecorder {
	t.Helper()

	// Create a new HTTP GET request.
	req := httptest.NewRequest("GET", url, nil)
	req.Header.Set("accept", "application/json")
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

// helper function that takes in the request body for a get request
func setupForGet(t *testing.T) (string, *database_host.Database_host, *sync.Map, *sync.Map, *jsonschema.Schema) {
	t.Helper()

	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)

	// Create a JSON Schema compiler
	compiler := jsonschema.NewCompiler()

	// Compile JSON schema
	schema, err := compiler.Compile("document-schema.json")
	if err != nil {
		slog.Error("schema compilation error", "error", err)
	}

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/db/doc"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	colURL := "http://localhost:4318/v1/db/doc%2Fcol/"
	doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token) // Assuming doPutRequest has a requestBody parameter

	return token, &owlDB, tokenMap, subscribers, schema
}

// testing put error with incorrect url
func TestDBPut400urlIncorrect(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:4318/v1/db/"

	// Get the Bearer Token
	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	// Perform PUT request and get the recorder
	w := doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 400, "put_db_400.json")

}

// testing for an invalid token
func TestDBPut401InvalidToken(t *testing.T) {
	// initialzie the owlDB database and token map
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")
	URL := "http://localhost:4318/v1/db/dpc"

	// Get the Bearer Token
	getBearerToken(t, &owlDB, tokenMap, subscribers, schema)
	token := "token"

	// Perform PUT request and get the recorder
	w := doPutRequest(t, URL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 401, "missingToken.json")
}

// testing put request with url that contains an extra slash
func TestDocPut400incorrectURL(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/db"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/db/doc/"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	w := doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 400, "put_doc_400.json")

}

// testing put request where database does not exist
func TestDocPut404(t *testing.T) {
	owlDB := database_host.Database_host{Name: "db_host", DBSkipList: skiplist.NewList[string, *database.Database]("", "zzz")}
	tokenMap := new(sync.Map)
	subscribers := new(sync.Map)
	compiler := jsonschema.NewCompiler()
	schema, _ := compiler.Compile("document-schema.json")

	token := getBearerToken(t, &owlDB, tokenMap, subscribers, schema)

	dbURL := "http://localhost:4318/v1/d"
	doPutRequest(t, dbURL, &owlDB, tokenMap, subscribers, schema, token)

	docURL := "http://localhost:3318/v1/db/doc/"
	requestBody := `{"additionalProp1":"string","additionalProp2":"string","additionalProp3":"string"}`
	w := doPutDocRequest(t, docURL, token, requestBody, &owlDB, tokenMap, subscribers, schema)

	CheckResponse(t, w, 404, "put_doc_404.json")

}

// testing replacin a collection
func TestColPut400DupCol(t *testing.T) {
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
	doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)
	w := doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 400, "put_col_400.json")

}

// test putting a colleciton iwth an incorrect url
func TestColPut400incorrectURL(t *testing.T) {
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

	colURL := "http://localhost:4318/v1/db/doc%2Fcol"
	w := doPutRequest(t, colURL, &owlDB, tokenMap, subscribers, schema, token)

	CheckResponse(t, w, 400, "put_db_400.json")

}

// another test for getting a collection but the url path is incorrect
func TestColGet400woslash(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/v1/db/doc%2Fcol"

	// Execute the GET request and receive the response.
	w := doGetRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 400, "get_doc_400.json")

}
