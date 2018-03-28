(module
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (import "env" "malloc" (func $malloc (param i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "initStu" (func $initStu))
 (func $initStu (; 1 ;) (param $0 i32) (param $1 i32) (param $2 i32) (result i32)
  (local $3 i32)
  (i32.store offset=4
   (tee_local $3
    (call $malloc
     (i32.const 12)
    )
   )
   (get_local $1)
  )
  (i32.store
   (get_local $3)
   (get_local $0)
  )
  (i32.store offset=8
   (get_local $3)
   (get_local $2)
  )
  (get_local $3)
 )
)