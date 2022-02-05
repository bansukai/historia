package historia

import "reflect"

// TypeOf is a convenience function that returns the name of a type
func TypeOf(i interface{}) string {
	return reflect.TypeOf(i).Elem().Name()
}

// PathOf is a convenience function that returns the package path of a type
func PathOf(i interface{}) string {
	return reflect.TypeOf(i).Elem().PkgPath()
}
