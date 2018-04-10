# wasm generate
## tools
1. [Emscripten](https://kripken.github.io/emscripten-site/docs/introducing_emscripten/index.html)

2. [WABT](https://github.com/WebAssembly/wabt)

## SEQ
1. write .c file test.c for example

2. use command ```emcc test.c -Os -s WASM=1 -s SIDE_MODULE=1 -o test.wasm ``` to generate wasm file

3. use command ```wasm2wat test.c > test.wast ``` generate wast file

##WAST Syntax

https://github.com/WebAssembly/spec/blob/master/interpreter/README.md#s-expression-syntax
