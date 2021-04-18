package ginutils

import (
	"fmt"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

var v = validator.New()

type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// for response
type StatusI interface {
	Status() int
}
type ContentTypeI interface {
	ContentType() string
}

// ErrCodeI interface for error with non-500 status
type ErrCodeI interface {
	error
	StatusI
}

// errCode implements ErrCodeI
type errCode struct {
	text string
	code int
}

func (e *errCode) Error() string {
	return e.text
}
func (e *errCode) Status() int {
	return e.code
}

func NewError(code int, text string) ErrCodeI {
	return &errCode{
		text: text,
		code: code,
	}
}

// data implements CodeI and ContentTypeI
type Data struct {
	Data  interface{} `json:",:inline"`
	Code  int
	CType string
}

// implements interface
func (d *Data) Status() int {
	return d.Code
}
func (d *Data) ContentType() string {
	return d.CType
}

// creating response
func Resp(resp interface{}) *Data {
	return &Data{
		Data:  resp,
		Code:  200,
		CType: "application/json",
	}
}
func (d *Data) WithCode(code int) *Data {
	d.Code = 200
	return d
}
func (d *Data) WithCType(ctype string) *Data {
	d.CType = ctype
	return d
}

//func InjectGin(handler interface{}) func(c *gin.Context) {
//	// handlerFunc should have signature
//	// func(c *gin.Context, input <struct>)
//
//	// type checkings
//	handlerT := reflect.TypeOf(handler)
//	if handlerT.Kind() != reflect.Func {
//		panic("handler must be function")
//	}
//	if handlerT.NumIn() != 2 {
//		panic("handler must have 2 arguments")
//	}
//
//	if handlerT.In(0) != reflect.TypeOf((*gin.Context)(nil)) {
//		panic("first argument must be *gin.Context")
//	}
//
//	// validate second struct argument
//	arg2T := handlerT.In(1)
//	if arg2T.Kind() != reflect.Struct {
//		panic("second argument must be struct type")
//	}
//	if _, ok := arg2T.FieldByName("Body"); !ok {
//		panic("field `Body` not found in struct")
//	}
//
//	// validate output
//	if handlerT.NumOut() != 0 {
//		panic("function shouldn't return anything")
//	}
//
//	handlerV := reflect.ValueOf(handler)
//
//	return func(c *gin.Context) {
//		// prepare injection data
//		argV := reflect.New(arg2T)
//		fieldPtr := argV.Elem().FieldByName("Body").Addr().Interface()
//
//		if err := c.BindJSON(fieldPtr); err != nil {
//			c.JSON(400, Response{
//				Success: false,
//				Message: "JSON binding failed",
//				Data: err.Error(),
//			})
//			return
//		}
//
//		if err := v.Struct(argV.Elem().Interface()); err != nil {
//			c.JSON(400, Response{
//				Success: false,
//				Message: "body validation failed",
//				Data: err.Error(),
//			})
//			return
//		}
//
//		handlerV.Call([]reflect.Value{
//			reflect.ValueOf(c),
//			argV.Elem(),
//		})
//	}
//}
//

func InjectGin(handler interface{}) func(c *gin.Context) {
	// handlerFunc should have signature
	// func(c *gin.Context, input <struct>)

	// type checkings
	handlerT := reflect.TypeOf(handler)
	if handlerT.Kind() != reflect.Func {
		panic("handler must be function")
	}
	if handlerT.NumIn() != 2 {
		panic("handler must have 2 arguments")
	}

	if handlerT.In(0) != reflect.TypeOf((*gin.Context)(nil)) {
		panic("first argument must be *gin.Context")
	}

	// validate second struct argument
	arg2T := handlerT.In(1)
	if arg2T.Kind() != reflect.Struct {
		panic("second argument must be struct type")
	}

	bindBody := false
	if _, ok := arg2T.FieldByName("Body"); ok {
		bindBody = true
	}

	errorT := reflect.TypeOf((*error)(nil)).Elem()
	if handlerT.NumOut() != 1 {
		panic("function should return 1 value: error ")
	} else if !handlerT.Out(0).AssignableTo(errorT) {
		panic("arg1 should be assignable to error")
	}

	handlerV := reflect.ValueOf(handler)

	return func(c *gin.Context) {
		// prepare injection data
		argV := reflect.New(arg2T)

		if bindBody {
			fieldPtr := argV.Elem().FieldByName("Body").Addr().Interface()

			if err := c.BindJSON(fieldPtr); err != nil {
				c.JSON(400, Response{
					Success: false,
					Message: "JSON binding failed",
					Data:    err.Error(),
				})
				return
			}

			if err := v.Struct(argV.Elem().Interface()); err != nil {
				c.JSON(400, Response{
					Success: false,
					Message: "body validation failed",
					Data:    err.Error(),
				})
				return
			}
		}

		// call real handler
		err := handlerV.Call([]reflect.Value{
			reflect.ValueOf(c),
			argV.Elem(),
		})[0].Interface()
		if err != nil {
			e := err.(error)
			fmt.Println("handle error")
			status := 500
			if ec, ok := err.(StatusI); ok {
				fmt.Println("err is status")
				status = ec.Status()
			}
			c.JSON(status, Response{
				Success: false,
				Message: e.Error(),
			})
		}

	}
}
