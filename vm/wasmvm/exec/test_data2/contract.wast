(module
  (type (;0;) (func (param i32)))
  (type (;1;) (func (param i32) (result i32)))
  (type (;2;) (func (param i32 i32) (result i32)))
  (type (;3;) (func (param i32 i32 i32)))
  (type (;4;) (func (param i32 i32)))
  (type (;5;) (func (param i32 i32 i32) (result i32)))
  (type (;6;) (func))
  (import "env" "memory" (memory (;0;) 1))
  (import "env" "memoryBase" (global (;0;) i32))
  (import "env" "DeleteStorage" (func (;0;) (type 0)))
  (import "env" "GetStorage" (func (;1;) (type 1)))
  (import "env" "JsonMashalResult" (func (;2;) (type 2)))
  (import "env" "JsonUnmashalInput" (func (;3;) (type 3)))
  (import "env" "PutStorage" (func (;4;) (type 4)))
  (import "env" "RuntimeNotify" (func (;5;) (type 0)))
  (import "env" "arrayLen" (func (;6;) (type 1)))
  (import "env" "malloc" (func (;7;) (type 1)))
  (import "env" "memcpy" (func (;8;) (type 5)))
  (import "env" "strcmp" (func (;9;) (type 2)))
  (func (;10;) (type 2) (param i32 i32) (result i32)
    get_local 1
    get_local 0
    i32.add)
  (func (;11;) (type 2) (param i32 i32) (result i32)
    (local i32 i32 i32)
    get_local 0
    call 6
    set_local 2
    get_local 1
    call 6
    tee_local 3
    get_local 2
    i32.add
    call 7
    set_local 4
    get_local 2
    i32.const 0
    i32.gt_s
    if  ;; label = @1
      get_local 4
      get_local 0
      get_local 2
      call 8
      drop
    end
    get_local 3
    i32.const 0
    i32.gt_s
    if  ;; label = @1
      get_local 4
      get_local 3
      i32.add
      get_local 1
      get_local 3
      call 8
      drop
    end
    get_local 4)
  (func (;12;) (type 2) (param i32 i32) (result i32)
    (local i32 i32 i32 i32)
    get_local 0
    call 6
    set_local 4
    get_local 1
    call 6
    set_local 5
    get_local 4
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
        get_local 4
        i32.ne
        br_if 0 (;@2;)
        get_local 2
        set_local 0
      end
    else
      i32.const 0
      set_local 0
    end
    get_local 5
    i32.const 0
    i32.gt_s
    if  ;; label = @1
      i32.const 0
      set_local 2
      loop  ;; label = @2
        get_local 1
        get_local 2
        i32.const 2
        i32.shl
        i32.add
        i32.load
        get_local 0
        i32.add
        set_local 0
        get_local 2
        i32.const 1
        i32.add
        tee_local 2
        get_local 5
        i32.ne
        br_if 0 (;@2;)
      end
    end
    get_local 0)
  (func (;13;) (type 2) (param i32 i32) (result i32)
    (local i32 i32)
    get_global 1
    set_local 3
    get_global 1
    i32.const 16
    i32.add
    set_global 1
    get_local 3
    set_local 2
    block  ;; label = @1
      get_local 0
      get_global 0
      call 9
      if  ;; label = @2
        get_local 0
        get_global 0
        i32.const 19
        i32.add
        call 9
        i32.eqz
        if  ;; label = @3
          get_local 2
          i32.const 8
          get_local 1
          call 3
          get_local 2
          i32.load
          get_local 2
          i32.load offset=4
          call 10
          get_global 0
          i32.const 23
          i32.add
          call 2
          tee_local 0
          call 5
          br 2 (;@1;)
        end
        get_local 0
        get_global 0
        i32.const 27
        i32.add
        call 9
        i32.eqz
        if  ;; label = @3
          get_local 2
          i32.const 8
          get_local 1
          call 3
          get_local 2
          i32.load
          get_local 2
          i32.load offset=4
          call 11
          get_global 0
          i32.const 34
          i32.add
          call 2
          tee_local 0
          call 5
          br 2 (;@1;)
        end
        get_local 0
        get_global 0
        i32.const 41
        i32.add
        call 9
        i32.eqz
        if  ;; label = @3
          get_local 2
          i32.const 8
          get_local 1
          call 3
          get_local 2
          i32.load
          get_local 2
          i32.load offset=4
          call 12
          get_global 0
          i32.const 23
          i32.add
          call 2
          tee_local 0
          call 5
          br 2 (;@1;)
        end
        get_local 0
        get_global 0
        i32.const 50
        i32.add
        call 9
        i32.eqz
        if  ;; label = @3
          get_local 2
          i32.const 8
          get_local 1
          call 3
          get_local 2
          i32.load
          get_local 2
          i32.load offset=4
          call 4
          get_global 0
          i32.const 61
          i32.add
          get_global 0
          i32.const 34
          i32.add
          call 2
          tee_local 0
          call 5
          br 2 (;@1;)
        end
        get_local 0
        get_global 0
        i32.const 66
        i32.add
        call 9
        i32.eqz
        if  ;; label = @3
          get_local 2
          i32.const 4
          get_local 1
          call 3
          get_local 2
          i32.load
          call 1
          get_global 0
          i32.const 34
          i32.add
          call 2
          tee_local 0
          call 5
          br 2 (;@1;)
        end
        get_local 0
        get_global 0
        i32.const 77
        i32.add
        call 9
        if  ;; label = @3
          i32.const 0
          set_local 0
        else
          get_local 2
          i32.const 4
          get_local 1
          call 3
          get_local 2
          i32.load
          call 0
          get_global 0
          i32.const 61
          i32.add
          get_global 0
          i32.const 34
          i32.add
          call 2
          tee_local 0
          call 5
        end
      else
        get_global 0
        i32.const 5
        i32.add
        set_local 0
      end
    end
    get_local 3
    set_global 1
    get_local 0)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (export "invoke" (func 13))
  (data (get_global 0) "init\00init success!\00add\00int\00concat\00string\00sumArray\00addStorage\00Done\00getStorage\00deleteStorage"))
