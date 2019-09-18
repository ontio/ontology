#![no_std]
#![feature(proc_macro_hygiene)]
extern crate ontio_std as ostd;

use ostd::abi::{Encoder, Sink, Source};
use ostd::prelude::*;
use ostd::runtime;
use ostd::str::from_utf8;

extern crate hexutil;

const KEY_TOTAL_SUPPLY: &str = "total_supply";
const NAME: &str = "wasm_token";
const SYMBOL: &str = "WTK";
const TOTAL_SUPPLY: u64 = 100000000000;

const _ADDR_EMPTY: Address = ostd::base58!("AFmseVrdL9f9oyCzZefL9tG6UbvhPbdYzM");

fn create_contract() -> Address {
    let code = "0061736d0100000001090260027f7f0060000002140103656e760c6f6e74696f5f72657475726e00000303020101040501700101010503010001070a0106696e766f6b6500010a130205001002000b0b0041808002410b1000000b0b130100418080020b0b68656c6c6f20776f726c64";
    let code_vec = hexutil::read_hex(code).expect("parse hex failed");
    let addr = runtime::contract_create(
        code_vec.as_slice(),
        3,
        "name",
        "version",
        "author",
        "email",
        "desc",
    );
    if let Some(addr_temp) = addr {
        return addr_temp;
    } else {
        return Address::zero();
    }
}

fn call_wasm(addr: &Address) -> String {
    let res = runtime::call_contract(addr, "".as_bytes());
    if let Some(r) = res {
        return String::from_utf8(r).unwrap();
    } else {
        "".to_string()
    }
}

#[no_mangle]
pub fn invoke() {
    let input = runtime::input();
    let mut source = Source::new(input);
    let action: String = source.read().unwrap();
    let mut sink = Sink::new(12);
    match action.as_str() {
        "create_contract" => {
            sink.write(create_contract());
        }
        "call_wasm" => {
            let addr = source.read().unwrap();
            sink.write(call_wasm(&addr));
        }
        "testcase" => {
            sink.write(testcase());
        }
        _ => panic!("unsupported action!"),
    }

    runtime::ret(sink.bytes())
}

fn testcase() -> String {
    r#"
    [
        [{"method":"create_contract", "expected":"address:ALZZb7uraVvnNcFvd9YFE6M4S5TRdc8cnJ"},
        {"method":"call_wasm","param":"address:ALZZb7uraVvnNcFvd9YFE6M4S5TRdc8cnJ","expected":"string:hello world"}
        ]
    ]
    "#.to_string()
}
