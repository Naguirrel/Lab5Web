package main

import (
	"bufio"
	"fmt"
	"html"
	"log"
	"net"
	"strings"
)

func main() {
	ln, err := net.Listen("tcp", ":8080")
	if err != nil {
		log.Fatal("Error escuchando en :8080:", err)
	}
	defer ln.Close()

	log.Println("Servidor TCP escuchando en http://localhost:8080")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Println("Error aceptando conexión:", err)
			continue
		}

		go handleConn(conn)
	}
}

func handleConn(conn net.Conn) {
	defer conn.Close()

	reader := bufio.NewReader(conn)

	// 1) Leer la primera línea del request HTTP
	line, err := reader.ReadString('\n')
	if err != nil {
		// Si el cliente cerró o hubo error, simplemente salimos
		return
	}

	// Ejemplo: "GET /hello HTTP/1.1\r\n"
	line = strings.TrimRight(line, "\r\n")
	parts := strings.Fields(line) // ["GET", "/hello", "HTTP/1.1"]

	path := "/"
	if len(parts) >= 2 {
		path = parts[1]
	}

	// (Opcional) log para ver qué llegó
	log.Printf("Request line: %q -> path=%q\n", line, path)

	// 2) Construir el body HTML
	escapedPath := html.EscapeString(path)
	body := fmt.Sprintf(`<!doctype html>
<html>
  <head><meta charset="utf-8"><title>mini server</title></head>
  <body>
    <h1>Hello!</h1>
    <p>You requested: <b>%s</b></p>
  </body>
</html>`, escapedPath)

	// 3) Construir respuesta HTTP manualmente (¡con CRLF!)
	contentLength := len([]byte(body))

	response := fmt.Sprintf(
		"HTTP/1.1 200 OK\r\n"+
			"Content-Type: text/html; charset=utf-8\r\n"+
			"Content-Length: %d\r\n"+
			"Connection: close\r\n"+
			"\r\n"+
			"%s",
		contentLength, body,
	)

	// 4) Enviar respuesta al cliente
	_, _ = conn.Write([]byte(response))
}
