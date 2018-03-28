(module
    (func(export "testif")(param i32)(result i32)
        (if(result i32) (i32.lt_s (get_local 0) (i32.const 5))
        (then (i32.add (get_local 0)(i32.const 10)))
        (else (i32.add (get_local 0)(i32.const 20)))
        )
    )

    (func(export "testifII")(param i32)(result i32)
        (local i32)
        (set_local 1 (i32.const 2 ))
        (i32.mul
            (if(result i32) (i32.lt_s (get_local 0) (i32.const 5))
                    (then (i32.add (get_local 0)(i32.const 10)))
                    (else (i32.add (get_local 0)(i32.const 20)))
                    )
            (get_local 1)
        )
    )


  (func (export "testfor") (param i32) (result i32)
    (local i32 i32)
    (set_local 1 (i32.const 0))
    (set_local 2 (i32.const 0))
    (block
      (loop
        (br_if 1 (i32.gt_u (get_local 2) (get_local 0)))
        (set_local 1 (i32.add (get_local 1) (i32.const 2)))
        (set_local 2 (i32.add (get_local 2) (i32.const 1)))
        (br 0)
      )
    )
    (get_local 1)
  )


    (func (export "testwhile") (param i32) (result i32)
      (local i32)
      (set_local 1 (get_local 0))
      (block
        (loop
          (br_if 1 (i32.eqz (get_local 0)))
          (set_local 1 (i32.add (get_local 1) (i32.const 1)))
          (set_local 0 (i32.sub (get_local 0) (i32.const 1)))
          (br 0)
        )
      )
      (get_local 1)
    )



)