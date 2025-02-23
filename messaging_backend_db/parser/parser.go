// Package parser, parse the path and makes sure that all of the database, documents and collections exist in the path
package parser

import (
	"encoding/hex" // new import for hexadecimals
	"fmt"
	"log/slog"
	"strings"
)

// parses the path and decodes strings when needs for slashes
func hexCheck(path string) []string {

	// Test path:
	// path := "database1/document1/collection1/doc2/coll2/"

	// Split the word based on /'
	words := strings.Split(path, "/")

	// Initialize slice to be filled with parsed words
	parsed := make([]string, len(words))

	// for each word in words
	for parsedidx, word := range words {
		// initialize empty current word
		curr := ""
		// for each index (character) in the word
		for idx := 0; idx < len(word); idx++ {
			// check if there is a hexadecimal
			if word[idx] == '%' {

				// hexval that needs to be decoded
				hexval := string(word[idx+1]) + string(word[idx+2])
				// slog.Info("this is the hexvalue")
				// slog.Info(hexval)
				// Imported decode function
				let, err := hex.DecodeString(hexval)
				// slog.Info("this is the letter decoded")
				// slog.Info(string(let))
				if err != nil {
					fmt.Print("error")
				}
				// After converting hex -> letter, add the letter to curr
				curr = curr + string(let)
			} else {
				// if it was not hex, add the letter to curr
				curr = curr + string(word[idx])
			}

		}
		// Add the parsed word to the slice
		parsed[parsedidx] = curr
	}

	// parsed should now be a slice of each word in the url
	return parsed
}

// this is called when we want to create a get request or generally need an object and want to use the entire get path
// an integer is returned called stopPoint which is the point at which the validator should stop check and is
// dependent on a put or get request
func ParseURL(url string, isPut bool) ([]string, int) {
	slog.Info("IN PARSEURLGET")
	// Split the path into segments, and check for hex values
	// this works if hexcheck is wrong: segments := strings.Split(url, "/")
	segments := hexCheck(url)

	// Extract information from the URL reading through values
	length := len(segments)
	var b = length
	if isPut {
		b = length - 1
	}
	//checking to see if there is a slash at the end
	if segments[length-1] == "" {
		slog.Info("taking away the extra slash")
		b = b - 1
	}
	return segments, b

}
