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

### Init
- `git clone https://github.com/yegorov-boris/multiplexer.git`
- `cd multiplexer`
- `cp .env.dist .env`

### Run
Set http server port and path to .json output files directory. Build and run multiplexer in docker.

Example:

```
docker build -t multiplexer . && docker run --rm -p 80:8080 -v ./store:/store --name=multiplexer-1 multiplexer
```

80 is port on your host multiplexer will work on

8080 is copied from HTTP_PORT in .env

store is copied from STORE_PATH in .env

### Test

#### Run autotests

Set multiplexer and test server http ports. Build and run tests in docker.

Example:

```
docker build -f ./Dockerfile.test -t multiplexer . && docker run --rm -p 8080:8080 -p 8081:8081 multiplexer
```

8080 is copied from HTTP_PORT in .env

8081 is copied from TEST_PORT in .env

#### Test manually

Build and run multiplexer in docker on default http port (usually 80). Consider `HTTP_BASE_PATH=/api/v1` in your .env file.

Example:

```
docker build -t multiplexer . && docker run --rm -p 80:8080 -v ./store:/store --name=multiplexer-1 multiplexer
```

Go to http://127.0.0.1/api/v1/docs

Send POST to create a task

Then GET results by id

Or send DELETE to cancel in-progress task by id