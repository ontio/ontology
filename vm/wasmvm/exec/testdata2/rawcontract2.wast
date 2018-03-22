(module
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (type $FUNCSIG$iiii (func (param i32 i32 i32) (result i32)))
 (import "env" "JsonMashal" (func $JsonMashal (param i32 i32) (result i32)))
 (import "env" "ReadInt32Param" (func $ReadInt32Param (param i32) (result i32)))
 (import "env" "ReadStringParam" (func $ReadStringParam (param i32) (result i32)))
 (import "env" "arrayLen" (func $arrayLen (param i32) (result i32)))
 (import "env" "malloc" (func $malloc (param i32) (result i32)))
 (import "env" "memcpy" (func $memcpy (param i32 i32 i32) (result i32)))
 (import "env" "strcmp" (func $strcmp (param i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 16) "init\00")
 (data (i32.const 32) "init success!\00")
 (data (i32.const 48) "add\00")
 (data (i32.const 64) "int\00")
 (data (i32.const 80) "concat\00")
 (data (i32.const 96) "string\00")
 (export "memory" (memory $0))
 (export "add" (func $add))
 (export "concat" (func $concat))
 (export "invoke" (func $invoke))
 (func $add (; 7 ;) (param $0 i32) (param $1 i32) (result i32)
  (i32.add
   (get_local $1)
   (get_local $0)
  )
 )
 (func $concat (; 8 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (local $3 i32)
  (local $4 i32)
  (set_local $4
   (call $malloc
    (i32.add
     (tee_local $2
      (call $arrayLen
       (get_local $0)
      )
     )
     (tee_local $3
      (call $arrayLen
       (get_local $1)
      )
     )
    )
   )
  )
  (block $label$0
   (br_if $label$0
    (i32.lt_s
     (get_local $2)
     (i32.const 1)
    )
   )
   (drop
    (call $memcpy
     (get_local $4)
     (get_local $0)
     (get_local $2)
    )
   )
  )
  (block $label$1
   (br_if $label$1
    (i32.lt_s
     (get_local $3)
     (i32.const 1)
    )
   )
   (drop
    (call $memcpy
     (i32.add
      (get_local $4)
      (get_local $3)
     )
     (get_local $1)
     (get_local $3)
    )
   )
  )
  (get_local $4)
 )
 (func $invoke (; 9 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (local $3 i32)
  (local $4 i32)
  (block $label$0
   (block $label$1
    (block $label$2
     (br_if $label$2
      (i32.eqz
       (call $strcmp
        (get_local $0)
        (i32.const 16)
       )
      )
     )
     (br_if $label$1
      (i32.eqz
       (call $strcmp
        (get_local $0)
        (i32.const 48)
       )
      )
     )
     (br_if $label$0
      (i32.eqz
       (call $strcmp
        (get_local $0)
        (i32.const 80)
       )
      )
     )
     (return
      (i32.const 0)
     )
    )
    (return
     (i32.const 32)
    )
   )
   (return
    (call $JsonMashal
     (i32.add
      (call $ReadInt32Param
       (get_local $1)
      )
      (call $ReadInt32Param
       (get_local $1)
      )
     )
     (i32.const 64)
    )
   )
  )
  (set_local $2
   (call $ReadStringParam
    (get_local $1)
   )
  )
  (set_local $3
   (call $ReadStringParam
    (get_local $1)
   )
  )
  (set_local $1
   (call $malloc
    (i32.add
     (tee_local $4
      (call $arrayLen
       (get_local $2)
      )
     )
     (tee_local $0
      (call $arrayLen
       (get_local $3)
      )
     )
    )
   )
  )
  (block $label$3
   (br_if $label$3
    (i32.lt_s
     (get_local $4)
     (i32.const 1)
    )
   )
   (drop
    (call $memcpy
     (get_local $1)
     (get_local $2)
     (get_local $4)
    )
   )
  )
  (block $label$4
   (br_if $label$4
    (i32.lt_s
     (get_local $0)
     (i32.const 1)
    )
   )
   (drop
    (call $memcpy
     (i32.add
      (get_local $1)
      (get_local $0)
     )
     (get_local $3)
     (get_local $0)
    )
   )
  )
  (call $JsonMashal
   (get_local $1)
   (i32.const 96)
  )
 )
)
