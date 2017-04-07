package vm

type InteropService struct {
	dictionary map[string]func(*ExecutionEngine) bool
}

func NewInteropService() *InteropService {
	var is InteropService
	is.dictionary = make(map[string]func(*ExecutionEngine) bool, 0)
	is.Register("System.ScriptEngine.GetScriptContainer", is.GetScriptContainer)
	is.Register("System.ScriptEngine.GetExecutingScriptHash", is.GetExecutingScriptHash)
	is.Register("System.ScriptEngine.GetCallingScriptHash", is.GetCallingScriptHash)
	is.Register("System.ScriptEngine.GetEntryScriptHash", is.GetEntryScriptHash)
	return &is
}

func (is *InteropService) Register(method string, handler func(*ExecutionEngine) bool) bool {
	if _, ok := is.dictionary[method]; ok {
		return false
	}
	is.dictionary[method] = handler
	return true
}

func (is *InteropService) Invoke(method string, engine *ExecutionEngine) bool {
	if v, ok := is.dictionary[method]; ok {
		return v(engine)
	}
	return false
}

func (is *InteropService) GetScriptContainer(engine *ExecutionEngine) bool {
	engine.evaluationStack.Push(engine.scriptContainer)
	return true
}

func (is *InteropService) GetExecutingScriptHash(engine *ExecutionEngine) bool {
	engine.evaluationStack.Push(engine.crypto.Hash160(engine.ExecutingScript()))
	return true
}

func (is *InteropService) GetCallingScriptHash(engine *ExecutionEngine) bool {
	engine.evaluationStack.Push(engine.crypto.Hash160(engine.CallingScript()))
	return true
}

func (is *InteropService) GetEntryScriptHash(engine *ExecutionEngine) bool {
	engine.evaluationStack.Push(engine.crypto.Hash160(engine.EntryScript()))
	return true
}
