(module
  (type (;0;) (func (param i32 i32 i32) (result i32)))
  (type (;1;) (func (param i32) (result i32)))
  (type (;2;) (func (param i32)))
  (type (;3;) (func (param i32 i32) (result i32)))
  (type (;4;) (func))
  (import "env" "memory" (memory (;0;) 1))
  (import "env" "memoryBase" (global (;0;) i32))
  (import "env" "CallContract" (func (;0;) (type 0)))
  (import "env" "RawMashalParams" (func (;1;) (type 1)))
  (import "env" "ReadStringParam" (func (;2;) (type 1)))
  (import "env" "RuntimeNotify" (func (;3;) (type 2)))
  (import "env" "malloc" (func (;4;) (type 1)))
  (import "env" "strcmp" (func (;5;) (type 3)))
  (func (;6;) (type 3) (param i32 i32) (result i32)
    (local i32)
    block  ;; label = @1
      get_local 0
      get_global 0
      call 5
      if  ;; label = @2
        get_local 0
        get_global 0
        i32.const 19
        i32.add
        call 5
        i32.eqz
        if  ;; label = @3
          get_local 1
          call 2
          set_local 1
          i32.const 8
          call 4
          tee_local 0
          get_global 0
          i32.const 28
          i32.add
          i32.store
          get_local 0
          get_local 1
          i32.store offset=4
          get_local 0
          call 1
          set_local 0
          get_global 0
          i32.const 35
          i32.add
          get_global 0
          i32.const 76
          i32.add
          get_local 0
          call 0
          tee_local 0
          call 3
          br 2 (;@1;)
        end
        get_local 0
        get_global 0
        i32.const 87
        i32.add
        call 5
        if  ;; label = @3
          i32.const 0
          set_local 0
        else
          get_local 1
          call 2
          set_local 2
          get_local 1
          call 2
          set_local 1
          i32.const 16
          call 4
          tee_local 0
          get_global 0
          i32.const 28
          i32.add
          i32.store
          get_local 0
          get_local 2
          i32.store offset=4
          get_local 0
          get_global 0
          i32.const 28
          i32.add
          i32.store offset=8
          get_local 0
          get_local 1
          i32.store offset=12
          get_local 0
          call 1
          set_local 0
          get_global 0
          i32.const 35
          i32.add
          get_global 0
          i32.const 96
          i32.add
          get_local 0
          call 0
          tee_local 0
          call 3
        end
      else
        get_global 0
        i32.const 5
        i32.add
        set_local 0
      end
    end
    get_local 0)
  (global (;1;) (mut i32) (i32.const 0))
  (global (;2;) (mut i32) (i32.const 0))
  (export "invoke" (func 6))
  (data (get_global 0) "init\00init success!\00getValue\00string\0090ca57034b30a3fe0da266c20b337ed2a0aa47a6\00getStorage\00putValue\00addStorage"))
