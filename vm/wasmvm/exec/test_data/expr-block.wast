(module
  (type (;0;) (func (result i32)))
  (func (;0;) (type 0) (result i32)
    block (result i32)  ;; label = @1
      i32.const 10
      drop
      i32.const 1
    end)
  (export "test" (func 0)))
