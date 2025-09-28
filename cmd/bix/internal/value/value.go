package value

import (
	"fmt"
	"github.com/billchow98/bixscript/cmd/bix/internal/assert"
	"log"
)

type Boolean bool

type Value struct {
	data Pointer
}

func fromObject(p Pointer) (v Value) {
	v.data = p
	return v
}

func (v Value) asObject() Pointer {
	return v.data
}

func (v Value) tag() tag {
	return *(*tag)(v.asObject())
}

func (v Value) String() string {
	switch v.tag() {
	case BixNumber:
		return fmt.Sprintf("%.16g", v.AsNumber())
	case BixBoolean:
		return fmt.Sprintf("%v", v.AsBoolean())
	case BixString:
		return fmt.Sprintf("%v", v.AsString())
	case BixUndefined:
		return "undefined"
	case BixFunction:
		if (*v.AsFunction().Name).s == "__anonymous__" {
			return "<unnamed function>"
		}
		return fmt.Sprintf("<function %s>", v.AsFunction().Name)
	case BixInteger, BixIdentifier, BixPointer:
		assert.Assert(false)
		return "" // Unreachable
	case BixNativeFunction:
		return fmt.Sprintf("<native fun %s>", v.AsNativeFunction().Name)
	default:
		assert.Assert(false) // Unreachable
		return ""
	}
}

func (v Value) DebugString() string {
	switch v.tag() {
	case BixInteger:
		return fmt.Sprintf("%v", v.AsInteger())
	case BixString:
		return fmt.Sprintf("%q", v.AsString())
	case BixIdentifier:
		return fmt.Sprintf("%s", v.AsIdentifier())
	default:
		return v.String()
	}
}

func (v Value) SameTypeAs(rhs Value) bool {
	return v.tag() == rhs.tag()
}

func (v Value) Equals(rhs Value) Boolean {
	if !v.SameTypeAs(rhs) {
		return false
	}

	switch v.tag() {
	case BixInteger:
		return v.AsInteger() == rhs.AsInteger()
	case BixNumber:
		return v.AsNumber() == rhs.AsNumber()
	case BixBoolean, BixIdentifier, BixFunction, BixPointer, BixNativeFunction:
		return v.asObject() == rhs.asObject()
	case BixString:
		return v.AsString() == rhs.AsString() // We have no string interning for now
	case BixUndefined:
		return true
	default:
		assert.Assert(false) // Unreachable
		return false
	}
}

func FromInteger(i Integer) Value {
	return fromObject(Pointer(newHeapInteger(i)))
}

func (v Value) IsInteger() bool {
	if v.tag() != BixInteger {
		log.Fatalf("v: %#v, v.tag(): %#v\n", v, v.tag())
	}
	return v.tag() == BixInteger
}

func (v Value) AsInteger() Integer {
	assert.Assert(v.IsInteger())
	o := *(*heapInteger)(v.asObject())
	return o.i
}

func FromNumber(n Number) Value {
	return fromObject(Pointer(newHeapNumber(n)))
}

func (v Value) IsNumber() bool {
	return v.tag() == BixNumber
}

func (v Value) AsNumber() Number {
	assert.Assert(v.IsNumber())
	o := *(*heapNumber)(v.asObject())
	return o.n
}

func FromBoolean(b Boolean) Value {
	if b {
		return fromObject(Pointer(trueObject))
	}
	return fromObject(Pointer(falseObject))
}

func (v Value) IsBoolean() bool {
	return v.asObject() == Pointer(trueObject) || v.asObject() == Pointer(falseObject)
}

func (v Value) AsBoolean() Boolean {
	assert.Assert(v.IsBoolean())
	return v.asObject() == Pointer(trueObject)
}

func FromString(s String) Value {
	return fromObject(Pointer(newHeapString(s)))
}

func (v Value) IsString() bool {
	return v.tag() == BixString
}

func (v Value) AsString() String {
	assert.Assert(v.IsString())
	o := *(*heapString)(v.asObject())
	return o.s
}

func (v Value) IsIdentifier() bool {
	return v.tag() == BixIdentifier
}

func (v Value) AsIdentifier() *Identifier {
	assert.Assert(v.IsIdentifier())
	return (*Identifier)(v.asObject())
}

func Undefined() Value {
	return fromObject(Pointer(undefinedObject))
}

func (v Value) IsUndefined() bool {
	return v.asObject() == Pointer(undefinedObject)
}

func FromFunction(f *Function) Value {
	return fromObject(Pointer(f))
}

func (v Value) IsFunction() bool {
	return v.tag() == BixFunction
}

func (v Value) AsFunction() *Function {
	assert.Assert(v.IsFunction())
	return (*Function)(v.asObject())
}

func FromPointer(p Pointer) Value {
	return fromObject(Pointer(newHeapPointer(p)))
}

func (v Value) IsPointer() bool {
	return v.tag() == BixPointer
}

func (v Value) AsPointer() Pointer {
	assert.Assert(v.IsPointer())
	o := *(*heapPointer)(v.asObject())
	return o.p
}

func FromNativeFunction(f *NativeFunction) Value {
	return fromObject(Pointer(f))
}

func (v Value) IsNativeFunction() bool {
	return v.tag() == BixNativeFunction
}

func (v Value) AsNativeFunction() *NativeFunction {
	assert.Assert(v.IsNativeFunction())
	return (*NativeFunction)(v.asObject())
}
