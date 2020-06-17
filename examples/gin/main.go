package main

import (
	"net/http"

	"github.com/fjah/luapi"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

// We need to implement the Router and Context interfaces of LuAPI. This means that the library can essentially
// support all routers.
type ginRouter struct {
	engine *gin.Engine
}

type ginContext struct {
	original *gin.Context
}

// Body should just return a RequestBody of the parsed application/json data.
func (c ginContext) Body() luapi.RequestBody {
	req := luapi.RequestBody{}
	if err := c.original.BindJSON(&req); err != nil {
		panic(err)
	}
	return req
}

// Respond should send an application/json response with the status code as Status and the body as the struct.
// The ResponseBody struct already has JSON tags, so there's no need to do that yourself.
func (c ginContext) Respond(resp luapi.ResponseBody) {
	c.original.JSON(resp.Status, resp)
}

// POST simply allows LuAPI to register POST requests.
func (r ginRouter) POST(path string, handler func(luapi.Context)) {
	ginHandler := func(c *gin.Context) {
		// We need to make sure to instantiate a ginContext and call handler here.
		ctx := ginContext{original: c}
		handler(ctx)
	}
	r.engine.POST(path, ginHandler)
}

func main() {
	router := gin.New()
	// Once we're done, just instantiate LuAPI with our router.
	api := luapi.New(ginRouter{engine: router})

	// The global handler is the one used when a namespace isn't provided. Namespaces are essentially way
	// to virtualise endpoints; they can have a set amount of functions, etc.
	api.Handlers["global"] = luapi.Handlers{
		// Req is called when a request is made to the server.
		Req: func(l *lua.LState, script string) error {
			// This simply tells the Lua state to execute the script. Note that the bootstrapper has already
			// been called.
			return l.DoString(script)
		},
		// Once Req finishes, Res' *closure* will be called.
		Res: func(c luapi.Context) lua.LGFunction {
			// Lets return a LGFunction; this lets us access variables passed in Lua.
			return func(state *lua.LState) int {
				// Just respond to the request with the first parameter passed to the `respond` Lua function as a
				// string. You can also push (return) values with state.Push; make sure to update `return 0` to be
				// the amount of returned values.
				c.Respond(luapi.ResponseBody{
					Status: http.StatusOK,
					Body:   state.ToString(1),
				})
				return 0
			}
		},
	}
	// Set up the API, running a test.
	api.Setup(true)

	router.Run(":80")
}
