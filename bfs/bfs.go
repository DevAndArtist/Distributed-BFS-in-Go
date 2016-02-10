//
//  bfs.go
//
//  Created by Adrian Zubarev.
//  Copyright © 2016 Adrian Zubarev.
//  All rights reserved.
//

package bfs

import . "fmt"
import . "./command"
import . "./../array"

import "sync"
import "strings"

type Host interface {
	SendMessage(sender string, receiver string, command uint8, value interface{})
}

type Node struct {
	once       sync.Once
	guard      sync.Mutex
	host       Host
	id         string
	parentID   string
	treeLevel  int64
	labeled    bool
	neighbors  *Array
	sendTo     *Array
	children   *Array
	echoedFrom map[string]bool
}

func NodeWith(host Host, id string, neighbors []string) *Node {

	var node = new(Node)
	node.Set(host, id, neighbors)
	return node
}

func (node *Node) Set(host Host, id string, neighbors []string) *Node {

	node.guard.Lock()
	var onceBody = func() {

		node.host = host
		node.id = id
		node.parentID = ""
		node.treeLevel = -1
		node.labeled = false
		node.neighbors = ArrayOfType("string")

		for _, neighborID := range neighbors {

			node.neighbors.AppendUnique(neighborID)
		}

		node.sendTo = ArrayOfType("string")
		node.children = ArrayOfType("string")
		node.echoedFrom = make(map[string]bool)
	}
	node.once.Do(onceBody)
	node.guard.Unlock()

	return node
}

func (node *Node) IsRoot() bool {

	return strings.Compare(node.parentID, node.id) == 0
}

func (node *Node) HandleMessage(sender string, receiver string, command uint8, value interface{}) {

	node.guard.Lock()
	switch command {

	case InitCommand:
		node.labeled = true
		node.parentID = node.id
		node.treeLevel = 0
		node.sendTo = node.neighbors.Clone()
		node.children.RemoveAll() // making array empty

		if node.sendTo.IsEmpty() {

			node.host.SendMessage(node.id, "server", CompleteCommand, nil)

		} else {

			for i := 0; i < node.sendTo.Count(); i++ {

				var id = node.sendTo.ElementAtIndex(i).(string)
				node.echoedFrom[id] = false
				node.host.SendMessage(node.id, id, LabelCommand, node.treeLevel)
			}
		}

	case LabelCommand:
		if node.labeled == false {

			node.labeled = true
			node.parentID = sender
			node.treeLevel = value.(int64) + 1

			node.sendTo = node.neighbors.Clone()
			node.sendTo.Remove(sender)
			node.children.RemoveAll()

			if node.sendTo.IsEmpty() {

				node.host.SendMessage(node.id, node.parentID, EndCommand, nil)

			} else {

				node.host.SendMessage(node.id, node.parentID, KeeponCommand, nil)
			}

		} else {

			if strings.Compare(node.parentID, sender) == 0 {

				for i := 0; i < node.sendTo.Count(); i++ {

					var id = node.sendTo.ElementAtIndex(i).(string)
					node.echoedFrom[id] = false
					node.host.SendMessage(node.id, id, LabelCommand, node.treeLevel)
				}
			} else {

				node.host.SendMessage(node.id, sender, StopCommand, nil)
			}
		}

	case KeeponCommand, StopCommand, EndCommand:

		node.echoedFrom[sender] = true

		switch command {

		case KeeponCommand:
			node.children.AppendUnique(sender)

		case StopCommand:
			node.sendTo.Remove(sender)

		case EndCommand:
			node.children.AppendUnique(sender)
			node.sendTo.Remove(sender)
		}

		if node.sendTo.IsEmpty() {

			if node.IsRoot() {
				node.host.SendMessage(node.id, "server", CompleteCommand, nil)
			} else {
				node.host.SendMessage(node.id, node.parentID, EndCommand, nil)
			}
		} else {

			var everyNodeEchoed = true

			for i := 0; i < node.sendTo.Count(); i++ {

				var id = node.sendTo.ElementAtIndex(i).(string)
				if node.echoedFrom[id] == false {

					everyNodeEchoed = false
					break // found a node that not yet echoed
				}
			}

			if everyNodeEchoed {

				if node.IsRoot() {

					for i := 0; i < node.sendTo.Count(); i++ {

						var id = node.sendTo.ElementAtIndex(i).(string)

						node.echoedFrom[id] = false

						node.host.SendMessage(node.id, id, LabelCommand, node.treeLevel)
					}
				} else {
					node.host.SendMessage(node.id, node.parentID, KeeponCommand, nil)
				}
			}
		}
	default:
		Printf("[BFS Algorithm]: Unknown command \"%d\"- do nothing\n", command)
	}
	node.guard.Unlock()
}

func (node *Node) Children() []string {

	node.guard.Lock()
	var children []string
	for i := 0; i < node.children.Count(); i++ {

		var childID = node.children.ElementAtIndex(i).(string)
		children = append(children, childID)
	}
	node.guard.Unlock()
	return children
}
