package interfaces

type IScriptTable interface {
	GetScript(script_hash []byte) ([]byte)
}
