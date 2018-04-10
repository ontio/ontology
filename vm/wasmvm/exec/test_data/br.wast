(module
  (type (;0;) (func (result i32)))
  (func (;0;) (type 0) (result i32)
    (local i32 i32)
    block  ;; label = @1
      i32.const 1
      if  ;; label = @2
        br 1 (;@1;)
      end
      i32.const 1
      set_local 0
    end
    i32.const 1
    set_local 1
    get_local 0
    i32.const 0
    i32.eq
    get_local 1
    i32.const 1
    i32.eq
    i32.add
    return)
  (func (;1;) (type 0) (result i32)
    (local i32 i32 i32)
    block  ;; label = @1
      block  ;; label = @2
        i32.const 1
        if  ;; label = @3
          br 2 (;@1;)
        end
        i32.const 1
        set_local 0
      end
      i32.const 1
      set_local 1
    end
    i32.const 1
    set_local 2
    get_local 0
    i32.const 0
    i32.eq
    get_local 1
    i32.const 0
    i32.eq
    i32.add
    get_local 2
    i32.const 1
    i32.eq
    i32.add
    return)
  (func (;2;) (type 0) (result i32)
    block  ;; label = @1
      block  ;; label = @2
        i32.const 1
        if  ;; label = @3
          br 2 (;@1;)
        end
        i32.const 1
        return
      end
    end
    i32.const 2
    return)
  (func (;3;) (type 0) (result i32)
    (local i32 i32)
    block  ;; label = @1
      loop  ;; label = @2
        get_local 0
        i32.const 1
        i32.add
        set_local 0
        get_local 0
        i32.const 5
        i32.ge_s
        if  ;; label = @3
          br 2 (;@1;)
        end
        get_local 0
        i32.const 4
        i32.eq
        if  ;; label = @3
          br 1 (;@2;)
        end
        get_local 0
        set_local 1
        br 0 (;@2;)
      end
    end
    get_local 1
    return)
  (export "br0" (func 0))
  (export "br1" (func 1))
  (export "br2" (func 2))
  (export "br3" (func 3)))
