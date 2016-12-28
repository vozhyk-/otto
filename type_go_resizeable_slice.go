package otto

import (
	"reflect"
	"strconv"
)

func (runtime *_runtime) newGoResizeableSliceObject(value reflect.Value) *_object {
	self := runtime.newObject()
	self.class = "GoArray" // TODO GoSlice?
	self.objectClass = _classGoResizeableSlice
	self.value = _newGoResizeableSliceObject(value)
	return self
}

type _goResizeableSliceObject struct {
	value reflect.Value
}

func _newGoResizeableSliceObject(value reflect.Value) *_goResizeableSliceObject {
	self := &_goResizeableSliceObject{
		value: value,
	}
	return self
}

func (self _goResizeableSliceObject) getLength() int {
	return reflect.Indirect(self.value).Len()
}

func (self _goResizeableSliceObject) getValue(index int64) (reflect.Value, bool) {
	value := reflect.Indirect(self.value)
	if index < int64(self.getLength()) {
		return value.Index(int(index)), true
	}
	return reflect.Value{}, false
}

func (self _goResizeableSliceObject) setValue(index int64, value Value) bool {
	reflectValue, err := value.toReflectValue(reflect.Indirect(self.value).Type().Elem())
	if err != nil {
		panic(err)
	}

	if index == int64(self.getLength()) {
		reflect.Indirect(self.value).Set(
			reflect.Append(reflect.Indirect(self.value), reflectValue))
		return true
	}

	indexValue, exists := self.getValue(index)
	if !exists {
		return false
	}
	indexValue.Set(reflectValue)
	return true
}

func goResizeableSliceGetOwnProperty(self *_object, name string) *_property {
	// length
	if name == "length" {
		return &_property{
			value: toValue(self.value.(*_goResizeableSliceObject).getLength()),
			mode:  0,
		}
	}

	// .0, .1, .2, ...
	index := stringToArrayIndex(name)
	if index >= 0 {
		object := self.value.(*_goResizeableSliceObject)
		value := Value{}
		reflectValue, exists := object.getValue(index)
		if exists {
			value = self.runtime.toValue(reflectValue.Interface())
		}
		return &_property{
			value: value,
			mode:  0110,
		}
	}

	return objectGetOwnProperty(self, name)
}

func goResizeableSliceEnumerate(self *_object, all bool, each func(string) bool) {
	object := self.value.(*_goResizeableSliceObject)
	// .0, .1, .2, ...

	for index, length := 0, object.getLength(); index < length; index++ {
		name := strconv.FormatInt(int64(index), 10)
		if !each(name) {
			return
		}
	}

	objectEnumerate(self, all, each)
}

func goResizeableSliceDefineOwnProperty(self *_object, name string, descriptor _property, throw bool) bool {
	if name == "length" {
		//return true
		return self.runtime.typeErrorResult(throw)
	} else if index := stringToArrayIndex(name); index >= 0 {
		if self.value.(*_goResizeableSliceObject).setValue(index, descriptor.value.(Value)) {
			return true
		}
		return self.runtime.typeErrorResult(throw)
	}
	return objectDefineOwnProperty(self, name, descriptor, throw)
}

func goResizeableSliceDelete(self *_object, name string, throw bool) bool {
	// length
	if name == "length" {
		return self.runtime.typeErrorResult(throw)
	}

	// .0, .1, .2, ...
	index := stringToArrayIndex(name)
	if index >= 0 {
		object := self.value.(*_goResizeableSliceObject)
		indexValue, exists := object.getValue(index)
		if exists {
			indexValue.Set(reflect.Zero(reflect.Indirect(object.value).Type().Elem()))
			return true
		}
		return self.runtime.typeErrorResult(throw)
	}

	return self.delete(name, throw)
}
