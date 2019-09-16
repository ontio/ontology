#include<ontiolib/ontio.hpp>
using std::string;
using std::vector;

#define KEY_MIGRATE_STORE 0x23
#define VAL_MIGRAGE_STORE 0x249308

#define KEY_MIGRATE_STORE2 0x24
#define VAL_MIGRAGE_STORE2 0x372430

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
	void test_contract_destroy(void) {
		int64_t b;
		int64_t a = VAL_MIGRAGE_STORE;
		key t = {KEY_MIGRATE_STORE};
		storage_put(t, a);
		check(storage_get(t, b), "get failed");
		check(b == a, "get wrong");

		a = VAL_MIGRAGE_STORE2;
		key t2 = {KEY_MIGRATE_STORE2};
		storage_put(t, a);
		check(storage_get(t, b), "get failed");
		check(b == a, "get wrong");

		ontio::contract_destroy();
		check(false, "should not be here");
	}

	string testcase(void) {
		return string(R"(
		[
			[{"method":"test_contract_destroy", "param":"", "expected":""}
			]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(test_contract_destroy))
