(module
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (import "env" "calloc" (func $calloc (param i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "combine" (func $combine))
 (func $combine (; 1 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (i64.store align=4
   (i32.add
    (tee_local $2
     (call $calloc
      (i32.const 8)
      (i32.const 4)
     )
    )
    (i32.const 8)
   )
   (i64.load align=4
    (i32.add
     (get_local $0)
     (i32.const 8)
    )
   )
  )
  (i64.store align=4
   (get_local $2)
   (i64.load align=4
    (get_local $0)
   )
  )
  (i64.store align=4
   (i32.add
    (get_local $2)
    (i32.const 24)
   )
   (i64.load align=4
    (i32.add
     (get_local $1)
     (i32.const 8)
    )
   )
  )
  (i64.store align=4
   (i32.add
    (get_local $2)
    (i32.const 16)
   )
   (i64.load align=4
    (get_local $1)
   )
  )
  (get_local $2)
 )
)
