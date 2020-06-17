package main

import (
	"net/http"

	luapi "github.com/fjah/LuAPI"
	"github.com/gin-gonic/gin"
	lua "github.com/yuin/gopher-lua"
)

type ginRouter struct {
	engine *gin.Engine
}

type ginContext struct {
	original *gin.Context
}

func (c ginContext) Body() luapi.RequestBody {
	req := luapi.RequestBody{}
	if err := c.original.BindJSON(&req); err != nil {
		panic(err)
	}
	return req
}

func (c ginContext) Respond(resp luapi.ResponseBody) {
	c.original.JSON(resp.Status, resp)
}

func (r ginRouter) POST(path string, handler func(luapi.Context)) {
	ginHandler := func(c *gin.Context) {
		ctx := ginContext{original: c}
		handler(ctx)
	}
	r.engine.POST(path, ginHandler)
}

func main() {
	router := gin.New()
	api := luapi.New(ginRouter{engine: router})

	api.Handlers["global"] = luapi.Handlers{
		Req: func(l *lua.LState, script string) error {
			return l.DoString(script)
		},
		Res: func(c luapi.Context) lua.LGFunction {
			return func(state *lua.LState) int {
				c.Respond(luapi.ResponseBody{
					Status: http.StatusOK,
					Body:   state.ToString(1),
				})
				return 0
			}
		},
	}
	api.Setup(true)

	router.Run(":80")
}
