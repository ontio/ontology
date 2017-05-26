package log

import (
	"testing"
)

func TestDebugPrint(t *testing.T) {
	CreatePrintLog("./")
	Debug("debug testing")
}

func TestInfoPrint(t *testing.T) {
	CreatePrintLog("./")
	Info("Info testing")
}

func TestWarningPrint(t *testing.T) {
	CreatePrintLog("./")
	Warn("Warning testing")
}

func TestErrorPrint(t *testing.T) {
	CreatePrintLog("./")
	Error("Error testing")
}

func TestFatalPrint(t *testing.T) {
	CreatePrintLog("./")
	Fatal("Fatal testing")
}
