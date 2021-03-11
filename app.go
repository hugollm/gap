package gap

import (
	"log"
	"net/http"
)

// App is the fundamental building block for applications
type App struct {
	routes       map[string]route
	errorHandler func(interface{}, http.ResponseWriter)
}

type route struct {
	method   string
	endpoint endpoint
}

// New is the proper way to create a new App
func New() *App {
	return &App{
		routes:       map[string]route{},
		errorHandler: defaultErrorHandler,
	}
}

func defaultErrorHandler(ierr interface{}, response http.ResponseWriter) {
	log.Print("PANIC: ", ierr)
	response.Header().Set("Content-Type", "application/json")
	response.WriteHeader(500)
	response.Write([]byte(`{"error":"internal server error"}`))
}

// Route binds request method and path to target endpoint
func (app *App) Route(method string, path string, fn interface{}) {
	app.routes[path] = route{method, newEndpoint(fn)}
}

// ErrorHandler allows replacing of the default error handler
func (app *App) ErrorHandler(handler func(interface{}, http.ResponseWriter)) {
	app.errorHandler = handler
}

// ServeHTTP fullfills the http.Handler interface implementation
func (app *App) ServeHTTP(response http.ResponseWriter, request *http.Request) {
	defer writeErrorOnPanic(response, app.errorHandler)
	route, found := app.routes[request.URL.Path]
	if !found {
		writeNotFound(response)
		return
	}
	if route.method != request.Method {
		writeMethodNotAllowed(response)
		return
	}
	route.endpoint.handle(request, response)
}

// Run is a shortcut for starting a web server for your app
func (app *App) Run() {
	log.Fatal(http.ListenAndServe(":8000", app))
}

func writeNotFound(response http.ResponseWriter) {
	response.WriteHeader(404)
	response.Write([]byte(`{"error":"not found"}`))
}

func writeMethodNotAllowed(response http.ResponseWriter) {
	response.WriteHeader(405)
	response.Write([]byte(`{"error":"method not allowed"}`))
}

func writeErrorOnPanic(httpResponse http.ResponseWriter, errorHandler func(interface{}, http.ResponseWriter)) {
	ierr := recover()
	if ierr != nil {
		errorHandler(ierr, httpResponse)
	}
}
