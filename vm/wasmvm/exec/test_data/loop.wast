(module
  (type (;0;) (func (result i32)))
  (func (;0;) (type 0) (result i32)
    (local i32 i32)
    loop  ;; label = @1
      get_local 1
      get_local 0
      i32.add
      set_local 1
      get_local 0
      i32.const 1
      i32.add
      set_local 0
      get_local 0
      i32.const 5
      i32.lt_s
      if  ;; label = @2
        br 1 (;@1;)
      end
    end
    get_local 1)
  (export "loop" (func 0)))
