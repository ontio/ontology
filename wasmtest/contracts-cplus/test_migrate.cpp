/*test multi storage value should be move to new contract address*/
#include<ontiolib/ontio.hpp>
using std::string;
using std::vector;

#define KEY_MIGRATE_STORE 0x23
#define VAL_MIGRAGE_STORE 0x138297
#define KEY_MIGRATE_STORE2 0x14
#define VAL_MIGRAGE_STORE2 0x32733

namespace ontio {
	struct test_conext {
		address admin;
		std::map<string, address> addrmap;
		ONTLIB_SERIALIZE( test_conext, (admin) (addrmap))
	};
};

using namespace ontio;

class hello: public contract {
	private:
		key t1 = {KEY_MIGRATE_STORE};
		uint64_t val1 = VAL_MIGRAGE_STORE;

		key t2 = {KEY_MIGRATE_STORE2};
		uint64_t val2 = VAL_MIGRAGE_STORE2;
	public:
	using contract::contract;

	void test_contract_migrate(void) {
		int64_t res;
		/** file: test_migrate3.wat
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
		vector<uint8_t> code2 = {0x00,0x61,0x73,0x6d,0x01,0x00,0x00,0x00,0x01,0x12,0x03,0x60,0x00,0x00,0x60,0x02,0x7f,0x7f,0x00,0x60,0x05,0x7f,0x7f,0x7f,0x7f,0x7f,0x01,0x7f,0x02,0x2d,0x02,0x03,0x65,0x6e,0x76,0x0c,0x6f,0x6e,0x74,0x69,0x6f,0x5f,0x72,0x65,0x74,0x75,0x72,0x6e,0x00,0x01,0x03,0x65,0x6e,0x76,0x12,0x6f,0x6e,0x74,0x69,0x6f,0x5f,0x73,0x74,0x6f,0x72,0x61,0x67,0x65,0x5f,0x72,0x65,0x61,0x64,0x00,0x02,0x03,0x02,0x01,0x00,0x05,0x03,0x01,0x00,0x01,0x07,0x0a,0x01,0x06,0x69,0x6e,0x76,0x6f,0x6b,0x65,0x00,0x02,0x0a,0x5b,0x01,0x59,0x00,0x41,0x00,0x41,0x23,0x3a,0x00,0x00,0x41,0x00,0x41,0x01,0x41,0x08,0x41,0x08,0x41,0x00,0x10,0x01,0x41,0x08,0x47,0x04,0x40,0x00,0x0b,0x41,0x08,0x29,0x03,0x00,0x42,0x97,0x85,0xce,0x00,0x52,0x04,0x40,0x00,0x0b,0x41,0x00,0x41,0x14,0x3a,0x00,0x00,0x41,0x00,0x41,0x01,0x41,0x08,0x41,0x08,0x41,0x00,0x10,0x01,0x41,0x08,0x47,0x04,0x40,0x00,0x0b,0x41,0x08,0x29,0x03,0x00,0x42,0xb3,0xce,0x0c,0x52,0x04,0x40,0x00,0x0b,0x41,0x08,0x41,0x08,0x10,0x00,0x0b};
		vector<char> code;
		code.resize(code2.size());
		std::copy(code2.begin(), code2.end(), code.begin());
		address t = ontio::contract_migrate(code, 3, "name", "version", "author", "email", "desc");
		vector<char> args = {};
		call_contract(t,args, res);
		check(res == VAL_MIGRAGE_STORE2, "migrate failed");

		uint64_t b;
		check(!storage_get(t1, b), "should not get storage");
		check(!storage_get(t2, b), "should not get storage");
	}

	void testStorage(void) {
		int64_t b;
		storage_put(t1, val1);
		check(storage_get(t1, b), "get failed");
		check(b == val1, "get wrong");
	}

	void testStorage2(void) {
		int64_t b;
		storage_put(t2, val2);
		check(storage_get(t2, b), "get failed");
		check(b == val2, "get wrong");
	}

	string testcase(void) {
		return string(R"(
		[
			[{"method":"testStorage", "param":"", "expected":""},
			{"method":"testStorage2", "param":"", "expected":""},
			{"method":"test_contract_migrate", "param":"", "expected":""}
			]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(test_contract_migrate)(testStorage)(testStorage2))
