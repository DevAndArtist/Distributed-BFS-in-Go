//
//  client.go
//
//  Created by Adrian Zubarev.
//  Copyright © 2016 Adrian Zubarev.
//  All rights reserved.
//

package main

import . "fmt"
import . "./bfs"
import . "./array"
import . "./helper"
import . "./message"
import . "encoding/gob"
import . "./bfs/command"
import . "./identification"

import "os"
import "net"
import "strings"

type Client struct {
	ID               string
	Listener         net.Listener
	ServerConnection net.Conn
	ServerEncoder    *Encoder
	Neighbors        *Array
	Node             *Node
	MessagePipe      chan Message
	Complete         chan bool
}

type Neighbor struct {
	ID         string
	Connection net.Conn
	Encoder    *Encoder
}

func init() {
	// register gob types
	Register(Identification{})

	// register neighbor type
	RegisterType(&Neighbor{})
}

func main() {

	Println("\nStarting client ...")

	var client = new(Client)

	client.ID = GenerateID()
	client.Neighbors = ArrayOfType("*Neighbor")
	client.MessagePipe = make(chan Message)
	client.Complete = make(chan bool)

	Printf("[Log]: clients <ID: %s>\n", client.ID)

	go client.HandleMessages()

	var listener, listenerError = net.Listen("tcp", "localhost:0")
	HandleError(listenerError, func() {

		Println(listenerError)
		os.Exit(1)
	})
	client.Listener = listener

	var listenForNewClients = func() {

		for {
			Println("[Log] [Go]: wait for neighbor client")
			// warte auf eingehende verbindung
			var clientConnection, connectionError = listener.Accept()
			// schaue ob ein spezieller fehler auftritt
			if connectionError != nil && strings.HasSuffix(connectionError.Error(), "use of closed network connection") {

				break // wir müssen nicht mehr zuhören
			}
			HandleError(connectionError, func() {

				Println(connectionError)
				os.Exit(10)
			})
			// erstelle eine neue instanz
			var neighbor = new(Neighbor)
			neighbor.Connection = clientConnection
			neighbor.Encoder = NewEncoder(clientConnection)
			// warte auf die übermittelte ID
			var tempDecoder = NewDecoder(clientConnection)
			var id string
			var decodingError = tempDecoder.Decode(&id)
			HandleError(decodingError, func() {

				Println(decodingError)
				os.Exit(20)
			})
			// speichere die übermittelte ID
			neighbor.ID = id
			Printf("[Log] [Go]: new client <%s> accepted\n", id)
			// füge den neuen nachbar in ein array
			client.Neighbors.Append(neighbor)
			// starte eine go routine um den client zu bearbeiten
			go client.ListenTo(&neighbor.Connection)
		}

		Println("[Log] [Go]: client stopped accepting new connections")
		Println("[Log] [Go]: safe to set node neighbors")

		var neighbors []string
		for i := 0; i < client.Neighbors.Count(); i++ {

			var neighbor = client.Neighbors.ElementAtIndex(i).(*Neighbor)
			neighbors = append(neighbors, neighbor.ID)
		}

		client.Node = NodeWith(client, client.ID, neighbors)
	}
	go listenForNewClients()

	//===========================================================================================
	//===========================================================================================
	//===========================================================================================
	Println("[Log]: client will dial the server")
	var connectionToServer, connectionError = net.Dial("tcp", "localhost:8081")
	HandleError(connectionError, func() {

		Println(connectionError)
		os.Exit(30)
	})
	client.ServerConnection = connectionToServer
	client.ServerEncoder = NewEncoder(connectionToServer)
	Println("[Log]: connection to server established")
	Println("[Log]: client will send its identification to the server")
	// sende die ID und Rückrufaddresse für clients an den server
	var identificationMessage = Identification{client.ID, listener.Addr().String()}
	var encodingError = client.ServerEncoder.Encode(identificationMessage)
	HandleError(encodingError, func() {

		Println(encodingError)
		os.Exit(40)
	})
	go client.ListenTo(&client.ServerConnection)

	// warte bis der client mit allem fertig ist
	<-client.Complete

	os.Exit(0)
}

func (client *Client) ListenTo(connection *net.Conn) {

	var decoder = NewDecoder(*connection)

	for {
		Println("[Log] [GO]: wait for incoming messages")
		var message Message
		var decodingError = decoder.Decode(&message)
		HandleError(decodingError, func() {

			Println(decodingError)
			os.Exit(50)
		})
		client.MessagePipe <- message
	}
}

func (client *Client) HandleMessages() {

	for {

		var message = <-client.MessagePipe

		Println("\n[Log] [Async]: new message")
		Printf("\tSender:   %s\n\tReceiver: %s\n\tCommand:  %s\n\tValue:\t  %v\n\n", message.Sender, message.Receiver, StringFor(message.Command), message.Value)

		if EqualStrings(message.Sender, "server") {

			switch message.Command {

			case NewNeighborCommand:
				var identification = message.Value.(Identification)
				go client.DialNeighbor(identification.ID, "tcp", identification.Address)

			case StopListeningCommand:
				client.Listener.Close()

			case InitCommand:
				go client.Node.HandleMessage(message.Sender, message.Receiver, message.Command, message.Value)

			case FinalCommand:

				var children = client.Node.Children()

				if len(children) > 0 {
					Printf("[Log] [Tree] [Code-Part]:\n")

					for _, childID := range children {

						Printf("%s -> %s, \n", client.ID, childID)
					}
				}
				// possible TODO - create an array of Neighbors from given children

				client.Complete <- true

			default:
				Println("[Log]: unknown command from server")
				os.Exit(110)
			}

		} else {

			if EqualStrings(message.Receiver, "server") {

				// this can only be the complete command in our application
				var encodingError = client.ServerEncoder.Encode(message)
				HandleError(encodingError, func() {

					Println(encodingError)
					os.Exit(120)
				})
				Println("[Log] [Go]: message send to server")

			} else if EqualStrings(message.Receiver, client.ID) {

				go client.Node.HandleMessage(message.Sender, message.Receiver, message.Command, message.Value)

			} else {

				var neighbor *Neighbor
				for i := 0; i < client.Neighbors.Count(); i++ {

					var aNeighbor = client.Neighbors.ElementAtIndex(i).(*Neighbor)

					if EqualStrings(aNeighbor.ID, message.Receiver) {

						neighbor = aNeighbor
						break // found the reciever
					}
				}

				if neighbor == nil {

					Println("[Log] [Go]: SOMETHING IS REALLY BROKEN :(")
					os.Exit(130)
				}

				var encodingError = neighbor.Encoder.Encode(message)
				HandleError(encodingError, func() {

					Println(encodingError)
					os.Exit(140)
				})
				Println("[Log] [Go]: message send to neighbor client")
			}
		}
	}
}

func (client *Client) SendMessage(sender string, receiver string, command uint8, value interface{}) {

	client.MessagePipe <- Message{sender, receiver, command, value}
}

func (client *Client) DialNeighbor(id string, network string, address string) {

	Printf("[Log] [Go]: client will dial another client <ID: %s Address: %s>\n", id, address)
	var connection, connectionError = net.Dial(network, address)
	HandleError(connectionError, func() {

		Println(connectionError)
		os.Exit(60)
	})
	Println("[Log] [Go]: successfully connected")

	var neighbor = new(Neighbor)
	neighbor.ID = id
	neighbor.Connection = connection
	neighbor.Encoder = NewEncoder(connection)

	Println("[Log] [Go]: send own ID to the connected client")
	var encodingError = neighbor.Encoder.Encode(client.ID)
	HandleError(encodingError, func() {

		Println(encodingError)
		os.Exit(70)
	})
	client.Neighbors.Append(neighbor)

	go client.ListenTo(&neighbor.Connection)
}
