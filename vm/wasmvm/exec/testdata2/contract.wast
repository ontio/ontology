(module
  (type (;0;) (func (param i32 i32) (result i32)))
  (type (;1;) (func (param i32 i32 i32)))
  (type (;2;) (func))
  (import "env" "memory" (memory (;0;) 1))
  (import "env" "memoryBase" (global (;0;) i32))
  (import "env" "JsonMashal" (func (;0;) (type 0)))
  (import "env" "JsonUnmashal" (func (;1;) (type 1)))
  (import "env" "strcmp" (func (;2;) (type 0)))
  (func (;3;) (type 0) (param i32 i32) (result i32)
    get_local 1
    get_local 0
    i32.add)
  (func (;4;) (type 0) (param i32 i32) (result i32)
    (local i32 i32)
    get_global 1
    set_local 3
    get_global 1
    i32.const 16
    i32.add
    set_global 1
    get_local 3
    set_local 2
    get_local 0
    get_global 0
    call 2
    if (result i32)  ;; label = @1
      get_local 0
      get_global 0
      i32.const 19
      i32.add
      call 2
      if (result i32)  ;; label = @2
        i32.const 0
      else
        get_local 2
        i32.const 8
        get_local 1
        call 1
        get_local 2
        i32.load
        get_local 2
        i32.load offset=4
        call 3
        get_global 0
        i32.const 23
        i32.add
        call 0
      end
    else
      get_global 0
      i32.const 5
      i32.add
    end
    set_local 0
    get_local 3
    set_global 1
    get_local 0)
  (func (;5;) (type 2)
    nop)
  (func (;6;) (type 2)
    get_global 0
    i32.const 32
    i32.add
    set_global 1
    get_global 1
    i32.const 5242880
    i32.add
    set_global 2)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (export "invoke" (func 4))
  (data (get_global 0) "init\00init success!\00add\00int"))
