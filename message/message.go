//
//  message.go
//
//  Created by Adrian Zubarev.
//  Copyright © 2016 Adrian Zubarev.
//  All rights reserved.
//

package message

type Message struct {
	Sender   string
	Receiver string
	Command  uint8
	Value    interface{}
}
