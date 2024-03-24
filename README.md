# HTTP multiplexer written in Go

### Task
The application is an HTTP server.
A POST handler gets a list of JSON-encoded URLs.
The server gets those links and responds with JSON-encoded results.
If at least one outgoing request fails, processing of any other URL stops, and the server responds with a text error message.
#### Limitations:
- use Go 1.18 or later
- use Go standard library only
- rate-limit POST requests, e.g. no more than 100 concurrent
- limit number of links per POST request, e.g. no more than 20
- limit outgoing GET requests per incoming POST requests, e.g. no more than 4
- set timeout for any outgoing GET request, e.g. 1 second
- clients can cancel processing of their requests
- implement graceful shutdown

### Prerequisites
- linux
- docker

### Run
Set http server port and path to .json output files directory, build and run multiplexer in docker.

Example:

```
export STORE_PATH=store
export HTTP_PORT=8080
docker build -t multiplexer . && docker run --rm -p 127.0.0.1:$HTTP_PORT:$HTTP_PORT/tcp --volume=./$STORE_PATH:/$STORE_PATH --name=multiplexer-1 multiplexer
```

### Test
Build and run tests in docker.

Example:

```
docker build -f ./Dockerfile.test -t multiplexer . && docker run --rm multiplexer
```