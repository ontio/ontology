(module
 (type $FUNCSIG$viii (func (param i32 i32 i32)))
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (import "env" "JsonMashal" (func $JsonMashal (param i32 i32) (result i32)))
 (import "env" "JsonUnmashal" (func $JsonUnmashal (param i32 i32 i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 16) "init\00")
 (data (i32.const 32) "init success!\00")
 (data (i32.const 48) "add\00")
 (data (i32.const 64) "int\00")
 (export "memory" (memory $0))
 (export "invoke" (func $invoke))
 (export "add" (func $add))
 (func $invoke (; 2 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (local $3 i32)
  (i32.store offset=4
   (i32.const 0)
   (tee_local $3
    (i32.sub
     (i32.load offset=4
      (i32.const 0)
     )
     (i32.const 16)
    )
   )
  )
  (block $label$0
   (block $label$1
    (br_if $label$1
     (i32.eq
      (get_local $0)
      (i32.const 16)
     )
    )
    (br_if $label$0
     (i32.ne
      (get_local $0)
      (i32.const 48)
     )
    )
    (call $JsonUnmashal
     (i32.add
      (get_local $3)
      (i32.const 8)
     )
     (i32.const 8)
     (get_local $1)
    )
    (set_local $2
     (call $JsonMashal
      (i32.load offset=12
       (get_local $3)
      )
      (i32.const 64)
     )
    )
    (br $label$0)
   )
   (set_local $2
    (i32.const 32)
   )
  )
  (i32.store offset=4
   (i32.const 0)
   (i32.add
    (get_local $3)
    (i32.const 16)
   )
  )
  (get_local $2)
 )
 (func $add (; 3 ;) (param $0 i32) (param $1 i32) (result i32)
  (i32.add
   (get_local $1)
   (get_local $0)
  )
 )
)
