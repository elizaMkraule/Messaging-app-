// jsonvisit is a package that implements the visitor design principle. There is an interface to call accept on different types of objects
package jsonvisit

import "fmt"

// A Visitor is used by Accept to process arbitrary values that only contain
// data of valid JSON types.  The Map and Slice methods may recursively call
// Accept on their consituent elements.  They may also use Equal to compare
// their constituent elements with other JSON values.
type Visitor[T any] interface {
	Map(map[string]any) (T, error)
	Slice([]any) (T, error)
	Bool(bool) (T, error)
	Float64(float64) (T, error)
	String(string) (T, error)
	Null() (T, error)
}

// Accept applies the given input visitor to the given input value by calling
// the appropriate visitor method given the type of the input value. Returns an
// error if value is of a type that is not a valid JSON type or if the visitor
// method returns an error.
func Accept[T any](value any, visitor Visitor[T]) (T, error) {
	switch val := value.(type) {
	case map[string]any:
		return visitor.Map(val)
	case []any:
		return visitor.Slice(val)
	case float64:
		return visitor.Float64(val)
	case bool:
		return visitor.Bool(val)
	case string:
		return visitor.String(val)
	case nil:
		return visitor.Null()
	default:
		var zero T
		return zero, fmt.Errorf("invalid JSON value: %v", value)
	}
}
