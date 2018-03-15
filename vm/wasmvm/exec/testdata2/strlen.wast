(module
  (type (;0;) (func (param i32) (result i32)))
  (import "env" "memoryBase" (global (;0;) i32))
  (import "env" "strlen" (func (;0;) (type 0)))
  (func (;1;) (type 0) (param i32) (result i32)
    get_local 0
    call 0)
  (export "stringlen" (func 1)))

