(module
 (table 0 anyfunc)
 (memory $0 1)
 (export "memory" (memory $0))
 (export "greeting" (func $greeting))
 (func $greeting (; 0 ;) (param $0 i32) (result i32)
  (get_local $0)
 )
)
