#ifndef WASMJIT_RUNTIME_H
#define WASMJIT_RUNTIME_H
#include"wasmjit.h"
#include<string.h>

typedef struct {
	uint32_t v;
	wasmjit_result_t res;
} wasmjit_u32;

typedef struct {
	uint64_t exec_step;
	uint64_t gas_left;
	wasmjit_bytes_t buffer;
	wasmjit_result_t res;
} wasmjit_ret;

wasmjit_result_t wasmjit_construct_result(uint8_t* data_buffer, uint32_t data_len, wasmjit_result_kind);

uint64_t wasmjit_service_index(wasmjit_vmctx_t *ctx);

wasmjit_ret wasmjit_invoke(wasmjit_slice_t code, wasmjit_chain_context_t *ctx);

void wasmjit_set_calloutput(wasmjit_vmctx_t *ctx, uint8_t *data, uint32_t len);

uint64_t wasmjit_get_gas(wasmjit_vmctx_t *ctx);
uint64_t wasmjit_get_exec_step(wasmjit_vmctx_t *ctx);
void wasmjit_set_gas(wasmjit_vmctx_t *ctx, uint64_t gas);
void wasmjit_set_exec_step(wasmjit_vmctx_t *ctx, uint64_t exec_step);
#endif
