(module
  (type (;0;) (func (param i32)))
  (type (;1;) (func (param i32) (result i32)))
  (type (;2;) (func (param i32 i32 i32) (result i32)))
  (type (;3;) (func (param i32 i32) (result i32)))
  (type (;4;) (func (result i32)))
  (type (;5;) (func))
  (import "env" "memory" (memory (;0;) 1))
  (import "env" "memoryBase" (global (;0;) i32))
  (import "env" "JsonMashalParams" (func (;0;) (type 0)))
  (import "env" "arrayLen" (func (;1;) (type 1)))
  (import "env" "malloc" (func (;2;) (type 1)))
  (import "env" "memcpy" (func (;3;) (type 2)))
  (func (;4;) (type 3) (param i32 i32) (result i32)
    (local i32 i32 i32)
    get_local 0
    call 1
    set_local 2
    get_local 1
    call 1
    tee_local 3
    get_local 2
    i32.add
    call 2
    set_local 4
    get_local 2
    i32.const 0
    i32.gt_s
    if  ;; label = @1
      get_local 4
      get_local 0
      get_local 2
      call 3
      drop
    end
    get_local 3
    i32.const 0
    i32.gt_s
    if  ;; label = @1
      get_local 4
      get_local 3
      i32.add
      get_local 1
      get_local 3
      call 3
      drop
    end
    get_local 4)
  (func (;5;) (type 4) (result i32)
    (local i32 i32 i32 i32 i32)
    i32.const 40
    call 2
    set_local 2
    loop  ;; label = @1
      i32.const 4
      call 2
      tee_local 1
      get_global 0
      i32.store
      get_local 1
      get_global 0
      i32.const 7
      i32.add
      i32.store offset=4
      get_local 1
      i32.load offset=4
      set_local 3
      get_local 2
      get_local 0
      i32.const 3
      i32.shl
      i32.add
      tee_local 4
      get_local 1
      i32.load
      i32.store
      get_local 4
      get_local 3
      i32.store offset=4
      get_local 0
      i32.const 1
      i32.add
      tee_local 0
      i32.const 10
      i32.ne
      br_if 0 (;@1;)
    end
    get_local 2
    call 0
    i32.const 0)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (export "invoke" (func 5))
  (data (get_global 0) "string\00abcdefg"))
