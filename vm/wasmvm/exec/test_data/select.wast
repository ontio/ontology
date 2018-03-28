(module
  (type (;0;) (func (param i32) (result i32)))
  (type (;1;) (func (param i32) (result i64)))
  (type (;2;) (func (param i32) (result f32)))
  (type (;3;) (func (param i32) (result f64)))
  (type (;4;) (func (result i32)))
  (type (;5;) (func (result i64)))
  (type (;6;) (func (result f32)))
  (type (;7;) (func (result f64)))
  (func (;0;) (type 0) (param i32) (result i32)
    i32.const 1
    i32.const 2
    get_local 0
    select)
  (func (;1;) (type 1) (param i32) (result i64)
    i64.const 1
    i64.const 2
    get_local 0
    select)
  (func (;2;) (type 2) (param i32) (result f32)
    f32.const 0x1p+0 (;=1;)
    f32.const 0x1p+1 (;=2;)
    get_local 0
    select)
  (func (;3;) (type 3) (param i32) (result f64)
    f64.const 0x1p+0 (;=1;)
    f64.const 0x1p+1 (;=2;)
    get_local 0
    select)
  (func (;4;) (type 4) (result i32)
    i32.const 0
    call 0)
  (func (;5;) (type 4) (result i32)
    i32.const 1
    call 0)
  (func (;6;) (type 5) (result i64)
    i32.const 0
    call 1)
  (func (;7;) (type 5) (result i64)
    i32.const 1
    call 1)
  (func (;8;) (type 6) (result f32)
    i32.const 0
    call 2)
  (func (;9;) (type 6) (result f32)
    i32.const 1
    call 2)
  (func (;10;) (type 7) (result f64)
    i32.const 0
    call 3)
  (func (;11;) (type 7) (result f64)
    i32.const 1
    call 3)
  (export "test_i32_l" (func 4))
  (export "test_i32_r" (func 5))
  (export "test_i64_l" (func 6))
  (export "test_i64_r" (func 7))
  (export "test_f32_l" (func 8))
  (export "test_f32_r" (func 9))
  (export "test_f64_l" (func 10))
  (export "test_f64_r" (func 11)))
