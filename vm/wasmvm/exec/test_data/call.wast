(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32 i64 f32 f64) (result i32)))
  (type (;2;) (func (param i32) (result i32)))
  (func (;0;) (type 0) (result i32)
    i32.const 1
    i64.const 2
    f32.const 0x1.8p+1 (;=3;)
    f64.const 0x1p+2 (;=4;)
    call 1)
  (func (;1;) (type 1) (param i32 i64 f32 f64) (result i32)
    get_local 1
    i32.wrap/i64
    get_local 0
    i32.add
    get_local 2
    i32.trunc_s/f32
    i32.add
    get_local 3
    i32.trunc_s/f64
    i32.add
    return)
  (func (;2;) (type 0) (result i32)
    i32.const 10
    call 3)
  (func (;3;) (type 2) (param i32) (result i32)
    get_local 0
    i32.const 0
    i32.gt_s
    if (result i32)  ;; label = @1
      get_local 0
      get_local 0
      i32.const 1
      i32.sub
      call 3
      i32.mul
      return
    else
      i32.const 1
      return
    end)
  (export "call" (func 0))
  (export "fac10" (func 2)))
