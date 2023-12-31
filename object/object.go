package object

import (
	"fmt"
	"hash/fnv"
	"strings"

	"github.com/valsov/gointerpreter/ast"
)

const (
	INTEGER_OBJ      = "INTEGER"
	STRING_OBJ       = "STRING"
	BOOLEAN_OBJ      = "BOOLEAN"
	NULL_OBJ         = "NULL"
	RETURN_VALUE_OBJ = "RETURN_VALUE"
	ERROR_OBJ        = "ERROR"
	FUNCTION_OBJ     = "FUNCTION"
	BUILTIN_OBJ      = "BUILTIN"
	ARRAY_OBJ        = "ARRAY"
	HASH_OBJ         = "HASH"
)

var hashKeyCache = map[Hashable]HashKey{}

type ObjectType string

type Object interface {
	Type() ObjectType
	Inspect() string
}

type Integer struct {
	Value int64
}

func (i *Integer) Inspect() string  { return fmt.Sprintf("%d", i.Value) }
func (i *Integer) Type() ObjectType { return INTEGER_OBJ }

type String struct {
	Value string
}

func (s *String) Inspect() string  { return s.Value }
func (s *String) Type() ObjectType { return STRING_OBJ }

type Boolean struct {
	Value bool
}

func (b *Boolean) Inspect() string  { return fmt.Sprintf("%t", b.Value) }
func (b *Boolean) Type() ObjectType { return BOOLEAN_OBJ }

type Null struct{}

func (n *Null) Inspect() string  { return "null" }
func (n *Null) Type() ObjectType { return NULL_OBJ }

type ReturnValue struct {
	Value Object
}

func (rv *ReturnValue) Inspect() string  { return rv.Value.Inspect() }
func (rv *ReturnValue) Type() ObjectType { return RETURN_VALUE_OBJ }

type Error struct {
	Message string
}

func (e *Error) Inspect() string  { return fmt.Sprintf("ERROR: %s", e.Message) }
func (e *Error) Type() ObjectType { return ERROR_OBJ }

type Function struct {
	Parameters []*ast.Identifier
	Body       *ast.BlockStatement
	Env        *Environment
}

func (f *Function) Inspect() string {
	parameters := []string{}
	for _, p := range f.Parameters {
		parameters = append(parameters, p.String())
	}
	return fmt.Sprintf("fn(%s) {\n%s\n}", strings.Join(parameters, ", "), f.Body.String())
}
func (f *Function) Type() ObjectType { return FUNCTION_OBJ }

type BuiltinFunction func(args ...Object) Object

type Builtin struct {
	Fn BuiltinFunction
}

func (bi *Builtin) Inspect() string  { return "built-in function" }
func (bi *Builtin) Type() ObjectType { return BUILTIN_OBJ }

type Array struct {
	Elements []Object
}

func (a *Array) Inspect() string {
	elements := []string{}
	for _, elem := range a.Elements {
		elements = append(elements, elem.Inspect())
	}
	return fmt.Sprintf("[%s]", strings.Join(elements, ", "))
}
func (a *Array) Type() ObjectType { return ARRAY_OBJ }

type HashPair struct {
	Key, Value Object
}

type Hash struct {
	Pairs map[HashKey]HashPair
}

func (h *Hash) Inspect() string {
	pairs := make([]string, len(h.Pairs))
	i := 0
	for _, pair := range h.Pairs {
		pairs[i] = fmt.Sprintf("%s: %s", pair.Key.Inspect(), pair.Value.Inspect())
		i++
	}
	return fmt.Sprintf("{%s}", strings.Join(pairs, ", "))
}
func (h *Hash) Type() ObjectType { return HASH_OBJ }

type HashKey struct {
	Type  ObjectType
	Value uint64
}

type Hashable interface {
	HashKey() HashKey
}

func (b *Boolean) HashKey() HashKey {
	if hashKey, found := hashKeyCache[b]; found {
		return hashKey
	}

	var value uint64
	if b.Value {
		value = 1
	} else {
		value = 0
	}

	hashKey := HashKey{Type: b.Type(), Value: value}
	hashKeyCache[b] = hashKey
	return hashKey
}

func (i *Integer) HashKey() HashKey {
	if hashKey, found := hashKeyCache[i]; found {
		return hashKey
	}

	hashKey := HashKey{Type: i.Type(), Value: uint64(i.Value)}
	hashKeyCache[i] = hashKey
	return hashKey
}

func (s *String) HashKey() HashKey {
	if hashKey, found := hashKeyCache[s]; found {
		return hashKey
	}
	hash := fnv.New64a()
	hash.Write([]byte(s.Value))

	hashKey := HashKey{Type: s.Type(), Value: hash.Sum64()}
	hashKeyCache[s] = hashKey
	return hashKey
}
