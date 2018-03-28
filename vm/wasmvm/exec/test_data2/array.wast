(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func))
  (import "env" "memoryBase" (global (;0;) i32))
  (func (;0;) (type 0) (result i32)
    (local i32)
    get_global 1
    set_local 0
    get_global 1
    i32.const 32
    i32.add
    set_global 1
    get_local 0
    set_global 1
    get_local 0)
  (func (;1;) (type 1)
    nop)
  (func (;2;) (type 1)
    get_global 0
    set_global 1
    get_global 1
    i32.const 5242880
    i32.add
    set_global 2)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (export "__post_instantiate" (func 2))
  (export "_getArray" (func 0))
  (export "runPostSets" (func 1)))
