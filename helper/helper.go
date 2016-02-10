//
//  helper.go
//
//  Created by Adrian Zubarev.
//  Copyright © 2016 Adrian Zubarev.
//  All rights reserved.
//

package helper

import . "fmt"

import "strings"
import "os/exec"

func HandleError(error error, handler func()) {

	if error != nil {

		if handler == nil {

			Println(error)

		} else {

			handler()
		}
	}
}

func EqualStrings(a, b string) bool {

	return strings.Compare(a, b) == 0
}

func GenerateID() string {

	var output, outputError = exec.Command("uuidgen").Output()

	HandleError(outputError, nil)

	return strings.TrimSuffix(string(output), "\n")
}
