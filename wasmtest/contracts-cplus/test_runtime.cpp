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

	uint64_t timestamp(void) {
		return ontio::timestamp();
	}

	address self_address(void) {
		return ontio::self_address();
	}

	address caller_address(void) {
		return ontio::caller_address();
	}

	address entry_address(void) {
		return ontio::entry_address();
	}

	uint128_t check_witness(test_conext &tc) {
		if(ontio::check_witness(tc.admin))
			return 1;
		else
			return 0;
	}

	H256 current_blockhash(void) {
		return ontio::current_blockhash();
	}

	H256 current_txhash(void) {
		return ontio::current_txhash();
	}

	uint32_t block_height(void) {
		return ontio::block_height();
	}

	string testStorage(int128_t index, string s) {
		string res;
		key t = make_key(index);
		storage_put(t,s);
		check(storage_get(t, res), "get failed");
		check(res == s, "string put failed");
		storage_delete(t);
		check(!storage_get(t, res), "get failed");
		return res;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"env":{"witness":[]}, "method":"self_address", "param":"", "expected":""},
    	    {"env":{"witness":[]}, "method":"entry_address", "param":"", "expected":""},
    	    {"env":{"witness":[]}, "method":"caller_address", "param":"", "expected":""},
    	    {"env":{"witness":[]}, "method":"timestamp", "param":"", "expected":""},
    	    {"needcontext":true, "method":"check_witness", "expected":"int:1"},
    	    {"env":{"witness":[]}, "method":"block_height", "param":"", "expected":""},
    	    {"env":{"witness":[]}, "method":"current_blockhash", "param":"", "expected":""},
    	    {"env":{"witness":[]}, "method":"current_txhash", "param":"", "expected":""},
    	    {"env":{"witness":[]}, "method":"testStorage", "param":"int:1,string:hello world", "expected":"string:hello world"},
    	    {"method":"testStorage", "param":"int:1,string:hello world", "expected":"string:hello world"}
    	    ]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(timestamp)(self_address)(caller_address)(entry_address)(check_witness)(block_height)(current_txhash)(current_blockhash)(testStorage))
