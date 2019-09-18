(module
  (type (;0;) (func))
  (type (;1;) (func (param i32 i32)))
  (import "env" "ontio_return" (func (;0;) (type 1)))
  (func (;1;) (type 0)
		i32.const 0
		i64.const 2222
		i64.store
		i32.const 0
		i32.const 8
		call 0
		)
  (memory (;0;) 1)
  (export "invoke" (func 1)))

