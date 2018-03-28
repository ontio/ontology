(module
  (type (;0;) (func (result i32)))
  (func (;0;) (type 0) (result i32)
    (local i32)
    i32.const 0
    set_local 0
    i32.const 1
    if  ;; label = @1
      get_local 0
      i32.const 1
      i32.add
      set_local 0
    end
    i32.const 0
    if  ;; label = @1
      get_local 0
      i32.const 1
      i32.add
      set_local 0
    end
    get_local 0
    return)
  (func (;1;) (type 0) (result i32)
    (local i32 i32)
    i32.const 1
    if  ;; label = @1
      i32.const 1
      set_local 0
    else
      i32.const 2
      set_local 0
    end
    i32.const 0
    if  ;; label = @1
      i32.const 4
      set_local 1
    else
      i32.const 8
      set_local 1
    end
    get_local 0
    get_local 1
    i32.add
    return)
  (export "if1" (func 0))
  (export "if2" (func 1)))
