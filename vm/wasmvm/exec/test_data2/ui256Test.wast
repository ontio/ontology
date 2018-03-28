(module
      (import "env" "memory" (memory (;0;) 1))
      (import "env" "memoryBase" (global (;0;) i32))
      (import "env" "getAcct" (func $getAcct (result i64)))
      (func (export "getAddress")(param i64 i64 i64 i64)(result i64)
           (i64.store (i32.const 0)(get_local 0))
           (i64.store (i32.const 8)(get_local 1))
           (i64.store (i32.const 16)(get_local 2))
           (i64.store (i32.const 24)(get_local 3))
           (call $getAcct)
      )

)