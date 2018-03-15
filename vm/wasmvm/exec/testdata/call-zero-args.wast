(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (param i32 i32) (result i32)))
  (func (;0;) (type 0) (result i32)
    i32.const 42)
  (func (;1;) (type 1) (param i32 i32) (result i32)
    get_local 0
    get_local 1
    i32.add)
  (func (;2;) (type 0) (result i32)
    i32.const 1
    call 0
    call 1)
  (export "h" (func 2)))
