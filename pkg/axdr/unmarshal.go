package axdr

import (
	"fmt"
	"reflect"
)

func UnmarshalData(data DlmsData, v interface{}) error {
	valuePtr := reflect.ValueOf(v)
	if valuePtr.Kind() != reflect.Ptr || valuePtr.IsNil() {
		return fmt.Errorf("v must be a non-nil pointer")
	}

	return unify(&data, reflect.Indirect(valuePtr))
}

func unify(data *DlmsData, rv reflect.Value) error {
	expectedKind := rv.Kind()
	gotKind := reflect.ValueOf(data.Value).Kind()

	if expectedKind == reflect.Struct && gotKind == reflect.Slice {
		gotKind = reflect.Struct
	}

	if expectedKind != gotKind {
		return fmt.Errorf("expected %s, got %s", expectedKind, gotKind)
	}

	if expectedKind == reflect.Struct || expectedKind == reflect.Slice {
		if expectedKind == reflect.Slice {
			return unifySlice(data, rv)
		} else {
			return unifyStruct(data, rv)
		}
	} else {
		rv.Set(reflect.ValueOf(data.Value))
		return nil
	}
}

func unifySlice(data *DlmsData, rv reflect.Value) error {
	slice := data.Value.([]*DlmsData)

	n := len(slice)
	if rv.IsNil() || rv.Cap() < n {
		rv.Set(reflect.MakeSlice(rv.Type(), n, n))
	}
	rv.SetLen(n)

	for i := 0; i < n; i++ {
		sliceval := reflect.Indirect(rv.Index(i))
		if err := unify(slice[i], sliceval); err != nil {
			return fmt.Errorf("slice error in field %d: %w", i, err)
		}
	}

	return nil
}

func unifyStruct(data *DlmsData, rv reflect.Value) error {
	slice := data.Value.([]*DlmsData)
	n := len(slice)

	if rv.NumField() != n {
		return fmt.Errorf("struct has %d fields, but data has %d fields", rv.NumField(), n)
	}

	for i := 0; i < n; i++ {
		sliceval := reflect.Indirect(rv.Field(i))
		if err := unify(slice[i], sliceval); err != nil {
			return fmt.Errorf("struct error in field %s: %w", rv.Type().Field(i).Name, err)
		}
	}

	return nil
}
