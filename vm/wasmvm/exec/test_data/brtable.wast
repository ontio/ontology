(module
  (type (;0;) (func (param i32) (result i32)))
  (type (;1;) (func (result i32)))
  (func (;0;) (type 0) (param i32) (result i32)
    block  ;; label = @1
      block  ;; label = @2
        block  ;; label = @3
          block  ;; label = @4
            get_local 0
            br_table 0 (;@4;) 1 (;@3;) 2 (;@2;) 3 (;@1;)
          end
          i32.const 0
          return
        end
        i32.const 1
        return
      end
    end
    i32.const 2
    return)
  (func (;1;) (type 1) (result i32)
    i32.const 0
    call 0)
  (func (;2;) (type 1) (result i32)
    i32.const 1
    call 0)
  (func (;3;) (type 1) (result i32)
    i32.const 2
    call 0)
  (func (;4;) (type 1) (result i32)
    i32.const 3
    call 0)
  (export "test0" (func 1))
  (export "test1" (func 2))
  (export "test2" (func 3))
  (export "test3" (func 4)))
