(module
  (type (;0;) (func (param i32) (result i32)))
  (type (;1;) (func (result i32)))
  (func (;0;) (type 0) (param i32) (result i32)
    get_local 0
    i32.const 0
    i32.eq
    if  ;; label = @1
      i32.const 1
      return
    end
    get_local 0
    i32.const 1
    i32.eq
    if  ;; label = @1
      i32.const 2
      return
    end
    i32.const 3
    return)
  (func (;1;) (type 1) (result i32)
    i32.const 0
    call 0)
  (func (;2;) (type 1) (result i32)
    i32.const 1
    call 0)
  (func (;3;) (type 1) (result i32)
    i32.const 5
    call 0)
  (export "test1" (func 1))
  (export "test2" (func 2))
  (export "test3" (func 3)))
