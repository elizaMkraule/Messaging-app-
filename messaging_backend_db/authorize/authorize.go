// Package authorize encapsulates all Authorization and Authentification porcesses.
package authorize

import (
	"encoding/json"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"sync"
	"time"
)

// A Username contains string Username, Username is used for reading in the authentification request body containing the username.
type Username struct {
	Username string `json:"username"`
}

// Token holds the information necessary for a token
type Token struct {
	TokenID    string    // TokenID: the token
	Username   string    // Username: the username associated with the toke
	Expiration time.Time // Expiration: time until the token is valid to
}

// New creates a new bearer token and maps it to its username and expiration, returns the bearer token string.
func New(username string, tokenmap *sync.Map) string {
	tokenValue := generateRandomString(14)
	token := Token{TokenID: tokenValue, Username: username, Expiration: time.Now().Add(time.Hour)}
	tokenmap.Store(token.TokenID, token)
	return token.TokenID
}

// Authorize authorizes all incoming requests by validating the bearer token.
func Authorize(w http.ResponseWriter, r *http.Request, tokenmap *sync.Map) (bool, string) {

	switch r.Method {
	// skip OPTIONS AND POST /auth
	case http.MethodOptions:
		// Retrun true as Options does not need to be authorized
		return true, ""
	case http.MethodPost:
		if r.URL.Path == "/auth" {
			// retrun true as this will be handled in the post handler
			return true, ""
		}
		return defaultAuth(w, r, tokenmap)
	default:
		return defaultAuth(w, r, tokenmap)
	}
}

// Authenticate handles all authentication requests by authenitificating the beaerer token.
func Authenticate(w http.ResponseWriter, r *http.Request, document []byte, tokenmap *sync.Map) {
	// initalize struct for storing the username
	var username Username

	var data map[string]string
	// Umarhsal the document to get the username
	errs := json.Unmarshal(document, &data)
	if errs != nil {
		slog.Error("Authentification: error unmarshaling username", "error", errs)
		http.Error(w, `"No username in request body"`, http.StatusBadRequest)
		return
	}
	temp, exists := data["username"]
	username.Username = temp
	if !exists || username.Username == "" {
		slog.Error("Authentification: error parsing body request username")
		http.Error(w, `"No username in request body"`, http.StatusBadRequest)
		return
	}

	// generate a new access token and store it in the tokenmap
	accessToken := New(username.Username, tokenmap)

	//Generate a response with the access token.
	response := map[string]string{"token": accessToken}
	jsonResponse, errors := json.MarshalIndent(response, "", "  ")
	if errors != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		slog.Error("Unable to Marshal token")
		return
	}

	//Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(jsonResponse)
}

// Delete deletes an authentificated token on a DELETE request.
func Delete(w http.ResponseWriter, r *http.Request, tokenmap *sync.Map) {
	// Get the token from header.
	token := r.Header.Get("Authorization")
	slog.Info(token)
	if token == "" || len(token) < 7 || token[:7] != "Bearer " {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"Missing or invalid bearer token"`))
		return
	}

	// Extract the token value.
	tokenValue := token[7:]

	// Check if the token exists and invalidate it.
	_, ok := tokenmap.Load(tokenValue)
	if ok {
		// Token found, invalidate it.
		tokenmap.Delete(tokenValue)
		w.WriteHeader(http.StatusNoContent) // TODO: DO NOT KNOW IF THIS IS THE RIGHT CODE
		w.Write([]byte("Logged Out"))
	} else {
		// Token not found, return an error response
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"Missing or invalid bearer token"`))
	}
}

// defaultAuth is a helper function for Authorize for authorizing all incoming requests by validating the bearer token.
func defaultAuth(w http.ResponseWriter, r *http.Request, tokenmap *sync.Map) (bool, string) {
	// Extract the bearer token from header
	bearer_token := r.Header.Get("Authorization")

	// Validate the header
	if bearer_token == "" || len(bearer_token) < 7 || bearer_token[:7] != "Bearer " {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"Missing or invalid bearer token"`))
		slog.Error("Authorize: Invalid or missing Bearer token")
		return false, ""
	}

	// Extract the token value
	bearer := bearer_token[7:]

	// Check if its authentificated by making sure the the user to token maping exists
	tokenValue, ok := tokenmap.Load(bearer)
	if ok {
		tokenStruct, ok := tokenValue.(Token) // validate the type
		if ok {
			// Token found, check expiration
			if !tokenStruct.Expiration.Before(time.Now()) {
				return true, tokenStruct.Username
			} else {
				http.Error(w, "Token is expired", http.StatusUnauthorized)
				return false, ""
			}
		}
	} else {
		// Token not found, return an error response
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`"Missing or invalid bearer token"`))
		return false, ""
	}
	w.WriteHeader(http.StatusUnauthorized)
	w.Write([]byte(`"Missing or invalid bearer token"`))
	return false, ""
}

// Initialize initializes the tokenMap from the given token file for the use of authentification.
func Initialize(tokenFile string, tokenMap *sync.Map) {

	// Open the tokenFile.
	jsonFile, err := os.Open(tokenFile)
	// If returns an error then handle it.
	if err != nil {
		slog.Error("Can't open token file", "error", err)
	}

	defer jsonFile.Close()
	// Read the JSON file.
	byteValue, err := io.ReadAll(jsonFile)
	if err != nil {
		slog.Error("Error reading JSON file", "error", err)
	}
	// Unmarshal it.
	var data map[string]string
	if err := json.Unmarshal(byteValue, &data); err != nil {
		slog.Error("Error decoding JSON", "error", err)
	}
	// Make sure the data is stored with expiration date in structs.
	for username, tokenID := range data {
		token := Token{
			TokenID:    tokenID,
			Username:   username,
			Expiration: time.Now().Add(time.Hour * 24), // Example expiration time set as 1 day from now as specified in the proj description
		}
		tokenMap.Store(tokenID, token)

	}
}

// generateRandomString Generates random string for token.
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789_"
	seed := rand.NewSource(time.Now().UnixNano()) //  TODO: should i use a random integer as a seed instead of the current time
	random := rand.New(seed)
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[random.Intn(len(charset))]
	}
	return string(result)
}
