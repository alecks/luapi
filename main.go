package luapi

import (
	"net/http"

	lua "github.com/yuin/gopher-lua"
)

type namespaceHandler = func(l *lua.LState, script string) error

// ResponseBody is a response sent by the API. Note that this will also be used for errors.
// The client should receive an application/json response with {"status":Status,"body":Body}. The status code
// should be set to Status.
type ResponseBody struct {
	Status int    `json:"status"`
	Body   string `json:"body"`
}

// RequestBody is an application/json body that LuAPI supports.
type RequestBody struct {
	// Namespace is the "endpoint" that the script will be executed for. Not required.
	Namespace string `json:"for"`
	// Script is the script that'll be run in the Lua sandbox. Required.
	Script string `json:"script"`
}

// Context is a generic request context.
type Context interface {
	Body() RequestBody
	Respond(ResponseBody)
}

// Router is a generic router used by LuAPI. This should be implemented.
type Router interface {
	POST(path string, handler func(ctx Context))
}

// Handlers are functions called on request and result.
type Handlers struct {
	Req namespaceHandler
	Res func(Context) lua.LGFunction
}

// LuAPI is the main server.
type LuAPI struct {
	// Router is the required router to be used by the server.
	Router Router
	// Lua is the gopher-lua sandbox used by LuAPI.
	Lua *lua.LState
	// Bootstrapper is the script to be run when setting up the Lua sandbox. This isn't a filepath;
	// see BootstrapperFile. BootstrapperFile will be prioritised. By default, this is a script making the
	// stdlib nil.
	Bootstrapper string
	// BootstrapperFile is the filepath to a Lua script that'll be run when setting up the sandbox.
	// This will be prioritised over Bootstrapper.
	BootstrapperFile string
	// Handlers are the user-defined script handlers.
	Handlers map[string]Handlers
}

// New instantiates LuAPI. To customise this, use the LuAPI struct yourself.
func New(router Router) *LuAPI {
	luaState := lua.NewState()
	api := LuAPI{
		Router:       router,
		Lua:          luaState,
		Bootstrapper: "debug = nil; io = nil; os = nil",
		Handlers:     make(map[string]Handlers),
	}

	return &api
}

// Setup runs the bootstrapper, sets routes and optionally runs a test script.
func (api *LuAPI) Setup(runTest bool) error {
	if err := api.bootstrap(api.Lua); err != nil {
		return err
	}
	api.Router.POST("/", api.mainHandler)

	if runTest {
		return api.Lua.DoString(`print("LuAPI setup: test succeeded. Executed from Lua.")`)
	}
	return nil
}

func (api *LuAPI) mainHandler(c Context) {
	body := c.Body()
	if body.Script == "" {
		c.Respond(ResponseBody{
			Status: http.StatusBadRequest,
			Body:   "`script` is required",
		})
		return
	}

	namespace := "global"
	if body.Namespace != "" {
		namespace = body.Namespace
	}
	if handlers, ok := api.Handlers[namespace]; ok {
		l := lua.NewState()
		api.bootstrap(l)
		l.SetGlobal("respond", l.NewFunction(handlers.Res(c)))

		if err := handlers.Req(l, body.Script); err != nil {
			c.Respond(ResponseBody{
				Status: http.StatusBadRequest,
				Body:   err.Error(),
			})
		}
		return
	}

	c.Respond(ResponseBody{
		Status: http.StatusNotFound,
		Body:   "Namespace doesn't exist: " + namespace,
	})
}

func (api *LuAPI) bootstrap(l *lua.LState) error {
	if api.BootstrapperFile == "" {
		return api.Lua.DoString(api.Bootstrapper)
	}

	return api.Lua.DoFile(api.BootstrapperFile)
}
