package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"httpfromtcp/internal/headers"
	"httpfromtcp/internal/request"
	"httpfromtcp/internal/response"
	"httpfromtcp/internal/server"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
)

const port = 42069
const httpbinUrl = "https://httpbin.org"

func main() {
	server, err := server.Serve(port, handler)
	if err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
	defer server.Close()
	fmt.Println("Server started on port", port)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	fmt.Println("Server gracefully stopped")
}

func handler(w *response.Writer, req *request.Request) {
	if strings.HasPrefix(req.RequestLine.RequestTarget, "/httpbin") {
		handlerHttpbin(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/yourproblem" {
		handler400(w, req)
		return
	}
	if req.RequestLine.RequestTarget == "/myproblem" {
		handler500(w, req)
		return
	}
	handler200(w, req)
	return
}

func handlerHttpbin(w *response.Writer, req *request.Request) {
	s := strings.TrimPrefix(req.RequestLine.RequestTarget, "/httpbin")
	fullUrl := httpbinUrl + s
	resp, err := http.Get(fullUrl)
	if err != nil {
		handler500(w, req)
	}
	defer resp.Body.Close()

	w.WriteStatusLine(response.StatusCodeSuccess)
	h := response.GetDefaultHeaders(0)
	h.Remove("Content-Length")
	h.Set("Trailer", "X-Content-SHA256")
	h.Set("Trailer", "X-Content-Length")
	w.WriteHeaders(h)

	buf := make([]byte, 1024)
	totalBody := []byte("")
	for {
		n, err := resp.Body.Read(buf)
		fmt.Println("Read", n, "bytes")
		if n > 0 {
			_, err = w.WriteChunkedBody(buf[:n])
			if err != nil {
				fmt.Println("Error writing chunked body:", err)
				break
			}
			totalBody = append(totalBody, buf[:n]...)
		}
		if err == io.EOF {
			break // all data read
		}
		if err != nil {
			fmt.Println("Error reading response body:", err)
			break
		}
	}
	_, err = w.WriteChunkedBodyDone()
	if err != nil {
		fmt.Println("Error writing chunked body done:", err)
	}
	bodyHash := sha256.Sum256(totalBody)
	trailers := headers.NewHeaders()
	trailers.Set("X-Content-Length", strconv.Itoa(len(totalBody)))
	trailers.Set("X-Content-SHA256", hex.EncodeToString(bodyHash[:]))
	err = w.WriteTrailers(trailers)
	if err != nil {
		fmt.Errorf("error whilst writing trailers to client, %s", err)
	}
	return

}

func handler400(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeBadRequest)
	body := []byte(`<html>
<head>
<title>400 Bad Request</title>
</head>
<body>
<h1>Bad Request</h1>
<p>Your request honestly kinda sucked.</p>
</body>
</html>
`)
	headers := response.GetDefaultHeaders(len(body))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(body)
	return
}

func handler500(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeInternalServerError)
	body := []byte(`<html>
<head>
<title>500 Internal Server Error</title>
</head>
<body>
<h1>Internal Server Error</h1>
<p>Okay, you know what? This one is on me.</p>
</body>
</html>
`)
	headers := response.GetDefaultHeaders(len(body))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(body)
	return
}

func handler200(w *response.Writer, _ *request.Request) {
	w.WriteStatusLine(response.StatusCodeSuccess)
	body := []byte(`<html>
<head>
<title>200 OK</title>
</head>
<body>
<h1>Success!</h1>
<p>Your request was an absolute banger.</p>
</body>
</html>
`)
	headers := response.GetDefaultHeaders(len(body))
	headers.Override("Content-Type", "text/html")
	w.WriteHeaders(headers)
	w.WriteBody(body)
	return
}
