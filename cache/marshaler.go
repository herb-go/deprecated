package cache

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"sync"
)

//Marshaler Marshaler inteface
type Marshaler interface {
	Marshal(v interface{}) ([]byte, error)
	Unmarshal(bytes []byte, v interface{}) error
}

//MarshalerFactory create marshaler.
//Reutrn marshaler created and any error if raised..
type MarshalerFactory func() (Marshaler, error)

var (
	marshalerFactorysMu sync.RWMutex
	marshalerFactories  = make(map[string]MarshalerFactory)
)

//DefaultMarshaler default marshaler name
var DefaultMarshaler = "msgpack"

// RegisterMarshaler makes a marshaler creator available by the provided name.
// If Register is called twice with the same name or if driver is nil,
// it panics.
func RegisterMarshaler(name string, f MarshalerFactory) {
	marshalerFactorysMu.Lock()
	defer marshalerFactorysMu.Unlock()
	if f == nil {
		panic(errors.New("cache: Register marshaler factory is nil"))
	}
	if _, dup := marshalerFactories[name]; dup {
		panic(errors.New("cache: Register marshaler twice for factory " + name))
	}
	marshalerFactories[name] = f
}

// UnregisterAllMarshalers Unregister all marshalers
func UnregisterAllMarshalers() {
	marshalerFactorysMu.Lock()
	defer marshalerFactorysMu.Unlock()
	// For tests.
	marshalerFactories = make(map[string]MarshalerFactory)
}

//MarshalerFactories returns a sorted list of the names of the registered marshaler factories.
func MarshalerFactories() []string {
	marshalerFactorysMu.RLock()
	defer marshalerFactorysMu.RUnlock()
	var list []string
	for name := range marshalerFactories {
		list = append(list, name)
	}
	sort.Strings(list)
	return list
}

//NewMarshaler create new marshaler with given name.
//Return marshaler created and any error if raised.
func NewMarshaler(name string) (Marshaler, error) {
	marshalerFactorysMu.RLock()
	factoryi, ok := marshalerFactories[name]
	marshalerFactorysMu.RUnlock()
	if !ok {
		errmsg := "cache: unknown marshaler %q (forgotten import?)"
		if name == "msgpack" {
			errmsg = errmsg + "\nYou should add \n\timport _ \"github.com/herb-go/deprecated/cache/marshalers/msgpackmarshaler\" \nto your code."
		}
		return nil, fmt.Errorf(errmsg, name)
	}
	return factoryi()
}

//JSONMarshaler Marshaler which marshal data as JSON format
type JSONMarshaler struct {
}

//Marshal Marshal data model to  bytes.
//Return marshaled bytes and any erro rasied.
func (m *JSONMarshaler) Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(v)
}

//Unmarshal Unmarshal bytes to data model.
//Parameter v should be pointer to empty data model which data filled in.
//Return any error raseid.
func (m *JSONMarshaler) Unmarshal(bytes []byte, v interface{}) error {
	return json.Unmarshal(bytes, v)
}

func init() {
	RegisterMarshaler("json", func() (Marshaler, error) {
		return &JSONMarshaler{}, nil
	})
}
