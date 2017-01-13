package vm

import (
)

type ScriptContext struct {
	Script []byte
	OpReader * VmReader
	//BreakPoints
}

func NewScriptContext(script []byte) *ScriptContext {
	var stackContext ScriptContext
	stackContext.Script = script
	stackContext.OpReader = NewVmReader( script )
	return &stackContext
}
/*
func (sc *ScriptContext) Dispose() {
	sc.OpReader.Dispose();

}
*/