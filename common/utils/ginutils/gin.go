package ginutils

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"reflect"
)

var v = validator.New()

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
	if _, ok := arg2T.FieldByName("Body"); !ok {
		panic("field `Body` not found in struct")
	}

	// validate output
	if handlerT.NumOut() != 0 {
		panic("function shouldn't return anything")
	}

	handlerV := reflect.ValueOf(handler)

	return func(c *gin.Context) {
		// prepare injection data
		argV := reflect.New(arg2T)
		fieldPtr := argV.Elem().FieldByName("Body").Addr().Interface()

		if err := c.BindJSON(fieldPtr); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "JSON binding error",
				"data":    err.Error(),
			})
			return
		}

		if err := v.Struct(argV.Elem().Interface()); err != nil {
			c.JSON(400, gin.H{
				"success": false,
				"message": "body validation failed",
				"data":    err.Error(),
			})
			return
		}

		// call real handler
		handlerV.Call([]reflect.Value{
			reflect.ValueOf(c),
			argV.Elem(),
		})
	}
}
