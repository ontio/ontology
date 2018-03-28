(module
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (import "env" "calloc" (func $calloc (param i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "retArray" (func $retArray))
 (func $retArray (; 1 ;) (result i32)
  (local $0 i32)
  (i64.store offset=4 align=4
   (tee_local $0
    (call $calloc
     (i32.const 10)
     (i32.const 4)
    )
   )
   (i64.const 8589934593)
  )
  (i64.store offset=12 align=4
   (get_local $0)
   (i64.const 17179869187)
  )
  (i64.store offset=20 align=4
   (get_local $0)
   (i64.const 25769803781)
  )
  (i64.store offset=28 align=4
   (get_local $0)
   (i64.const 34359738375)
  )
  (i32.store offset=36
   (get_local $0)
   (i32.const 9)
  )
  (get_local $0)
 )
)
