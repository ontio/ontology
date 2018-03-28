(module
  (type (;0;) (func (param i32 i32) (result f64)))
  (type (;1;) (func (param i32 i32) (result i32)))
  (type (;2;) (func))
  (import "env" "memory" (memory (;0;) 256))
  (import "env" "memoryBase" (global (;0;) i32))
  (func (;0;) (type 0) (param i32 i32) (result f64)
    (local i32 f64)
    get_local 1
    i32.const 0
    i32.gt_s
    if  ;; label = @1
      loop  ;; label = @2
        get_local 3
        get_local 0
        get_local 2
        i32.const 3
        i32.shl
        i32.add
        f64.load
        f64.add
        set_local 3
        get_local 2
        i32.const 1
        i32.add
        tee_local 2
        get_local 1
        i32.ne
        br_if 0 (;@2;)
      end
    end
    get_local 3)
  (func (;1;) (type 1) (param i32 i32) (result i32)
    (local i32 i32)
    get_local 1
    i32.const 0
    i32.gt_s
    if  ;; label = @1
      loop  ;; label = @2
        get_local 0
        get_local 3
        i32.const 2
        i32.shl
        i32.add
        i32.load
        get_local 2
        i32.add
        set_local 2
        get_local 3
        i32.const 1
        i32.add
        tee_local 3
        get_local 1
        i32.ne
        br_if 0 (;@2;)
      end
    end
    get_local 2)
  (func (;2;) (type 2)
    nop)
  (func (;3;) (type 2)
    get_global 0
    set_global 1
    get_global 1
    i32.const 5242880
    i32.add
    set_global 2)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (export "_DoubleSum" (func 0))
  (export "_IntSum" (func 1))
  (export "__post_instantiate" (func 3))
  (export "runPostSets" (func 2)))
