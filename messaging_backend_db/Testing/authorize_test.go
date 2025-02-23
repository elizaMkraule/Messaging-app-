// this is a Testing suite that compares different https writer responses against a json file each test
// is given different handler requests and we store the response in a response writter to compare
package Testing

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"sync"
	"testing"

	// Needs to be commented because it disappears when program saves

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/database_host"
	"github.com/RICE-COMP318-FALL23/owldb-p1group07/handler"
	"github.com/santhosh-tekuri/jsonschema"
	// No longer importing docAndColl
)

// these specific tests focus on POST http requests and authorization of users
// tests a sucessful authorization of a user and bearer token
func TestAuthorizeSuccess(t *testing.T) {

	// initialzie the owlDB database and token map
	owlDB := database_host.Database_host{DatabaseMap: make(map[string]*database.Database)}
	tokenMap := new(sync.Map)
	schema := new(jsonschema.Schema)

	url := "http://localhost:3318/auth"
	username := []byte(`{
		"username": "a_user"
	}`)

	// Create a new HTTP request with method POST.

	req := httptest.NewRequest("POST", url, bytes.NewBuffer(username))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.HndlRequest(w, r, &owlDB, tokenMap, schema)
		})
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	expectedJSON := map[string]interface{}{"token": "a88dBdX3z9kl3Q"}

	expectedJSONString, err := json.MarshalIndent(expectedJSON, "", "  ")
	if err != nil {
		t.Fatalf("Failed to marshal expected JSON: %v", err)
	}
	fmt.Println(string(expectedJSONString))
	if !tokenEqual(string(body), string(expectedJSONString), "token") {
		t.Errorf("output incorrect: format correct then incorrect")
		fmt.Println(string(expectedJSONString))
		fmt.Println(string(body))

	} else if resp.StatusCode != 200 {
		t.Errorf("status code incorrect: format correct then incorrect")
		fmt.Println("200")
		fmt.Println(resp.StatusCode)
	}
}

// If the authorization body is not given, throw an error code 400
func TestAuthorizeBearMispell(t *testing.T) {

	// initialzie the owlDB database and token map
	owlDB := database_host.Database_host{DatabaseMap: make(map[string]*database.Database)}
	tokenMap := new(sync.Map)
	schema := new(jsonschema.Schema)

	url := "http://localhost:3318/auth"
	username := []byte(`{
		"userna": "a_user"
	}`)

	// Create a new HTTP request with method POST.

	req := httptest.NewRequest("POST", url, bytes.NewBuffer(username))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.HndlRequest(w, r, &owlDB, tokenMap, schema)
		})
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println(string(body))
	var result = `"No username in request body"` + "\n"

	if string(body) != result {
		t.Errorf("output incorrect: format correct then incorrect")
		fmt.Println(result)
		fmt.Println(string(body))

	} else if resp.StatusCode != 400 {
		t.Errorf("status code incorrect: format correct then incorrect")
		fmt.Println("400")
		fmt.Println(resp.StatusCode)

	}
}

// throw error code 400 if no athorization exists
func TestAuthorizeEmpty(t *testing.T) {

	// initialzie the owlDB database and token map
	owlDB := database_host.Database_host{DatabaseMap: make(map[string]*database.Database)}
	tokenMap := new(sync.Map)
	schema := new(jsonschema.Schema)

	url := "http://localhost:3318/auth"
	username := []byte(`{
		"username": 
	}`)

	// Create a new HTTP request with method POST.

	req := httptest.NewRequest("POST", url, bytes.NewBuffer(username))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	handler := http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			handler.HndlRequest(w, r, &owlDB, tokenMap, schema)
		})
	handler(w, req)

	resp := w.Result()
	body, _ := io.ReadAll(resp.Body)

	fmt.Println(string(body))
	var result = `"No username in request body"` + "\n"

	if string(body) != result {
		t.Errorf("output incorrect: format correct then incorrect")
		fmt.Println(result)
		fmt.Println(string(body))

	} else if resp.StatusCode != 400 {
		t.Errorf("status code incorrect: format correct then incorrect")
		fmt.Println("400")
		fmt.Println(resp.StatusCode)
	}
}

// compares to see if tokens are equal in the same output formate
func tokenEqual(json1, json2 string, ignoreKey string) bool {
	var m1 map[string]interface{}
	var m2 map[string]interface{}
	if err := json.Unmarshal([]byte(json1), &m1); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(json2), &m2); err != nil {
		return false
	}

	// Remove specified keys from the maps.

	delete(m1, ignoreKey)
	delete(m2, ignoreKey)

	return reflect.DeepEqual(m1, m2)
}

// test a sucessful deletion of a user
func TestAuthDelete204(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/auth"

	// Execute the GET request and receive the response.
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	if w.Code != 204 {
		t.Errorf("Expected status code %d, got %d", 204, w.Code)
	}

}

// throws an error if the user cannot be authroized
func TestAuthDelete404(t *testing.T) {
	token, owlDB, tokenMap, subscribers, schema := setupForGet(t)

	// Define the URL to test the GET request.
	getURL := "http://localhost:4318/auth"

	// Execute the GET request and receive the response.
	doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)
	w := doDeleteRequest(t, getURL, token, owlDB, tokenMap, subscribers, schema)

	CheckResponseGet(t, w, 401, "missingToken.json")

}
