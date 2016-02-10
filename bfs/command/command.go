//
//  command.go
//
//  Created by Adrian Zubarev.
//  Copyright © 2016 Adrian Zubarev.
//  All rights reserved.
//

package command

const /* Message Command constants */ (
	NewNeighborCommand   uint8 = iota
	StopListeningCommand uint8 = iota
	InitCommand          uint8 = iota
	LabelCommand         uint8 = iota
	EndCommand           uint8 = iota
	KeeponCommand        uint8 = iota
	StopCommand          uint8 = iota
	CompleteCommand      uint8 = iota
	FinalCommand         uint8 = iota
)

func StringFor(command uint8) string {

	switch command {
	case NewNeighborCommand:
		return "New Neighbor"
	case StopListeningCommand:
		return "Stop Listening"
	case InitCommand:
		return "Init"
	case LabelCommand:
		return "Label"
	case EndCommand:
		return "End"
	case KeeponCommand:
		return "Keepon"
	case StopCommand:
		return "Stop"
	case CompleteCommand:
		return "Complete"
	case FinalCommand:
		return "Final"
	}
	return "Unknown Command"
}
