(module
  (type (;0;) (func (result i32)))
  (type (;1;) (func (result i64)))
  (type (;2;) (func (result f32)))
  (type (;3;) (func (result f64)))
  (func (;0;) (type 0) (result i32)
    i32.const 1
    i32.const 2
    i32.add)
  (func (;1;) (type 0) (result i32)
    i32.const 20
    i32.const 4
    i32.sub)
  (func (;2;) (type 0) (result i32)
    i32.const 3
    i32.const 7
    i32.mul)
  (func (;3;) (type 0) (result i32)
    i32.const -4
    i32.const 2
    i32.div_s)
  (func (;4;) (type 0) (result i32)
    i32.const -4
    i32.const 2
    i32.div_u)
  (func (;5;) (type 0) (result i32)
    i32.const -5
    i32.const 2
    i32.rem_s)
  (func (;6;) (type 0) (result i32)
    i32.const -5
    i32.const 2
    i32.rem_u)
  (func (;7;) (type 0) (result i32)
    i32.const 11
    i32.const 5
    i32.and)
  (func (;8;) (type 0) (result i32)
    i32.const 11
    i32.const 5
    i32.or)
  (func (;9;) (type 0) (result i32)
    i32.const 11
    i32.const 5
    i32.xor)
  (func (;10;) (type 0) (result i32)
    i32.const -100
    i32.const 3
    i32.shl)
  (func (;11;) (type 0) (result i32)
    i32.const -100
    i32.const 3
    i32.shr_u)
  (func (;12;) (type 0) (result i32)
    i32.const -100
    i32.const 3
    i32.shr_s)
  (func (;13;) (type 0) (result i32)
    i32.const -100
    i32.const 3
    i32.rotl)
  (func (;14;) (type 0) (result i32)
    i32.const -100
    i32.const 3
    i32.rotr)
  (func (;15;) (type 1) (result i64)
    i64.const 1
    i64.const 2
    i64.add)
  (func (;16;) (type 1) (result i64)
    i64.const 20
    i64.const 4
    i64.sub)
  (func (;17;) (type 1) (result i64)
    i64.const 3
    i64.const 7
    i64.mul)
  (func (;18;) (type 1) (result i64)
    i64.const -4
    i64.const 2
    i64.div_s)
  (func (;19;) (type 1) (result i64)
    i64.const -4
    i64.const 2
    i64.div_u)
  (func (;20;) (type 1) (result i64)
    i64.const -5
    i64.const 2
    i64.rem_s)
  (func (;21;) (type 1) (result i64)
    i64.const -5
    i64.const 2
    i64.rem_u)
  (func (;22;) (type 1) (result i64)
    i64.const 11
    i64.const 5
    i64.and)
  (func (;23;) (type 1) (result i64)
    i64.const 11
    i64.const 5
    i64.or)
  (func (;24;) (type 1) (result i64)
    i64.const 11
    i64.const 5
    i64.xor)
  (func (;25;) (type 1) (result i64)
    i64.const -100
    i64.const 3
    i64.shl)
  (func (;26;) (type 1) (result i64)
    i64.const -100
    i64.const 3
    i64.shr_u)
  (func (;27;) (type 1) (result i64)
    i64.const -100
    i64.const 3
    i64.shr_s)
  (func (;28;) (type 1) (result i64)
    i64.const -100
    i64.const 3
    i64.rotl)
  (func (;29;) (type 1) (result i64)
    i64.const -100
    i64.const 3
    i64.rotr)
  (func (;30;) (type 2) (result f32)
    f32.const 0x1.4p+0 (;=1.25;)
    f32.const 0x1.ep+1 (;=3.75;)
    f32.add)
  (func (;31;) (type 2) (result f32)
    f32.const 0x1.2p+2 (;=4.5;)
    f32.const 0x1.388p+13 (;=10000;)
    f32.sub)
  (func (;32;) (type 2) (result f32)
    f32.const 0x1.34ap+10 (;=1234.5;)
    f32.const -0x1.b8p+2 (;=-6.875;)
    f32.mul)
  (func (;33;) (type 2) (result f32)
    f32.const 0x1.6bcc42p+46 (;=1e+14;)
    f32.const -0x1.86ap+17 (;=-200000;)
    f32.div)
  (func (;34;) (type 2) (result f32)
    f32.const 0x0p+0 (;=0;)
    f32.const 0x0p+0 (;=0;)
    f32.min)
  (func (;35;) (type 2) (result f32)
    f32.const 0x0p+0 (;=0;)
    f32.const 0x0p+0 (;=0;)
    f32.max)
  (func (;36;) (type 2) (result f32)
    f32.const 0x0p+0 (;=0;)
    f32.const 0x0p+0 (;=0;)
    f32.copysign)
  (func (;37;) (type 3) (result f64)
    f64.const 0x1.d6f34588p+29 (;=9.87654e+08;)
    f64.const 0x1.d6f3454p+26 (;=1.23457e+08;)
    f64.add)
  (func (;38;) (type 3) (result f64)
    f64.const 0x1.3a8a41d39b24ep+196 (;=1.234e+59;)
    f64.const 0x1.d1de3d2d5c713p+78 (;=5.5e+23;)
    f64.sub)
  (func (;39;) (type 3) (result f64)
    f64.const -0x1.2c4bp+20 (;=-1.23e+06;)
    f64.const 0x1.789fe4p+23 (;=1.23412e+07;)
    f64.mul)
  (func (;40;) (type 3) (result f64)
    f64.const 0x1.4e718d7d7625ap+664 (;=1e+200;)
    f64.const 0x1.11b0ec57e649ap+166 (;=1e+50;)
    f64.div)
  (func (;41;) (type 3) (result f64)
    f64.const 0x0p+0 (;=0;)
    f64.const 0x0p+0 (;=0;)
    f64.min)
  (func (;42;) (type 3) (result f64)
    f64.const 0x0p+0 (;=0;)
    f64.const 0x0p+0 (;=0;)
    f64.max)
  (func (;43;) (type 3) (result f64)
    f64.const 0x0p+0 (;=0;)
    f64.const 0x0p+0 (;=0;)
    f64.copysign)
  (export "i32_add" (func 0))
  (export "i32_sub" (func 1))
  (export "i32_mul" (func 2))
  (export "i32_div_s" (func 3))
  (export "i32_div_u" (func 4))
  (export "i32_rem_s" (func 5))
  (export "i32_rem_u" (func 6))
  (export "i32_and" (func 7))
  (export "i32_or" (func 8))
  (export "i32_xor" (func 9))
  (export "i32_shl" (func 10))
  (export "i32_shr_u" (func 11))
  (export "i32_shr_s" (func 12))
  (export "i32_rotl" (func 13))
  (export "i32_rotr" (func 14))
  (export "i64_add" (func 15))
  (export "i64_sub" (func 16))
  (export "i64_mul" (func 17))
  (export "i64_div_s" (func 18))
  (export "i64_div_u" (func 19))
  (export "i64_rem_s" (func 20))
  (export "i64_rem_u" (func 21))
  (export "i64_and" (func 22))
  (export "i64_or" (func 23))
  (export "i64_xor" (func 24))
  (export "i64_shl" (func 25))
  (export "i64_shr_u" (func 26))
  (export "i64_shr_s" (func 27))
  (export "i64_rotl" (func 28))
  (export "i64_rotr" (func 29))
  (export "f32_add" (func 30))
  (export "f32_sub" (func 31))
  (export "f32_mul" (func 32))
  (export "f32_div" (func 33))
  (export "f32_min" (func 34))
  (export "f32_max" (func 35))
  (export "f32_copysign" (func 36))
  (export "f64_add" (func 37))
  (export "f64_sub" (func 38))
  (export "f64_mul" (func 39))
  (export "f64_div" (func 40))
  (export "f64_min" (func 41))
  (export "f64_max" (func 42))
  (export "f64_copysign" (func 43)))
