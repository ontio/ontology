(module
 (type $FUNCSIG$i (func (result i32)))
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (type $FUNCSIG$iiii (func (param i32 i32 i32) (result i32)))
 (import "env" "arrayLen" (func $arrayLen (param i32) (result i32)))
 (import "env" "malloc" (func $malloc (param i32) (result i32)))
 (import "env" "memcpy" (func $memcpy (param i32 i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (data (i32.const 16) "hello \00")
 (export "memory" (memory $0))
 (export "scontact" (func $scontact))
 (export "contactHello" (func $contactHello))
 (func $scontact (; 3 ;) (param $0 i32) (param $1 i32) (result i32)
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
      (get_local $2)
     )
     (get_local $1)
     (get_local $3)
    )
   )
  )
  (get_local $4)
 )
 (func $contactHello (; 4 ;) (param $0 i32) (result i32)
  (local $1 i32)
  (local $2 i32)
  (local $3 i32)
  (set_local $3
   (call $malloc
    (i32.add
     (tee_local $1
      (call $arrayLen
       (i32.const 16)
      )
     )
     (tee_local $2
      (call $arrayLen
       (get_local $0)
      )
     )
    )
   )
  )
  (block $label$0
   (br_if $label$0
    (i32.lt_s
     (get_local $1)
     (i32.const 1)
    )
   )
   (drop
    (call $memcpy
     (get_local $3)
     (i32.const 16)
     (get_local $1)
    )
   )
  )
  (block $label$1
   (br_if $label$1
    (i32.lt_s
     (get_local $2)
     (i32.const 1)
    )
   )
   (drop
    (call $memcpy
     (i32.add
      (get_local $3)
      (get_local $1)
     )
     (get_local $0)
     (get_local $2)
    )
   )
  )
  (get_local $3)
 )
)
