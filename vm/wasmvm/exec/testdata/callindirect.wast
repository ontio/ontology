(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32 i32) (result i32)))
  (type (;2;) (func (param i32) (result i32)))
  (type (;3;) (func (param i32 i32 i32) (result i32)))
  (func (;0;) (type 0) (result i32)
    i32.const 0)
  (func (;1;) (type 0) (result i32)
    i32.const 1)
  (func (;2;) (type 2) (param i32) (result i32)
    get_local 0
    call_indirect (type 0))
  (func (;3;) (type 1) (param i32 i32) (result i32)
    get_local 0
    get_local 1
    i32.add)
  (func (;4;) (type 1) (param i32 i32) (result i32)
    get_local 0
    get_local 1
    i32.sub)
  (func (;5;) (type 3) (param i32 i32 i32) (result i32)
    get_local 0
    get_local 1
    get_local 2
    call_indirect (type 1))
  (func (;6;) (type 0) (result i32)
    i32.const 0
    call 2)
  (func (;7;) (type 0) (result i32)
    i32.const 1
    call 2)
  (func (;8;) (type 0) (result i32)
    i32.const 10
    i32.const 4
    i32.const 2
    call 5)
  (func (;9;) (type 0) (result i32)
    i32.const 10
    i32.const 4
    i32.const 3
    call 5)
  (func (;10;) (type 0) (result i32)
    i32.const 10
    i32.const 4
    i32.const 4
    call 5)
  (func (;11;) (type 0) (result i32)
    i32.const 10
    i32.const 4
    i32.const 0
    call 5)
  (table (;0;) 4 4 anyfunc)
  (export "test_zero" (func 6))
  (export "test_one" (func 7))
  (export "test_add" (func 8))
  (export "test_sub" (func 9))
  (export "trap_oob" (func 10))
  (export "trap_sig_mismatch" (func 11))
  (elem (i32.const 0) 0 1 3 4))
