(module
  (type (;0;) (func (param i32 i32) (result i32)))
  (type (;1;) (func (param i32 i32)))
  (type (;2;) (func))
  (import "env" "memoryBase" (global (;0;) i32))
  (import "env" "read_message" (func (;0;) (type 0)))
  (func (;1;) (type 1) (param i32 i32)
    (local i32)
    get_global 1
    set_local 0
    get_global 1
    i32.const 16
    i32.add
    set_global 1
    get_local 0
    set_local 2
    get_local 1
    i32.eqz
    if  ;; label = @1
      get_local 2
      i32.const 12
      call 0
      drop
    end
    get_local 0
    set_global 1)
  (func (;2;) (type 2)
    nop)
  (func (;3;) (type 2)
    get_global 0
    set_global 1
    get_global 1
    i32.const 5242880
    i32.add
    set_global 2)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (memory $0 1)
  (export "__post_instantiate" (func 3))
  (export "apply" (func 1))
  (export "runPostSets" (func 2)))
