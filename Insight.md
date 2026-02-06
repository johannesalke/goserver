# Mux
The mux decodes incoming requests by URL path and routs them to the correct handler.

# Handlers

Handlers process the incoming requests. They have to be attached to a Mux via a Handle or HandleFunc method called on the mux (or the base http object). 
The two methods both attach a handler to the mux, but have different inputs. 


## HandleFunc
This method attaches an often self-created handler function to the mux
mux.HandleFunc(path string, funcName func(w http.ResponseWriter, r *http.Request))

## Handle
This method attaches something else, such as a fileserver object/process to the mux. 


These two shouldn't be confused with the types/function archetypes Handler and HandlerFunc, which they act on to attach them to the mux. 


## Handlerfunc
type HandlerFunc func(ResponseWriter, *Request)
Request contains all the information of the request, ready to be used for output generation.
ResponseWriter is an object with methods that determine the response. Specifically, it has one method each for Header, Statuscode and Body. 

- w.Header().Set("Content-Type", "text/plain; charset=utf-8") | The header is a dictionary/map. Each application of .Set() inserts one entry into it.
		
- w.WriteHeader(200) | This one sets the status code. If unspecified, it defaults to 200.

- w.Write([]byte(arg)) | with arg being a string (?) that contains the response.


# Middleware
Middleware is a way to wrap a handler with additional functionality. 
For example, we can write a middleware that logs every request to the server. We can then wrap our handler with this middleware and every request will be logged without us having to write the logging code in every handler.

To do that, we can write the middleware function like this:
```
func middlewareLog(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
```
Then, any handler that needs logging can be wrapped by this middleware function:

mux.Handle("/app/", middlewareLog(handler))