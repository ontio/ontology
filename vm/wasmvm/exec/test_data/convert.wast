(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (result i64)))
  (type (;2;) (func (result f32)))
  (type (;3;) (func (result f64)))
  (func (;0;) (type 0) (result i32)
    i64.const -1
    i32.wrap/i64)
  (func (;1;) (type 0) (result i32)
    f32.const -0x1.907e6ap+6 (;=-100.123;)
    i32.trunc_s/f32)
  (func (;2;) (type 0) (result i32)
    f32.const 0x1.65a0bcp+31 (;=3e+09;)
    i32.trunc_u/f32)
  (func (;3;) (type 0) (result i32)
    f64.const -0x1.907e69ad42c3dp+6 (;=-100.123;)
    i32.trunc_s/f64)
  (func (;4;) (type 0) (result i32)
    f64.const 0x1.65a0bcp+31 (;=3e+09;)
    i32.trunc_u/f64)
  (func (;5;) (type 1) (result i64)
    i32.const -1
    i64.extend_u/i32)
  (func (;6;) (type 1) (result i64)
    i32.const -1
    i64.extend_s/i32)
  (func (;7;) (type 0) (result i32)
    f32.const -0x1.907e6ap+6 (;=-100.123;)
    i64.trunc_s/f32
    i64.const -100
    i64.eq)
  (func (;8;) (type 0) (result i32)
    f32.const 0x1.65a0bcp+31 (;=3e+09;)
    i64.trunc_u/f32
    i64.const 3000000000
    i64.eq)
  (func (;9;) (type 0) (result i32)
    f64.const -0x1.907e69ad42c3dp+6 (;=-100.123;)
    i64.trunc_s/f64
    i64.const -100
    i64.eq)
  (func (;10;) (type 0) (result i32)
    f64.const 0x1.65a0bcp+31 (;=3e+09;)
    i64.trunc_u/f64
    i64.const 3000000000
    i64.eq)
  (func (;11;) (type 2) (result f32)
    i32.const -1
    f32.convert_s/i32)
  (func (;12;) (type 2) (result f32)
    i32.const -1
    f32.convert_u/i32)
  (func (;13;) (type 2) (result f32)
    f64.const 0x1.78c29dccccccdp+23 (;=1.23457e+07;)
    f32.demote/f64)
  (func (;14;) (type 2) (result f32)
    i64.const 0
    f32.convert_s/i64)
  (func (;15;) (type 2) (result f32)
    i64.const 0
    f32.convert_u/i64)
  (func (;16;) (type 3) (result f64)
    i32.const -1
    f64.convert_s/i32)
  (func (;17;) (type 3) (result f64)
    i32.const -1
    f64.convert_u/i32)
  (func (;18;) (type 3) (result f64)
    f32.const 0x1.78c29ep+23 (;=1.23457e+07;)
    f64.promote/f32)
  (func (;19;) (type 3) (result f64)
    i64.const 0
    f64.convert_s/i64)
  (func (;20;) (type 3) (result f64)
    i64.const 0
    f64.convert_u/i64)
  (export "i32_wrap_i64" (func 0))
  (export "i32_trunc_s_f32" (func 1))
  (export "i32_trunc_u_f32" (func 2))
  (export "i32_trunc_s_f64" (func 3))
  (export "i32_trunc_u_f64" (func 4))
  (export "i64_extend_u_i32" (func 5))
  (export "i64_extend_s_i32" (func 6))
  (export "i64_trunc_s_f32" (func 7))
  (export "i64_trunc_u_f32" (func 8))
  (export "i64_trunc_s_f64" (func 9))
  (export "i64_trunc_u_f64" (func 10))
  (export "f32_convert_s_i32" (func 11))
  (export "f32_convert_u_i32" (func 12))
  (export "f32_demote_f64" (func 13))
  (export "f32_convert_s_i64" (func 14))
  (export "f32_convert_u_i64" (func 15))
  (export "f64_convert_s_i32" (func 16))
  (export "f64_convert_u_i32" (func 17))
  (export "f64_demote_f32" (func 18))
  (export "f64_convert_s_i64" (func 19))
  (export "f64_convert_u_i64" (func 20)))
