# Outyet: A Go Version Release Checker

This project is a simple Go web server that checks if a specific version of Go has been released. It provides a clear "YES" or "NO" answer on a webpage.

## How it Works

The application works as follows:

1.  **Web Server**: It runs a web server that listens on a specified port.
2.  **Background Polling**: A background goroutine periodically checks the official Go source repository to see if the specified Go version tag is available.
3.  **Simple UI**: It serves a minimal HTML page that displays "YES!" if the Go version is out, and "No. :-(" if it is not.

## How to Run

To run the server, use the following command:

```bash
go run main.go
```

You can customize the server's behavior with these command-line flags:

*   `-addr`: The address and port to listen on (default: `0.0.0.0:8080`).
*   `-poll`: The time interval for polling for the Go version (default: `5s`).
*   `-version`: The Go version to check for (default: `1.20`).

### Example

To check for Go version 1.21 and serve on port 9000, you would run:

```bash
go run main.go -version=1.21 -addr=0.0.0.0:9000
```

## Project Structure

*   `main.go`: Contains the main application logic, including the web server and the polling mechanism.
*   `go.mod` and `go.sum`: Manage the project's dependencies.
*   `README.md`: This file.
