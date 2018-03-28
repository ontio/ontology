(module
  (type (;0;) (func (result i32)))
  (import "env" "memory" (memory (;0;) 256))
  (import "env" "memoryBase" (global (;0;) i32))
  (func (;0;) (type 0) (result i32)
      get_global 0)
  (export "_getStr" (func 0))
  (data (get_global 0) "hello world!"))
  ;;(data (i32.const 0) "hello world!"))
