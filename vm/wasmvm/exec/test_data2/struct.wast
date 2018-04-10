(module
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "getMath" (func $getMath))
 (export "getEng" (func $getEng))
 (export "getSum" (func $getSum))
 (func $getMath (; 0 ;) (param $0 i32) (result i32)
  (i32.load
   (get_local $0)
  )
 )
 (func $getEng (; 1 ;) (param $0 i32) (result i32)
  (i32.load offset=4
   (get_local $0)
  )
 )
 (func $getSum (; 2 ;) (param $0 i32) (result i32)
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
