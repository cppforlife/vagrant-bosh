package index

import (
	"encoding/json"
	"reflect"

	bosherr "bosh/errors"
	boshsys "bosh/system"
)

type FileIndex struct {
	path string
	fs   boshsys.FileSystem
}

type indexEntry struct {
	Key   map[string]interface{}
	Value json.RawMessage
}

func NewFileIndex(path string, fs boshsys.FileSystem) FileIndex {
	return FileIndex{path: path, fs: fs}
}

func (ri FileIndex) List(entries interface{}) error {
	rawEntries, err := ri.readRawEntries()
	if err != nil {
		return err
	}

	var rawEntryValues []json.RawMessage

	for _, rawEntry := range rawEntries {
		rawEntryValues = append(rawEntryValues, rawEntry.Value)
	}

	// todo avoid serializing already collected entries
	rawEntryValuesBytes, err := json.Marshal(rawEntryValues)
	if err != nil {
		return err
	}

	err = json.Unmarshal(rawEntryValuesBytes, entries)
	if err != nil {
		return err
	}

	return nil
}

func (ri FileIndex) ListKeys(keys interface{}) error {
	rawEntries, err := ri.readRawEntries()
	if err != nil {
		return err
	}

	keysElem := reflect.ValueOf(keys).Elem()

	for _, rawEntry := range rawEntries {
		key, err := ri.mapToStructFromSlice(rawEntry.Key, keys)
		if err != nil {
			return err
		}

		keysElem.Set(reflect.Append(keysElem, key))
	}

	return nil
}

func (ri FileIndex) Find(key interface{}, entry interface{}) error {
	rawEntries, err := ri.readRawEntries()
	if err != nil {
		return err
	}

	rawKey, err := ri.structToMap(key)
	if err != nil {
		return err
	}

	for _, rawEntry := range rawEntries {
		if reflect.DeepEqual(rawEntry.Key, rawKey) {
			err := json.Unmarshal(rawEntry.Value, entry)
			if err != nil {
				return err
			}

			return nil
		}
	}

	return ErrNotFound
}

func (ri FileIndex) Save(key interface{}, entry interface{}) error {
	rawEntries, err := ri.readRawEntries()
	if err != nil {
		return err
	}

	rawKey, err := ri.structToMap(key)
	if err != nil {
		return err
	}

	rawValue, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	foundI := -1

	for i, rawEntry := range rawEntries {
		if reflect.DeepEqual(rawEntry.Key, rawKey) {
			foundI = i
			break
		}
	}

	if foundI >= 0 {
		rawEntries[foundI].Value = rawValue
	} else {
		rawEntries = append(rawEntries, indexEntry{
			Key:   rawKey,
			Value: rawValue,
		})
	}

	err = ri.writeRawEntries(rawEntries)
	if err != nil {
		return err
	}

	return nil
}

func (ri FileIndex) Remove(key interface{}) error {
	rawEntries, err := ri.readRawEntries()
	if err != nil {
		return err
	}

	rawKey, err := ri.structToMap(key)
	if err != nil {
		return err
	}

	for i, rawEntry := range rawEntries {
		if reflect.DeepEqual(rawEntry.Key, rawKey) {
			rawEntries = append(rawEntries[:i], rawEntries[i+1:]...)
			break
		}
	}

	err = ri.writeRawEntries(rawEntries)
	if err != nil {
		return err
	}

	return nil
}

func (ri FileIndex) readRawEntries() ([]indexEntry, error) {
	var entries []indexEntry

	if ri.fs.FileExists(ri.path) {
		bytes, err := ri.fs.ReadFile(ri.path)
		if err != nil {
			return entries, bosherr.WrapError(err, "Reading index file %s", ri.path)
		}

		err = json.Unmarshal(bytes, &entries)
		if err != nil {
			return entries, bosherr.WrapError(err, "Unmarshalling index entries")
		}
	}

	return entries, nil
}

func (ri FileIndex) writeRawEntries(entries []indexEntry) error {
	bytes, err := json.Marshal(entries)
	if err != nil {
		return bosherr.WrapError(err, "Marshalling index entries")
	}

	err = ri.fs.WriteFile(ri.path, bytes)
	if err != nil {
		return bosherr.WrapError(err, "Writing index file %s", ri.path)
	}

	return nil
}

// structToMap extracts fields from a struct and populates a map
func (ri FileIndex) structToMap(s interface{}) (map[string]interface{}, error) {
	res := map[string]interface{}{}
	st := reflect.TypeOf(s)
	stv := reflect.ValueOf(s)

	if stv.Kind() != reflect.Struct {
		return res, bosherr.New(
			"Must be reflect.Struct: %#v (%#v)", stv, ri.kindToStr(stv.Kind()))
	}

	for i := 0; i < st.NumField(); i++ {
		res[st.Field(i).Name] = stv.Field(i).Interface()
	}

	return res, nil
}

// mapToStruct returns new struct value with data from a map
func (ri FileIndex) mapToStruct(m map[string]interface{}, t interface{}) (reflect.Value, error) {
	return ri.mapToNewStruct(m, reflect.ValueOf(t).Elem().Type())
}

// mapToStructFromSlice returns new struct value with data from a map
func (ri FileIndex) mapToStructFromSlice(m map[string]interface{}, t interface{}) (reflect.Value, error) {
	slice := reflect.ValueOf(t).Elem()

	if slice.Kind() != reflect.Slice {
		return reflect.Value{}, bosherr.New(
			"Must be reflect.Slice: %#v (%#v)",
			slice, ri.kindToStr(slice.Kind()),
		)
	}

	return ri.mapToNewStruct(m, slice.Type().Elem())
}

// mapToNewStruct returns new struct of type t with data from a map
func (ri FileIndex) mapToNewStruct(m map[string]interface{}, t reflect.Type) (reflect.Value, error) {
	if t.Kind() != reflect.Struct {
		return reflect.Value{}, bosherr.New(
			"Must be reflect.Struct: %#v (%#v)",
			t, ri.kindToStr(t.Kind()),
		)
	}

	newStruct := reflect.New(t).Elem()

	for k, v := range m {
		f := newStruct.FieldByName(k)
		if f.IsValid() && f.CanSet() {
			// todo float64 -> int
			// todo pointer values
			// todo slices
			f.Set(reflect.ValueOf(v))
		}
	}

	return newStruct, nil
}

var kindToStrMap = map[reflect.Kind]string{
	reflect.Invalid:       "Invalid",
	reflect.Bool:          "Bool",
	reflect.Int:           "Int",
	reflect.Int8:          "Int8",
	reflect.Int16:         "Int16",
	reflect.Int32:         "Int32",
	reflect.Int64:         "Int64",
	reflect.Uint:          "Uint",
	reflect.Uint8:         "Uint8",
	reflect.Uint16:        "Uint16",
	reflect.Uint32:        "Uint32",
	reflect.Uint64:        "Uint64",
	reflect.Uintptr:       "Uintptr",
	reflect.Float32:       "Float32",
	reflect.Float64:       "Float64",
	reflect.Complex64:     "Complex64",
	reflect.Complex128:    "Complex128",
	reflect.Array:         "Array",
	reflect.Chan:          "Chan",
	reflect.Func:          "Func",
	reflect.Interface:     "Interface",
	reflect.Map:           "Map",
	reflect.Ptr:           "Ptr",
	reflect.Slice:         "Slice",
	reflect.String:        "String",
	reflect.Struct:        "Struct",
	reflect.UnsafePointer: "UnsafePointer",
}

func (ri FileIndex) kindToStr(k reflect.Kind) string {
	return kindToStrMap[k]
}
