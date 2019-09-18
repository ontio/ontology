#![cfg_attr(not(feature = "mock"), no_std)]
#![feature(proc_macro_hygiene)]
extern crate ontio_std as ostd;
use ostd::abi::{Sink, ZeroCopySource};
use ostd::prelude::*;
use ostd::runtime;

extern crate alloc;
use alloc::collections::BTreeMap;
use ontio_std::console::debug;

#[no_mangle]
pub fn add(a: u64, b: u64) -> u64 {
    a + b
}

pub struct TestContext<'a> {
    admin: &'a Addr,
    map: BTreeMap<String, &'a Addr>,
}

#[no_mangle]
pub fn invoke() {
    let input = runtime::input();
    let mut source = ZeroCopySource::new(&input);
    let action: &[u8] = source.read().unwrap();
    let mut sink = Sink::new(12);
    match action {
        b"add" => {
            let (a, b) = source.read().unwrap();
            sink.write(add(a, b))
        }
        b"test_migrate" => {
            /** file: test_migrate3.wat
            (module
              (type (;0;) (func))
              (type (;1;) (func (param i32 i32)))
              (type (;2;) (func (param i32 i32 i32 i32 i32) (result i32)))
              (import "env" "ontio_return" (func (;0;) (type 1)))
              (import "env" "ontio_storage_read" (func (;1;) (type 2)))
              (func (;2;) (type 0)
                i32.const 0  ;; storge offset
                i32.const 37;; init key 0x25(KEY_MIGRATE_STORE)
                i32.store8   ;; store key
                i32.const 0  ;; key addr
                i32.const 1  ;; key size 1 byte
                i32.const 8  ;; data addr
                i32.const 8  ;; data is i64
                i32.const 0  ;; always 0
                call 1		 ;; read key to data addr of data size
                i32.const 8  ;; should read 8
                i32.ne
                if
                    unreachable ;; if read length not 8 will panic
                end
                i32.const 8  ;; data addr
                i64.load
                i64.const 1278615 ;; data(VAL_MIGRAGE_STORE)
                i64.ne
                if
                    unreachable ;;
                end			 ;; ===first key storage check  end===
                i32.const 0  ;; storge offset
                i32.const 20 ;; init key 0x14(KEY_MIGRATE_STORE2)
                i32.store8   ;; store key
                i32.const 0  ;; key addr
                i32.const 1  ;; key size 1 byte
                i32.const 8  ;; data addr
                i32.const 8  ;; data is i64
                i32.const 0  ;; always 0
                call 1		 ;; read key to data addr of data size
                i32.const 8  ;; should read 8
                i32.ne
                if
                    unreachable ;; if read length not 8 will panic
                end
                i32.const 8  ;; data addr
                i64.load
                i64.const 206643;; data(VAL_MIGRAGE_STORE2)
                i64.ne
                if
                    unreachable ;;
                end
                i32.const 8  ;; data addr
                i32.const 8  ;; data is i64
                call 0		 ;; return
                )
              (memory (;0;) 1)
              (export "invoke" (func 2)))
             **/
            let code = &[
                0x00, 0x61, 0x73, 0x6d, 0x01, 0x00, 0x00, 0x00, 0x01, 0x12, 0x03, 0x60, 0x00, 0x00,
                0x60, 0x02, 0x7f, 0x7f, 0x00, 0x60, 0x05, 0x7f, 0x7f, 0x7f, 0x7f, 0x7f, 0x01, 0x7f,
                0x02, 0x2d, 0x02, 0x03, 0x65, 0x6e, 0x76, 0x0c, 0x6f, 0x6e, 0x74, 0x69, 0x6f, 0x5f,
                0x72, 0x65, 0x74, 0x75, 0x72, 0x6e, 0x00, 0x01, 0x03, 0x65, 0x6e, 0x76, 0x12, 0x6f,
                0x6e, 0x74, 0x69, 0x6f, 0x5f, 0x73, 0x74, 0x6f, 0x72, 0x61, 0x67, 0x65, 0x5f, 0x72,
                0x65, 0x61, 0x64, 0x00, 0x02, 0x03, 0x02, 0x01, 0x00, 0x05, 0x03, 0x01, 0x00, 0x01,
                0x07, 0x0a, 0x01, 0x06, 0x69, 0x6e, 0x76, 0x6f, 0x6b, 0x65, 0x00, 0x02, 0x0a, 0x5b,
                0x01, 0x59, 0x00, 0x41, 0x00, 0x41, 0x25, 0x3a, 0x00, 0x00, 0x41, 0x00, 0x41, 0x01,
                0x41, 0x08, 0x41, 0x08, 0x41, 0x00, 0x10, 0x01, 0x41, 0x08, 0x47, 0x04, 0x40, 0x00,
                0x0b, 0x41, 0x08, 0x29, 0x03, 0x00, 0x42, 0x97, 0x85, 0xce, 0x00, 0x52, 0x04, 0x40,
                0x00, 0x0b, 0x41, 0x00, 0x41, 0x14, 0x3a, 0x00, 0x00, 0x41, 0x00, 0x41, 0x01, 0x41,
                0x08, 0x41, 0x08, 0x41, 0x00, 0x10, 0x01, 0x41, 0x08, 0x47, 0x04, 0x40, 0x00, 0x0b,
                0x41, 0x08, 0x29, 0x03, 0x00, 0x42, 0xb3, 0xce, 0x0c, 0x52, 0x04, 0x40, 0x00, 0x0b,
                0x41, 0x08, 0x41, 0x08, 0x10, 0x00, 0x0b,
            ];

            if let Some(address) =
                runtime::contract_migrate(code, 3, "name", "version", "author", "email", "desc")
            {
                if let Some(resv) = runtime::call_contract(&address, &[]) {
                    let mut source = ZeroCopySource::new(&resv);
                    let val: u64 = source.read().unwrap();
                    assert!(val == 0x32733);
                } else {
                    panic!("call migrate failed");
                }
            } else {
                panic!("migrate failed");
            }
        }
        b"storage_write" => {
            let mut sink = Sink::new(8);
            let key: [u8; 1] = [0x25];
            let val: u64 = 0x138297;
            sink.write(val);
            runtime::storage_write(&key, sink.bytes());
            if let Some(val) = runtime::storage_read(&key) {
                let mut ressource = ZeroCopySource::new(&val);
                let res: u64 = ressource.read().unwrap();
                assert!(res == 0x138297);
            } else {
                panic!("write do not get val");
            }
        }
        b"storage_write2" => {
            let mut sink = Sink::new(8);
            let key: [u8; 1] = [0x14];
            let val: u64 = 0x32733;
            sink.write(val);
            runtime::storage_write(&key, sink.bytes());
            if let Some(val) = runtime::storage_read(&key) {
                let mut ressource = ZeroCopySource::new(&val);
                let res: u64 = ressource.read().unwrap();
                assert!(res == 0x32733);
            } else {
                panic!("write2 do not get val");
            }
        }
        b"test_callwasm" => {
            let mut isink = Sink::new(20);
            let helloaction: &[u8] = source.read().unwrap();
            let (a, b): (u64, u64) = source.read().unwrap();
            //debug(&format!("{:}", String::from_utf8(helloaction.to_vec()).unwrap()));

            isink.write(helloaction);
            isink.write(a);
            isink.write(b);
            let tc = get_tc(&mut source);
            let address = tc.map[&String::from("helloworld.wasm")];
            if let Some(resv) = runtime::call_contract(&address, isink.bytes()) {
                runtime::ret(&resv);
                return;
            } else {
                panic!("get no return");
            }
        }
        b"test_native" => {}
        b"testcase" => sink.write(testcase()),
        _ => panic!("unsupported action!"),
    }

    runtime::ret(sink.bytes())
}

fn get_tc<'a>(source: &mut ZeroCopySource<'a>) -> TestContext<'a> {
    let mut map = BTreeMap::new();
    let addr = source.read_addr().unwrap();
    let n = source.read_varuint().unwrap();
    for _i in 0..n {
        let (file, addr): (&str, &Addr) = source.read().unwrap();
        map.insert(file.to_string(), addr);
    }

    TestContext { admin: addr, map: map }
}

fn testcase() -> String {
    r#"
    [
        [{"method":"storage_write"},
        {"method":"storage_write2"},
        {"needcontext":true, "env":{"witness":[]}, "method":"test_callwasm", "param":"string:add, int:1, int:2", "expected":"int:3"},
        {"method":"test_migrate"}
        ]
    ]
        "#
    .to_string()
}
