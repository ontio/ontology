(module
 (type $FUNCSIG$i (func (result i32)))
 (type $FUNCSIG$iii (func (param i32 i32) (result i32)))
 (type $FUNCSIG$ii (func (param i32) (result i32)))
 (type $FUNCSIG$iiii (func (param i32 i32 i32) (result i32)))
 (import "env" "arrayLen" (func $arrayLen (param i32) (result i32)))
 (import "env" "calloc" (func $calloc (param i32 i32) (result i32)))
 (import "env" "memcpy" (func $memcpy (param i32 i32 i32) (result i32)))
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "combine" (func $combine))
 (func $combine (; 3 ;) (param $0 i32) (param $1 i32) (result i32)
  (local $2 i32)
  (local $3 i32)
  (local $4 i32)
  (set_local $4
   (call $calloc
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
    (i32.const 4)
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
     (i32.shl
      (get_local $2)
      (i32.const 2)
     )
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
      (i32.shl
       (get_local $2)
       (i32.const 2)
      )
     )
     (get_local $1)
     (i32.shl
      (get_local $3)
      (i32.const 2)
     )
    )
   )
  )
  (get_local $4)
 )
)
