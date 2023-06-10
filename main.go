package main

import "micro/net/client"

func main() {
	//server.Serve("127.0.0.1:8080")
	client.Connect("127.0.0.1:8080")
}
