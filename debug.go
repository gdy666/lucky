//go:build debug
// +build debug

package main

import (
	"fmt"

	"github.com/gdy666/lucky/thirdlib/gdylib/recoverutil"
)

func init() {
	defer func() {
		recoverErr := recover()
		if recoverErr == nil {
			return
		}
		panicFile := fmt.Sprintf("闪退.log")
		recoverutil.RecoverHandler(recoverErr, true, true, panicFile)
	}()
}
