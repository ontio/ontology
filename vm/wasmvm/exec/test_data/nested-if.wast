(module
  (type (;0;) (func (result i32)))
  (func (;0;) (type 0) (result i32)
    block  ;; label = @1
      i32.const 1
      if  ;; label = @2
        i32.const 2
        drop
        i32.const 3
        if  ;; label = @3
          br 2 (;@1;)
        end
      end
    end
    i32.const 4)
  (export "f" (func 0)))
