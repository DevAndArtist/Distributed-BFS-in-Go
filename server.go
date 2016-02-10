//
//  server.go
//
//  Created by Adrian Zubarev.
//  Copyright © 2016 Adrian Zubarev.
//  All rights reserved.
//

package main

import . "fmt"
import . "./graph"
import . "./array"
import . "./helper"
import . "./message"
import . "encoding/gob"
import . "./bfs/command"
import . "./identification"

import "net"
import "time"
import "math/rand"
import "strconv"
import "os"

import "sync"

type Server struct {
	Clients     *Array
	Complete    chan bool
	MessagePipe chan Message
}

type Client struct {
	Identification Identification
	Address        string
	Connection     net.Conn
	Encoder        *Encoder
	Decoder        *Decoder
}

var seed *rand.Rand

func init() {
	// initialize global time based seed
	seed = rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	// register for gob
	Register(Identification{})

	// register for array usage
	RegisterType(&Client{})
}

func main() {

	Println("\nStarting server ...")

	var arguments = os.Args[1:]

	var maxClientNumber = 3 // default value is 3

	if len(arguments) >= 1 {

		var parsedNumber, parseError = strconv.ParseUint(arguments[0], 10, 32)
		HandleError(parseError, func() {

			Println(parseError)
			os.Exit(1)
		})

		if int(parsedNumber) > maxClientNumber {

			maxClientNumber = int(parsedNumber)
		}
	}

	// start listening for clients
	var listener, listenerError = net.Listen("tcp", "localhost:8081")
	HandleError(listenerError, func() {

		Println(listenerError)
		os.Exit(2)
	})

	// listening for clients now
	Printf("[Log]: server will accept exact %d clients\n", maxClientNumber)

	// now we are safe to create and initialize the server instance
	var server = new(Server)
	server.Clients = ArrayOfType("*Client")
	server.Complete = make(chan bool)
	server.MessagePipe = make(chan Message)

	go server.HandleMessages()

	var waitGroup = new(sync.WaitGroup)

	// wait for all needed clients to join the network
	for server.Clients.Count() < maxClientNumber {

		Printf("[Log]: waiting for %d more clients\n", (maxClientNumber - server.Clients.Count()))

		// wait and accept new clients
		var newConnection, acceptingError = listener.Accept()
		HandleError(acceptingError, nil)

		waitGroup.Add(1)

		// when a new connection is established we are safe to create a new client instance
		var client = new(Client)
		client.Connection = newConnection
		client.Encoder = NewEncoder(newConnection)
		client.Decoder = NewDecoder(newConnection)
		// save the pointer to the client instance for later communication
		server.Clients.Append(client)

		Printf("[Log]: client at <Index: %d> joined the network\n", (server.Clients.Count() - 1))

		// handle the client on a different routine
		go server.ListenToClient(client, waitGroup)
	}

	// wait until every client told the clients its ID
	waitGroup.Wait()

	// stop listening for other connections
	listener.Close()

	Println("[Log]: calculating random graph")

	// create random graph
	var graph = CreateRandomGraph(maxClientNumber)
	LogGraph(&graph)

	for _, edge := range graph {

		var client_1 = server.Clients.ElementAtIndex(int(edge[0])).(*Client)
		var client_2 = server.Clients.ElementAtIndex(int(edge[1])).(*Client)
		server.MessagePipe <- Message{"server", client_1.Identification.ID, NewNeighborCommand, client_2.Identification}
	}

	time.Sleep(time.Second * 5)

	for i := 0; i < server.Clients.Count(); i++ {

		var client = server.Clients.ElementAtIndex(i).(*Client)
		server.MessagePipe <- Message{"server", client.Identification.ID, StopListeningCommand, nil}
	}

	time.Sleep(time.Second * 5)

	// send init message to a random node
	var startClient = server.Clients.ElementAtIndex(int(graph[0][0])).(*Client)
	server.MessagePipe <- Message{"server", startClient.Identification.ID, InitCommand, nil}

	// wait until the algorithm is done and a complete message
	// is recieved from a different go routine
	<-server.Complete
	Println("[Log]: server will terminate without errors")
	os.Exit(0)
}

func (server *Server) HandleMessages() {

	for {

		var message = <-server.MessagePipe

		Println("\n[Log] [Go]: new message")
		Printf("\tSender:   %s\n\tReceiver: %s\n\tCommand:  %s\n\tValue:\t  %v\n", message.Sender, message.Receiver, StringFor(message.Command), message.Value)

		if EqualStrings(message.Receiver, "server") {

			Println("[Log] [Go]: server is happy about completion")

			var finalStep = func() {

				for i := 0; i < server.Clients.Count(); i++ {

					var client = server.Clients.ElementAtIndex(i).(*Client)
					server.MessagePipe <- Message{"server", client.Identification.ID, FinalCommand, nil}
				}
				// give the message handling routine time to send
				time.Sleep(5 * time.Second)

				server.Complete <- true
			}
			go finalStep()

		} else {

			for i := 0; i < server.Clients.Count(); i++ {

				var client = server.Clients.ElementAtIndex(i).(*Client)

				if EqualStrings(client.Identification.ID, message.Receiver) {

					Printf("[Log] [Go]: server will send message to client <ID: %s>\n\n", client.Identification.ID)

					var encodingError = client.Encoder.Encode(message)
					HandleError(encodingError, func() {

						server.RemoveClient(client)
						Println(encodingError)
						os.Exit(10)
					})

					break
				}
			}
		}
	}
}

func (server *Server) ListenToClient(client *Client, waitGroup *sync.WaitGroup) {

	var decodingError = client.Decoder.Decode(&client.Identification)
	HandleError(decodingError, nil)

	waitGroup.Done()

	var clientIndex = server.Clients.Count() - 1

	Printf("[Log] [Go]: client at <Index: %s> is <ID: %s>\n", strconv.Itoa(clientIndex), client.Identification.ID)

	for run := true; run; {

		var message Message
		var decodingError = client.Decoder.Decode(&message)
		HandleError(decodingError, func() {

			Printf("[Log] [Go]: lost connection to client <ID: %s> at <Index: %s> \n", client.Identification.ID, strconv.Itoa(clientIndex))
			server.RemoveClient(client)
			run = false
		})

		if run == false {
			break
		}

		server.MessagePipe <- message
	}
}

func (server *Server) RemoveClient(client *Client) {

	client.Connection.Close()

	for i := 0; i < server.Clients.Count(); i++ {

		var aClient = server.Clients.ElementAtIndex(i).(*Client)
		if aClient == client {

			server.Clients.Remove(aClient)
			Println("[Log] [Go]: client was removed")
			break
		}
	}
}
