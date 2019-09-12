#include<ontiolib/ontio.hpp>
using std::string;
using std::vector;

#define KEY_MIGRATE_STORE 0x23
#define VAL_MIGRAGE_STORE 0x138297

namespace ontio {
	struct test_conext {
		address admin;
		std::map<string, address> addrmap;
		ONTLIB_SERIALIZE( test_conext, (admin) (addrmap))
	};
};

using namespace ontio;

class hello: public contract {
	public:
	using contract::contract;

	int64_t test_contract_migrate(void) {
		int64_t res;
		/** file: test_migrate.wat
		(module
		  (type (;0;) (func))
		  (type (;1;) (func (param i32 i32)))
		  (type (;2;) (func (param i32 i32 i32 i32 i32) (result i32)))
		  (import "env" "ontio_return" (func (;0;) (type 1)))
		  (import "env" "ontio_storage_read" (func (;1;) (type 2)))
		  (func (;2;) (type 0)
			i32.const 0  ;; storge offset
			i32.const 35;; init key 0x23(KEY_MIGRATE_STORE)
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
			i32.const 8  ;; data is i64
			call 0		 ;; return
			)
		  (memory (;0;) 1)
		  (export "invoke" (func 2)))
		 **/
		vector<char> code = {0x00,0x61,0x73,0x6d,0x01,0x00,0x00,0x00,0x01,0x12,0x03,0x60,0x00,0x00,0x60,0x02,0x7f,0x7f,0x00,0x60,0x05,0x7f,0x7f,0x7f,0x7f,0x7f,0x01,0x7f,0x02,0x2d,0x02,0x03,0x65,0x6e,0x76,0x0c,0x6f,0x6e,0x74,0x69,0x6f,0x5f,0x72,0x65,0x74,0x75,0x72,0x6e,0x00,0x01,0x03,0x65,0x6e,0x76,0x12,0x6f,0x6e,0x74,0x69,0x6f,0x5f,0x73,0x74,0x6f,0x72,0x61,0x67,0x65,0x5f,0x72,0x65,0x61,0x64,0x00,0x02,0x03,0x02,0x01,0x00,0x05,0x03,0x01,0x00,0x01,0x07,0x0a,0x01,0x06,0x69,0x6e,0x76,0x6f,0x6b,0x65,0x00,0x02,0x0a,0x24,0x01,0x22,0x00,0x41,0x00,0x41,0x23,0x3a,0x00,0x00,0x41,0x00,0x41,0x01,0x41,0x08,0x41,0x08,0x41,0x00,0x10,0x01,0x41,0x08,0x47,0x04,0x40,0x00,0x0b,0x41,0x08,0x41,0x08,0x10,0x00,0x0b};
		address t = ontio::contract_migrate(code, 3, "name", "version", "author", "email", "desc");
		vector<char> args = {};
		call_contract(t,args, res);
		check(res == VAL_MIGRAGE_STORE, "migrate failed");
		return res;
	}

	int64_t testStorage(void) {
		int64_t b;
		int64_t a = VAL_MIGRAGE_STORE;
		key t = {KEY_MIGRATE_STORE};
		storage_put(t, a);
		check(storage_get(t, b), "get failed");
		uint8_t *ptr = (uint8_t *)&b;
		check(ptr[0] == 0x97, "wrong");
		check(ptr[1] == 0x82, "wrong");
		check(ptr[2] == 0x13, "wrong");
		check(ptr[3] == 0x00, "wrong");
		check(b == a, "get wrong");
		return b;
	}

	string testcase(void) {
		return string(R"(
		[
			[{"method":"testStorage", "param":"", "expected":""},
			{"method":"test_contract_migrate", "param":"", "expected":""}
			]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(test_contract_migrate)(testStorage))
