#![cfg_attr(not(feature = "mock"), no_std)]
#![feature(proc_macro_hygiene)]
extern crate ontio_std as ostd;
use ostd::abi::{Sink, Source, EventBuilder};
use ostd::prelude::*;
use ostd::runtime;

#[no_mangle]
pub fn add(a: U128, b: U128) -> U128 {
    a + b
}

#[no_mangle]
pub fn invoke() {
    let input = runtime::input();
    let mut source = Source::new(&input);
    let action: &[u8] = source.read().unwrap();
    let mut sink = Sink::new(12);
    match action {
        b"add" => {
            let (a, b) = source.read().unwrap();
            sink.write(add(a, b))
        }
        b"timestamp" => sink.write(runtime::timestamp()),
        b"block_height" => sink.write(runtime::block_height()),
        b"self_address" => sink.write(runtime::address()),
        b"caller_address" => sink.write(runtime::caller()),
        b"entry_address" => sink.write(runtime::entry_address()),
        b"check_witness" => {
            let addr: Address = source.read().unwrap();
            sink.write(runtime::check_witness(&addr))
        }
        b"current_txhash" => sink.write(runtime::current_txhash()),
        b"current_blockhash" => sink.write(runtime::current_blockhash()),
        b"storage_write" => {
            let (key, val): (&[u8], &[u8]) = source.read().unwrap();
            runtime::storage_write(key, val);
        }
        b"storage_read" => {
            let key: &[u8] = source.read().unwrap();
            if let Some(val) = runtime::storage_read(key) {
                sink.write(val);
            }
        }
        b"storage_delete" => {
            let key: &[u8] = source.read().unwrap();
            runtime::storage_delete(key);
        }
        b"sha256" => {
            let data: &[u8] = source.read().unwrap();
            sink.write(runtime::sha256(&data))
        }
        b"notify" => {
            EventBuilder::new().string("hello").notify();
        },
        b"testcase" => sink.write(testcase()),
        _ => panic!("unsupported action!"),
    }

    runtime::ret(sink.bytes())
}

fn testcase() -> String {
    r#"
    [
        [{"env":{"witness":[]}, "method":"add", "param":"int:1, int:2", "expected":"int:3"},
        {"method":"timestamp"}, {"method":"block_height"}, {"method":"self_address"},
        {"method":"caller_address"}, {"method":"entry_address"},
        {"method":"current_txhash"}, {"method":"current_blockhash"},
        {"method":"storage_write", "param":"string:abc, string:123"},
        {"method":"storage_read", "param":"string:abc", "expected":"string:123"},
        {"method":"storage_delete", "param":"string:abc", "expected":""},
        {"method":"notify", "notify":"hello"}
        ]
    ]
        "#
    .to_string()
}
