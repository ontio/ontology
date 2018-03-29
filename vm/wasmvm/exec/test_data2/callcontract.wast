(module
  (type (;0;) (func (param i32 i32 i32) (result i32)))
  (type (;1;) (func (param i32 i32) (result i32)))
  (type (;2;) (func))
  (import "env" "memory" (memory (;0;) 1))
  (import "env" "memoryBase" (global (;0;) i32))
  (import "env" "callContract" (func (;0;) (type 0)))
  (import "env" "strcmp" (func (;1;) (type 1)))
  (func (;2;) (type 1) (param i32 i32) (result i32)
    get_local 1
    get_local 0
    i32.add)
  (func (;3;) (type 1) (param i32 i32) (result i32)
    get_local 0
    get_global 0
    call 1
    if (result i32)  ;; label = @1
      get_local 0
      get_global 0
      i32.const 19
      i32.add
      call 1
      if (result i32)  ;; label = @2
        i32.const 0
      else
        get_global 0
        i32.const 23
        i32.add
        get_global 0
        i32.const 19
        i32.add
        get_local 1
        call 0
      end
    else
      get_global 0
      i32.const 5
      i32.add
    end)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (export "invoke" (func 3))
  (data (get_global 0) "init\00init success!\00add\00rawcontract"))
