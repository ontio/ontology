#![cfg_attr(not(feature = "mock"), no_std)]
extern crate ontio_std as ostd;
use ostd::abi::{Sink, Source};
use ostd::prelude::*;
use ostd::runtime;
use ostd::contract::ont;

extern crate alloc;
use alloc::collections::BTreeMap;
use ontio_std::console::debug;

pub struct TestContext<'a> {
    admin: &'a Address,
    map: BTreeMap<String, &'a Address>,
}

#[no_mangle]
pub fn invoke() {
    let input = runtime::input();
    let mut source = Source::new(&input);
    let action: &[u8] = source.read().unwrap();
    let mut sink = Sink::new(12);
    match action {
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

            let address =
                runtime::contract_migrate(code, 3, "name", "version", "author", "email", "desc")
                    .expect("migrate failed");
            let resv = runtime::call_contract(&address, &[]).expect("call_contract failed");
            let mut source = Source::new(&resv);
            let val: u64 = source.read().unwrap();

            assert_eq!(val, 0x32733);
        }
        b"storage_write" => {
            let mut sink = Sink::new(8);
            let key: [u8; 1] = [0x25];
            let val: u64 = 0x138297;
            sink.write(val);
            runtime::storage_write(&key, sink.bytes());
            let val = runtime::storage_read(&key).expect("read val error");
            let mut ressource = Source::new(&val);
            let res: u64 = ressource.read().unwrap();
            assert_eq!(res, 0x138297);
        }
        b"storage_write2" => {
            let mut sink = Sink::new(8);
            let key: [u8; 1] = [0x14];
            let val: u64 = 0x32733;
            sink.write(val);
            runtime::storage_write(&key, sink.bytes());
            let val = runtime::storage_read(&key).expect("read val error");
            let mut ressource = Source::new(&val);
            let res: u64 = ressource.read().unwrap();
            assert_eq!(res, 0x32733);
        }
        b"test_callwasm" => {
            let mut isink = Sink::new(20);
            let helloaction: &[u8] = source.read().unwrap();
            let (a, b): (u128, u128) = source.read().unwrap();
            //debug(&format!("{:}", String::from_utf8(helloaction.to_vec()).unwrap()));

            isink.write(helloaction);
            isink.write(a);
            isink.write(b);
            let tc = get_tc(&mut source);
            let address = tc.map["helloworld.wasm"];
            let resv = runtime::call_contract(&address, isink.bytes()).expect("get no return");
            runtime::ret(&resv);
            return;
        }
        b"test_calljsvm" => {
            let mut isink = Sink::new(20);
            let helloaction: &[u8] = source.read().unwrap();
            let a:String = source.read().unwrap();
            //debug(&format!("{:}", String::from_utf8(helloaction.to_vec()).unwrap()));

            isink.write(helloaction);
            isink.write(a);
            let tc = get_tc(&mut source);
            let address = tc.map["jsvm.wasm"];
            let resv = runtime::call_contract(&address, isink.bytes()).expect("get no return");
            runtime::ret(&resv);
            return;
        }
        b"balanceOf" => {
            let balance = ont::balance_of(&runtime::address());
            sink.write(balance)
        }
        b"transferFromAdmin" => {
            let amount = source.read().unwrap();
            let tc = get_tc(&mut source);
            sink.write(ont::transfer(tc.admin, &runtime::address(), amount))
        }
        b"transferToAdmin" => {
            let amount = source.read().unwrap();
            let tc = get_tc(&mut source);
            sink.write(ont::transfer(&runtime::address(),tc.admin, amount))
        }
        b"approveAndTransferFromAdmin" => {
            let amount = source.read().unwrap();
            let tc = get_tc(&mut source);
            let res = ont::approve(tc.admin, &runtime::address(), amount);
            assert_eq!(res, true);
            let allo = ont::allowance(tc.admin, &runtime::address());
            assert_eq!(allo, amount);
            sink.write(ont::transfer_from(&runtime::address(), tc.admin, &runtime::address(), amount))
        }
        b"testcase" => sink.write(testcase()),
        _ => panic!("unsupported action!"),
    }

    runtime::ret(sink.bytes())
}

fn get_tc<'a>(source: &mut Source<'a>) -> TestContext<'a> {
    let mut map = BTreeMap::new();
    let admin = source.read().unwrap();
    let n = source.read_varuint().unwrap();
    for _i in 0..n {
        let (file, addr): (&str, &Address) = source.read().unwrap();
        map.insert(file.to_string(), addr);
    }

    TestContext { admin, map }
}

fn testcase() -> String {
    r#"
    [
        [
        {"method":"balanceOf", "expected":"int:0"},
        {"method":"transferFromAdmin", "needcontext":true, "param":"int:100", "expected":"bool:true"},
        {"method":"balanceOf", "expected":"int:100"},
        {"method":"approveAndTransferFromAdmin", "needcontext":true, "param":"int:100", "expected":"bool:true"},
        {"method":"balanceOf", "expected":"int:200"},
        {"method":"transferToAdmin", "needcontext":true, "param":"int:200", "expected":"bool:true"},
        {"method":"balanceOf", "expected":"int:0"},

        {"method":"storage_write"},
        {"method":"storage_write2"},
        {"needcontext":true, "env":{"witness":[]}, "method":"test_callwasm", "param":"string:add, int:1, int:2", "expected":"int:3"},
        {"needcontext":true, "env":{"witness":[]}, "method":"test_calljsvm", "param":"string:evaluate,string:function fib(n) {if(n<=0){return 1;} else if (n == 1) {return 1;} else { return fib(n-1) + fib(n-2);}} fib(3)", "expected":"string:3"},
        {"needcontext":true, "env":{"witness":[]}, "method":"test_calljsvm", "param":"string:evaluate,string:function sum(a) {var sum = 0;var i = 0;while (i<a) {sum = sum + i;i = i + 1;} return sum;};sum(5);", "expected":"string:10"},
        {"needcontext":true, "env":{"witness":[]}, "method":"test_calljsvm", "param":"string:evaluate,string:function mul(a) { return a*a;}; function sum(n) { let sum = 0; let i = 0; while( i<n) { sum = sum + mul(i);i = i + 1}; return sum;};sum(10);", "expected":"string:285"},
        {"method":"test_migrate"}
        ]
    ]
        "#
    .to_string()
}
