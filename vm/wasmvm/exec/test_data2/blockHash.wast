(module
    (type (;0;) (func(result i32)))
    (import "env" "getBlockHeight" (func (;0;) (type 0)))
    (import "env" "getBlockHash" (func(;1;)(param i32 i32)(result i32)))
    (memory  1)
    (func (export "getBlockHeight")(result i32)
        call 0
    )
    (func (export "getBlockHash")(param $i i32)(result i32)
        get_local $i
        i32.const 10
        call 1
    )
)