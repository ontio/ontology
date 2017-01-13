/*
go vm External API Interface
 */
package vm

type IApiService interface {
	Invoke(method string, engine *ExecutionEngine) (bool, error)
}
