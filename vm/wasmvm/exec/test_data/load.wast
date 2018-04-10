(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (result i64)))
  (type (;2;) (func (result f32)))
  (type (;3;) (func (result f64)))
  (func (;0;) (type 0) (result i32)
    i32.const 0
    i32.load8_s)
  (func (;1;) (type 0) (result i32)
    i32.const 0
    i32.load16_s)
  (func (;2;) (type 0) (result i32)
    i32.const 0
    i32.load)
  (func (;3;) (type 0) (result i32)
    i32.const 0
    i32.load8_u)
  (func (;4;) (type 0) (result i32)
    i32.const 0
    i32.load16_u)
  (func (;5;) (type 1) (result i64)
    i32.const 0
    i64.load8_s)
  (func (;6;) (type 1) (result i64)
    i32.const 0
    i64.load16_s)
  (func (;7;) (type 1) (result i64)
    i32.const 0
    i64.load32_s)
  (func (;8;) (type 1) (result i64)
    i32.const 16
    i64.load)
  (func (;9;) (type 1) (result i64)
    i32.const 0
    i64.load8_u)
  (func (;10;) (type 1) (result i64)
    i32.const 0
    i64.load16_u)
  (func (;11;) (type 1) (result i64)
    i32.const 0
    i64.load32_u)
  (func (;12;) (type 2) (result f32)
    i32.const 4
    f32.load)
  (func (;13;) (type 3) (result f64)
    i32.const 8
    f64.load)
  (memory (;0;) 1)
  (export "i32_load8_s" (func 0))
  (export "i32_load16_s" (func 1))
  (export "i32_load" (func 2))
  (export "i32_load8_u" (func 3))
  (export "i32_load16_u" (func 4))
  (export "i64_load8_s" (func 5))
  (export "i64_load16_s" (func 6))
  (export "i64_load32_s" (func 7))
  (export "i64_load" (func 8))
  (export "i64_load8_u" (func 9))
  (export "i64_load16_u" (func 10))
  (export "i64_load32_u" (func 11))
  (export "f32_load" (func 12))
  (export "f64_load" (func 13))
  (data (i32.const 0) "\ff\ff\ff\ff")
  (data (i32.const 4) "\00\00\ceA")
  (data (i32.const 8) "\00\00\00\00\00\ff\8f@")
  (data (i32.const 16) "\ff\ff\ff\ff\ff\ff\ff\ff"))
