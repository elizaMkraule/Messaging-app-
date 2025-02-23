// Package jsonPatch processes a PATCH request by calling the visitor on the given document and applying the patch.
package jsonPatch

import (
	"errors"
	"log/slog"
	"reflect"

	"github.com/RICE-COMP318-FALL23/owldb-p1group07/jsonvisit"
)

// JsonPatchVisitor is contains all the necessary information to perform a PATCH operation. The fields are described below.
type JsonPatchVisitor struct {
	path     []string    // Field to store the path
	idx      int         // Field to store the idx in path
	len      int         // Field to store len of path
	value    interface{} // Field to store value to be appended
	op       string      // Field to store patch operation type
	curpath  []string    // Field to store the current path
	newkey   string      // Field to store newkey for ObjectAdd
	addval   bool        // Field to store flag for adding a value
	valadded bool        // Field to store a falg indicating if a value was added
}

// New created a new JsonPatchVisitor, on initialization the path in the json doc, value to be patched, operation type and newkey string is passed in.
func New(path []string, value interface{}, op string, newkey string) JsonPatchVisitor {
	return JsonPatchVisitor{
		path:     path,      // Path in the JSON doc
		idx:      0,         // index in the path
		len:      len(path), // length of path
		value:    value,     // value to be patched
		op:       op,        // patch operation
		newkey:   newkey,    // newkey if patch operation is ObjectAdd
		addval:   false,     // Flag to indicate that the value must be added
		valadded: false,     // Flag to to indicate that the value was added
	}
}

// Map proceses JSON Map by iterating through map and calling Accept, if path is found the patch is applied at path.
func (v JsonPatchVisitor) Map(m map[string]interface{}) (interface{}, error) {

	result := make(map[string]interface{})
	addedobj := false

	// if op is object add --> will be added to the following map where the path was found
	if v.op == "ObjectAdd" {
		if v.addval {
			slog.Info("adding value")
			result[v.newkey] = v.value
			v.addval = false
			slog.Info("added val")
			addedobj = true
		}
	}

	found_key := false

	// looking for key
	if v.idx < v.len {
		slog.Info("looking for key:", v.path[v.idx])
	} else {
		slog.Info("past the path?")
		found_key = true
	}

	for key, val := range m {
		if !found_key {
			// checking path
			slog.Info("loop key:", key)
			if v.idx < v.len {
				if key == v.path[v.idx] {
					found_key = true
					v.curpath = append(v.curpath, key)
					slog.Info("path", v.path[v.idx])
					v.idx = v.idx + 1
					if v.op == "ObjectAdd" {
						if v.idx == v.len {
							if reflect.DeepEqual(v.curpath, v.path) {
								v.addval = true
								slog.Info("path was found w object add")
							}
						}
					}

				} else {
					slog.Info("not on the path")
				}

			} else {
				slog.Info("past the path")
			}
		}

		res, err := jsonvisit.Accept(val, v)
		if v.addval {
			v.curpath = nil
			slog.Info("nill curr path as we found what we need to add")

		}
		if err != nil {
			return nil, err
		}
		result[key] = res
	}
	if found_key {
		slog.Info("found the key")
	} else if addedobj {
		slog.Info("found the key --> added obj")
	} else {

		slog.Info("didnt find key on this level should return that the path doesnt exists?")
		return nil, errors.New("given path coudlnt be found in the document")
	}
	return result, nil
}

// Slice proceses JSON slice by iterating through slice and calling Accept, if path is found the patch is applied at path.
func (v JsonPatchVisitor) Slice(s []interface{}) (interface{}, error) {
	var apval bool
	slog.Info("in slice")
	var result []interface{}

	for _, val := range v.curpath {
		slog.Info("cur val", val)
	}
	for _, vals := range v.path {
		slog.Info("path val", vals)
	}

	if reflect.DeepEqual(v.curpath, v.path) {
		apval = true
		slog.Info("path was found", apval)
	}

	for _, val := range s {
		res, err := jsonvisit.Accept(val, v)
		if err != nil {
			return nil, err
		}
		result = append(result, res)
	}
	if !v.valadded {
		if apval {
			if v.op == "ArrayAdd" {
				exists := false
				for _, existingValue := range result {
					if existingValue == v.value {
						exists = true
						break
					}
				}

				// If the value doesn't exist, add it to the array
				if !exists {
					slog.Info("adding value to array")
					result = append(result, v.value)
					v.valadded = true
				} else {
					slog.Info("value already exists in the array, skipping addition")
				}

			} else if v.op == "ArrayRemove" {

				// Remove value from the target array
				indexToRemove := -1
				for j, item := range result {
					if item == v.value {
						indexToRemove = j
						break
					}
				}
				if indexToRemove != -1 {
					result = append(result[:indexToRemove], result[indexToRemove+1:]...)
				}
				v.valadded = true
			}
		}
	}
	return result, nil
}

// Process JSON bool by returning bool
func (v JsonPatchVisitor) Bool(b bool) (interface{}, error) {
	return b, nil
}

// Process JSON float
func (v JsonPatchVisitor) Float64(f float64) (interface{}, error) {
	return f, nil
}

// Process JSON string
func (v JsonPatchVisitor) String(s string) (interface{}, error) {
	return s, nil
}

// Process JSON null value
func (v JsonPatchVisitor) Null() (interface{}, error) {
	return nil, nil
}
