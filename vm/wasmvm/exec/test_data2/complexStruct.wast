(module
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "sumAcct1" (func $sumAcct1))
 (export "getAcct1" (func $getAcct1))
 (func $sumAcct1 (; 0 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (block $label$0
   (br_if $label$0
    (i32.lt_s
     (get_local $1)
     (i32.const 1)
    )
   )
   (set_local $0
    (i32.load
     (get_local $0)
    )
   )
   (set_local $2
    (i32.const 0)
   )
   (loop $label$1
    (set_local $2
     (i32.add
      (i32.load
       (get_local $0)
      )
      (get_local $2)
     )
    )
    (set_local $0
     (i32.add
      (get_local $0)
      (i32.const 4)
     )
    )
    (br_if $label$1
     (tee_local $1
      (i32.add
       (get_local $1)
       (i32.const -1)
      )
     )
    )
   )
   (return
    (get_local $2)
   )
  )
  (i32.const 0)
 )
 (func $getAcct1 (; 1 ;) (param $0 i32) (param $1 i32) (result i32)
  (i32.load
   (i32.add
    (i32.load
     (get_local $0)
    )
    (i32.shl
     (get_local $1)
     (i32.const 2)
    )
   )
  )
 )
)
