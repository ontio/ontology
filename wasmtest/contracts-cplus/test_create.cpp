#include<ontiolib/ontio.hpp>
using std::string;
using std::vector;

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
	void test_contract_create(void) {
		int64_t res;
		/** file: test_create.wat
			(module
			  (type (;0;) (func))
			  (type (;1;) (func (param i32 i32)))
			  (import "env" "ontio_return" (func (;0;) (type 1)))
			  (func (;1;) (type 0)
				i32.const 0
			    i64.const 2222
				i64.store
				i32.const 0
				i32.const 8
				call 0
				)
			  (memory (;0;) 1)
			  (export "invoke" (func 1)))
		 **/
		vector<uint8_t> code2 = {0x00,0x61,0x73,0x6d,0x01,0x00,0x00,0x00,0x01,0x09,0x02,0x60,0x00,0x00,0x60,0x02,0x7f,0x7f,0x00,0x02,0x14,0x01,0x03,0x65,0x6e,0x76,0x0c,0x6f,0x6e,0x74,0x69,0x6f,0x5f,0x72,0x65,0x74,0x75,0x72,0x6e,0x00,0x01,0x03,0x02,0x01,0x00,0x05,0x03,0x01,0x00,0x01,0x07,0x0a,0x01,0x06,0x69,0x6e,0x76,0x6f,0x6b,0x65,0x00,0x01,0x0a,0x12,0x01,0x10,0x00,0x41,0x00,0x42,0xae,0x11,0x37,0x03,0x00,0x41,0x00,0x41,0x08,0x10,0x00,0x0b};
		vector<char> code;
		code.resize(code2.size());
		std::copy(code2.begin(), code2.end(), code.begin());
		address t = ontio::contract_create(code, 3, "name", "version", "author", "email", "desc");
		vector<char> args = {};
		call_contract(t,args, res);
		check(res == 2222, "migrate failed");
	}
	
	string testcase(void) {
		return string(R"(
		[
			[{"method":"test_contract_create", "param":"", "expected":""}
			]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(test_contract_create))
