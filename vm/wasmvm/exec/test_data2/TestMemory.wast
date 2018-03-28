(module
  (type (;0;) (func (param i32 i32) (result i32)))
  (type (;1;) (func (result i32)))
  (import "env" "getString" (func (;0;) (type 0)))
  (import "env" "memory" (memory 1))
  (data (i32.const 0) "Hello world")
  (func (;1;) (type 1) (result i32)
    i32.const 0
    i32.const 11
    call 0)
  (export "getStrLen" (func 1)))
