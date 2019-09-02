#![cfg_attr(not(feature = "mock"), no_std)]
#![feature(proc_macro_hygiene)]
extern crate ontio_std as ostd;
use ostd::abi::{Sink, ZeroCopySource};
use ostd::prelude::*;
use ostd::runtime;
//use ostd::String;

#[no_mangle]
pub fn add(a: u32, b:u32) -> u32 {
    a + b
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
        },
        b"testcase" => sink.write(testcase()),
        _ => panic!("unsupported action!"),
    }

    runtime::ret(sink.bytes())
}

fn testcase() -> String {
    r#"
    [
        [{"env":{"witness":[]}, "method":"add", "param":"int:1, int:2", "expected":"int:3"}
        ]
    ]
        "#.to_string()
}

