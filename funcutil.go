// Package funcutil is a simple utility to enable a method to be called by name.
//
// Registering methods
//
// The structs must be pointer type
//
// 		type service struct {
//			m string
//		}
//
//		func (s service)Hello() {}
// 		func (s *service)SetHello(m string) {
//			s.m = m
//		}
//
//		f := funcutil.New()
//		// must be pointer type
//		f.Register(&service{})
//
// Method for pointer or value receiver identification will be normalized into something:
// <struct name>.MethodName.
//
// Calling a method by name
//
// Method can be called by specified the normalized method name
//
//		f.Call("service.SetHello", "Hello world!")
//		f.Call("service.Hello")
//
package funcutil

import (
	"errors"
	"fmt"
	"log"
	"reflect"
	"strings"
	"sync"
)

type callInfo struct {
	argTypes  []reflect.Type
	retTypes  []reflect.Type
	m         *reflect.Method
	v         reflect.Value
	signature string
}

var (
	ErrMethodNotFound = errors.New("Method not found")
)

func (mi *callInfo) parametersMatch(params ...interface{}) error {
	var paramTypes []reflect.Type
	if len(mi.argTypes) > 1 {
		paramTypes = mi.argTypes[1:]
	}
	if len(params) != len(paramTypes) {
		return errors.New("Parameters mismatches")
	}
	for i, p := range paramTypes {
		pt := reflect.TypeOf(params[i])
		if pt == p {
			continue
		}
		if !pt.ConvertibleTo(p) {
			return fmt.Errorf("arguments: %v is not convertible to %v", pt, p)
		}
	}
	return nil
}

type FuncUtil struct {
	sync.Mutex
	calls map[string]callInfo
	ns    string
}

func (f *FuncUtil) getReturnTypes(t reflect.Type) []reflect.Type {
	if t.NumOut() == 0 {
		return nil
	}
	rets := []reflect.Type{}
	for i := 0; i < t.NumOut(); i++ {
		rets = append(rets, t.Out(i))
	}
	return rets
}

func (f *FuncUtil) getArgumentTypes(t reflect.Type) []reflect.Type {
	if t.NumIn() == 0 {
		return nil
	}
	params := []reflect.Type{}
	for i := 0; i < t.NumIn(); i++ {
		params = append(params, t.In(i))
	}
	return params
}

func (f *FuncUtil) generateSignature(name string, ci callInfo) string {
	args := []string{}
	if len(ci.argTypes) > 1 {
		argTypes := ci.argTypes[1:]
		for _, t := range argTypes {
			args = append(args, t.Name())
		}
	}
	rets := []string{}
	if len(ci.retTypes) > 0 {
		for _, t := range ci.retTypes {
			rets = append(rets, t.Name())
		}
	}
	ret := strings.Join(rets, ",")
	if len(rets) > 1 {
		ret = "(" + ret + ")"
	}
	return fmt.Sprintf("%s(%s) %s", name, strings.Join(args, ","), ret)
}

func (f *FuncUtil) register(s interface{}) {
	t := reflect.TypeOf(s)
	// element type
	et := t
	v := reflect.ValueOf(s)
	if t.Kind() == reflect.Ptr {
		et = t.Elem()
	}
	if !(et.Kind() == reflect.Struct && t.Kind() == reflect.Ptr) {
		log.Fatal("Type must be kind of *struct")
	}
	// enum methods
	for i := 0; i < t.NumMethod(); i++ {
		m := t.Method(i)
		// exclude non exported methods
		if m.PkgPath != "" {
			continue
		}
		// normalize the name regardless the receiver type
		namespace := ""
		if f.ns != "" {
			namespace = f.ns + "."
		}
		mn := fmt.Sprintf("%s%s.%s", namespace, et.Name(), m.Name)
		funcType := m.Func.Type()
		argTypes := f.getArgumentTypes(funcType)
		retTypes := f.getReturnTypes(funcType)
		mi := callInfo{
			argTypes: argTypes,
			retTypes: retTypes,
			m:        &m,
			v:        v,
		}
		mi.signature = f.generateSignature(mn, mi)
		f.calls[mn] = mi
	}
}

// Register registers the structs that implement the some exported methods.
// Each struct in vars must be pointer type
func (f *FuncUtil) Register(vars ...interface{}) {
	f.Lock()
	defer f.Unlock()
	for _, s := range vars {
		f.register(s)
	}
}

// Call invokes the registered methods using the matching arguments
// Argument type could be converted if they are convertible
func (f *FuncUtil) Call(methodName string, params ...interface{}) ([]interface{}, error) {
	f.Lock()
	defer f.Unlock()

	ci, exists := f.calls[methodName]
	if !exists {
		return nil, ErrMethodNotFound
	}
	err := ci.parametersMatch(params...)
	if err != nil {
		return nil, err
	}
	// exclude the receiver type
	argTypes := ci.argTypes[1:]
	// make first argument receiver value
	callParams := []reflect.Value{ci.v}
	// construct the rest arguments from supplied params
	for i, p := range params {
		serviceParamType := argTypes[i]
		callParamType := reflect.TypeOf(p)
		v := reflect.ValueOf(p)
		if callParamType != serviceParamType {
			// try to convert if they are convertible
			if callParamType.ConvertibleTo(serviceParamType) {
				v = v.Convert(serviceParamType)
			}
		}
		callParams = append(callParams, v)
	}
	// calls the method
	rets := ci.m.Func.Call(callParams)
	// verify the returned values whether they are compatible and convertible
	retValues := []interface{}{}
	for i, ret := range rets {
		retType := ci.retTypes[i]
		retValues = append(retValues, ret.Convert(retType).Interface())
	}
	if len(retValues) > 0 {
		return retValues, nil
	}
	return nil, nil
}

func (f *FuncUtil) Dump() []string {
	services := []string{}
	for _, v := range f.calls {
		services = append(services, v.signature)
	}
	return services
}

// New creates and returns new FuncUtil value
func New(vars ...string) *FuncUtil {
	ns := ""
	if len(vars) > 0 {
		ns = vars[0]
	}
	return &FuncUtil{
		calls: map[string]callInfo{},
		ns:    ns,
	}
}
