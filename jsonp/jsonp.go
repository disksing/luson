package jsonp

import (
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/pkg/errors"
)

type (
	// Any represents any JSON node.
	Any = interface{}
	// Object represents JSON objects.
	Object = map[string]Any
	// Array represents JSON arrays.
	Array = []Any
)

// Clone makes a copy of a JSON node.
func Clone(x Any) Any {
	if obj, ok := x.(Object); ok {
		obj2 := make(Object, len(obj))
		for k, v := range obj {
			obj2[k] = Clone(v)
		}
		return obj2
	}
	if arr, ok := x.(Array); ok {
		arr2 := make(Array, len(arr))
		for i := range arr {
			arr2[i] = Clone(arr[i])
		}
		return arr2
	}
	return x
}

// Get returns child node of a node.
func Get(x Any, pointer string) (Any, error) {
	t := newTokenizer(pointer)
	for {
		key, err := t.Next()
		if err == io.EOF {
			return x, nil
		}
		if err != nil {
			return nil, err
		}
		x, err = getChild(x, key)
		if err != nil {
			return nil, err
		}
	}
}

func getChild(x Any, key string) (Any, error) {
	if obj, ok := x.(Object); ok {
		o, ok := obj[key]
		if !ok {
			return nil, errors.Errorf("no element with key %s", key)
		}
		return o, nil
	}
	if arr, ok := x.(Array); ok {
		idx, err := parseIndex(arr, key, false)
		if err != nil {
			return nil, err
		}
		return arr[idx], nil
	}
	return nil, errors.Errorf("node is not array or object")
}

// Add adds a node to a node by pointer.
func Add(x Any, pointer string, v Any) (Any, error) {
	return addRecr(x, newTokenizer(pointer), v)
}

func addRecr(x Any, t *tokenizer, v Any) (Any, error) {
	key, err := t.Next()
	if err == io.EOF {
		return v, nil
	}
	if err != nil {
		return nil, err
	}
	if obj, ok := x.(Object); ok {
		child, err := addRecr(obj[key], t, v)
		if err != nil {
			return nil, err
		}
		obj[key] = child
		return obj, nil
	}
	if arr, ok := x.(Array); ok {
		idx, err := parseIndex(arr, key, !t.More())
		if err != nil {
			return nil, errors.Wrap(err, "path:"+t.pointer)
		}
		if !t.More() {
			if idx == len(arr) {
				return append(arr, v), nil
			}
			return append(arr[:idx], append([]Any{v}, arr[idx:]...)...), nil
		}
		o, err := addRecr(arr[idx], t, v)
		if err != nil {
			return nil, err
		}
		arr[idx] = o
		return arr, nil
	}
	return nil, errors.Errorf("node is not array or object")
}

// Remove removes child of a node.
func Remove(x Any, pointer string) (Any, error) {
	x, _, err := removeRecr(x, newTokenizer(pointer))
	return x, err
}

// returns (left, removed, error)
func removeRecr(x Any, t *tokenizer) (Any, Any, error) {
	key, err := t.Next()
	if err == io.EOF {
		return nil, x, nil
	}
	if err != nil {
		return nil, nil, err
	}
	if obj, ok := x.(Object); ok {
		o, r, err := removeRecr(obj[key], t)
		if err != nil {
			return nil, nil, err
		}
		if o == nil {
			delete(obj, key)
		} else {
			obj[key] = o
		}
		return obj, r, nil
	}
	if arr, ok := x.(Array); ok {
		idx, err := parseIndex(arr, key, false)
		if err != nil {
			return nil, nil, errors.Wrap(err, "path:"+t.pointer)
		}
		o, r, err := removeRecr(arr[idx], t)
		if err != nil {
			return nil, nil, errors.Wrap(err, "path:"+t.pointer)
		}
		if o == nil {
			arr = append(arr[:idx], arr[idx+1:]...)
		} else {
			arr[idx] = o
		}
		return arr, r, nil
	}
	return nil, nil, errors.Errorf("node is not array or object")
}

// Replace replaces child of a node with another.
func Replace(x Any, pointer string, v Any) (Any, error) {
	return replaceRecr(x, newTokenizer(pointer), v)
}

func replaceRecr(x Any, t *tokenizer, v Any) (Any, error) {
	key, err := t.Next()
	if err == io.EOF {
		return v, nil
	}
	if err != nil {
		return nil, err
	}
	if obj, ok := x.(Object); ok {
		o, err := replaceRecr(obj[key], t, v)
		if err != nil {
			return nil, err
		}
		obj[key] = o
		return obj, nil
	}
	if arr, ok := x.(Array); ok {
		idx, err := parseIndex(arr, key, false)
		if err != nil {
			return nil, errors.Wrap(err, "path:"+t.pointer)
		}
		o, err := replaceRecr(arr[idx], t, v)
		if err != nil {
			return nil, errors.Wrap(err, "path:"+t.pointer)
		}
		arr[idx] = o
		return arr, nil
	}
	return nil, errors.Errorf("node is not array or object")
}

// Move moves a child to another place.
func Move(x Any, from, to string) (Any, error) {
	x2, r, err := removeRecr(x, newTokenizer(from))
	if err != nil {
		return nil, err
	}
	return Add(x2, to, r)
}

// Move2 moves a child to another place.
// The source and target can be different.
func Move2(x, y Any, from, to string) (Any, Any, error) {
	x2, r, err := removeRecr(x, newTokenizer(from))
	if err != nil {
		return nil, nil, err
	}
	y2, err := Add(y, to, r)
	if err != nil {
		return nil, nil, err
	}
	return x2, y2, nil
}

// Copy copies a child node to another place.
func Copy(x Any, from, to string) (Any, error) {
	v, err := Get(x, from)
	if err != nil {
		return nil, err
	}
	return Add(x, to, Clone(v))
}

// Copy2 copies a child node to another place.
// The source and target can be different.
func Copy2(x, y Any, from, to string) (Any, error) {
	v, err := Get(x, from)
	if err != nil {
		return nil, err
	}
	return Add(y, to, Clone(v))
}

// Test checks if child matches the node.
func Test(x Any, pointer string, v Any) error {
	x, err := Get(x, pointer)
	if err != nil {
		return err
	}
	if !reflect.DeepEqual(x, v) {
		return errors.Errorf("test fail, value of %s is %v, expect value is %v", pointer, x, v)
	}
	return nil
}

// Merge merges two nodes.
func Merge(x, v Any) Any {
	if vo, ok := v.(Object); ok {
		xo, ok := x.(Object)
		if !ok {
			xo = Object{}
		}
		for k, v := range vo {
			if v == nil {
				delete(xo, k)
			} else {
				xo[k] = Merge(xo[k], v)
			}
		}
		return xo
	}
	return v
}

type tokenizer struct {
	pointer string
	i       int
}

func newTokenizer(pointer string) *tokenizer {
	return &tokenizer{
		pointer: pointer,
	}
}

func (t *tokenizer) Next() (string, error) {
	if t.i == len(t.pointer) {
		return "", io.EOF
	}
	if t.pointer[t.i] != '/' {
		return "", errors.Errorf("invalid path, expected '/', pos=%d", t.i)
	}
	t.i++
	var sb strings.Builder
	for ; t.i < len(t.pointer); t.i++ {
		if t.pointer[t.i] == '/' {
			return sb.String(), nil
		} else if t.pointer[t.i] == '~' {
			t.i++
			if t.i < len(t.pointer) && t.pointer[t.i] == '0' {
				sb.WriteByte('~')
			} else if t.i < len(t.pointer) && t.pointer[t.i] == '1' {
				sb.WriteByte('/')
			} else {
				return "", errors.Errorf(`invalid path %s ('~'should encode to '~0')`, t.pointer)
			}
		} else {
			sb.WriteByte(t.pointer[t.i])
		}
	}
	return sb.String(), nil
}

func (t *tokenizer) More() bool {
	return t.i < len(t.pointer)
}

func parseIndex(arr Array, str string, allowAppend bool) (idx int, err error) {
	if str == "-" {
		idx = len(arr)
	} else {
		idx, err = strconv.Atoi(str)
		if err != nil {
			return 0, errors.Errorf("bad format index '%v'", str)
		}
	}
	if idx < 0 || idx > len(arr) || (idx == len(arr) && !allowAppend) {
		return 0, errors.Errorf("index '%v' out of range '%v'", idx, len(arr))
	}
	return idx, nil
}
