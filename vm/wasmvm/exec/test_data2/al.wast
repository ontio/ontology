(module
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "add" (func $add))
 (export "sum" (func $sum))
 (func $add (; 0 ;) (param $0 i32) (param $1 i32) (result i32)
  (i32.add
   (get_local $1)
   (get_local $0)
  )
 )
 (func $sum (; 1 ;) (param $0 i32) (result i32)
  (i32.add
   (i32.load offset=12
    (get_local $0)
   )
   (i32.add
    (i32.load offset=8
     (get_local $0)
    )
    (i32.add
     (i32.load offset=4
      (get_local $0)
     )
     (i32.load
      (get_local $0)
     )
    )
   )
  )
 )
)