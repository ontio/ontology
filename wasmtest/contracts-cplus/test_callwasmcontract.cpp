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

	int128_t call_wasm_contract(int128_t a, int128_t b, test_conext &tc) {
		int128_t res;
		address test_add = tc.addrmap["test_add.wasm"];
		auto args = pack(string("add"), a, b);
		call_contract(test_add, args, res);
		check(res == a + b, "call wasm contract wrong");
		return res;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"needcontext":true,"method":"call_wasm_contract", "param":"int:1,int:2", "expected":"int:3"}
    	    ]
		]
		)");
	}

};

ONTIO_DISPATCH( hello,(testcase)(call_wasm_contract))
