# LuAPI

LuAPI is a tiny but powerful library that allows you to create easily-accessible web APIs with Go. Clients can `POST` to your API with a Lua script; the environment that you defined will be able to be interfaced with. You write **Go**, and clients can send **Lua scripts** to use that code.

**Note**: LuAPI does *not* have a sandbox yet; clients have access to the entire stdlib and have the ability to — for example — nuke files. You'll need to find a way to make your own while it isn't implemented:

- [Stack Overflow post](https://stackoverflow.com/questions/1224708/how-can-i-create-a-secure-lua-sandbox)
- [lua-users.org](http://lua-users.org/wiki/SandBoxes)
- [sandbox.lua](https://github.com/kikito/sandbox.lua)

Make sure to thoroughly test your sandbox before deploying it for production.

## Usage

```
go get github.com/fjah/luapi
```

LuAPI uses interfaces so that it can be used with all routers; simply implement them. If you're using [gin](https://github.com/gin-gonic/gin), you can use [this example](https://github.com/fjah/luapi/tree/master/examples/gin).

In these examples, `lua` refers to the `github.com/yuin/gopher-lua` package. Make sure to import this.

To get started, you'll want to define structs for your HTTP router.

```go
type ginRouter struct {
	engine *gin.Engine
}

type ginContext struct {
	original *gin.Context
}
```

Note that `original` can really be anything; if you were using net/http's router, you'd likely make the context struct have `request` and `writer`. Next, you'll need to implement LuAPI's `Context` and `Router` interfaces.

Let's start off with `luapi.Context.Body`:

```go
func (c ginContext) Body() luapi.RequestBody {
	req := luapi.RequestBody{
		Script: `respond("Invalid request body.")`,
	}
	_ = c.original.BindJSON(&req)
	return req
}
```

This should simply return the request as a `luapi.RequestBody`. In this example, we used gin's `BindJSON` to get the JSON body of the request and bind it to the struct. We also define a default script; this returns "Invalid request body."

Now let's implement `luapi.Context.Respond`:

```go
func (c ginContext) Respond(resp luapi.ResponseBody) {
	c.original.JSON(resp.Status, resp)
}
```

We're simply responding to the request with the status code of `Status` and the JSON body of the marshalled `luapi.ResponseBody` struct. Note that it already has JSON struct tags; there's no need to add them yourself.

We need to do the same for `luapi.Router.POST`:

```go
func (r ginRouter) POST(path string, handler func(luapi.Context)) {
	r.engine.POST(path, func(c *gin.Context) {
		ctx := ginContext{original: c}
		handler(ctx)
	})
}
```

This should simply register a `POST` endpoint for `path` that calls the `handler` with your context.

All of the interfaces have now been implemented; let's use them. First, let's define a `luapi.LuAPI`. You can use the `luapi.New` helper or instantiate the struct yourself. Note that `router` is an example for gin.

```go
router := gin.New()
api := luapi.New(ginRouter{engine: router})
```

Once we've done that, let's set it up:

```go
api.Setup(true)
```

This will register routes, etc, and run a test. You can disable the test by replacing `true` with `false`.

That's all of the setup finished. Now we can define functions that can be used by clients in Lua. LuAPI has namespaces. Basically, these are a way to create multiple Lua environments. The default namespace that'll be used is `global`; let's define it!

```go
api.Handlers["global"] = luapi.Handlers{
	// Req is called when a request is made to the server.
	Req: func(l *lua.LState, script string) error {
		// Set a function called test that returns one string.
		l.SetGlobal("test", l.NewFunction(func(state *lua.LState) int {
			// Push simply returns a value to Lua.
			state.Push(lua.LString("Test succeeded!"))
			// We need to state how many values we returned.
			return 1
		}))

		// This simply tells the Lua state to execute the script.
		return l.DoString(script)
	},
	// Once Req finishes, Res' *closure* will be called.
	Res: func(c luapi.Context) lua.LGFunction {
		called := 0

		return func(state *lua.LState) int {
			// Checking against called lets us limit the amount of calls to `respond`. We really only want one response.
			if called == 0 {
				c.Respond(luapi.ResponseBody{
					Status: http.StatusOK,
					// state.ToString(1) is the first parameter passed to the Lua `respond` function as a string.
					Body:   state.ToString(1),
				})
			}

			called++
			// We haven't returned any values, so state that the amount we returned is 0.
			return 0
		}
	},
}
```

And finally, we can start our server!

```go
router.Run(":80")
```

Now that our server is up, we should test it. [The LuAPI sandbox](https://luapi.owo.gg) is really useful for this, but note that it'll only work if your server address isn't local. In the case that it is, we can send a `POST` request to `/` with the `application/json` body of

```json
{
  "script": "respond(test())"
}
```

This responds with the result of the `test` function we made. We can choose what the response is. If this was a success, you would've received a response like this:

```json
{
  "status": 200,
  "body": "Test succeeded!"
}
```

Now it's up to you to define more functions, namespaces, etc. If your request JSON body includes a `namespace` key, it'll be used.
