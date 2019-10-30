#include<ontiolib/ontio.hpp>
using std::string;
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
	void test_oep4(test_conext &tc) {
		address OWNER = base58toaddress("AQf4Mzu1YJrhz9f3aRkkwSm9n3qhXGSh4p");
		bool res;
		int64_t balanceres = 0;
		address selfaddr = self_address();

		address oep4addr = tc.addrmap["test_oep4.avm"];
		call_neo_contract(oep4addr, pack_neoargs("init", "ignore"), res);
		check(res == true, "oep 4init error");

		call_neo_contract(oep4addr, pack_neoargs("balanceOf", std::tuple<address>(OWNER)), balanceres);
		check(balanceres == 1000000000, "balance error");

		// need OWNER sig.
		call_neo_contract(oep4addr, pack_neoargs("transfer", std::tuple<address, address, asset>(OWNER, selfaddr, 9876)), res);
		check(res == true, "transfer falied");

		call_neo_contract(oep4addr, pack_neoargs("balanceOf", std::tuple<address>(selfaddr)), balanceres);
		check(balanceres == 9876, "balance error");

		call_neo_contract(oep4addr, pack_neoargs("transfer", std::tuple<address, address, asset>(selfaddr, OWNER, 9876)), res);
		check(res == true, "transfer failed");

		call_neo_contract(oep4addr, pack_neoargs("balanceOf", std::tuple<address>(selfaddr)), balanceres);
		check(balanceres == 0, "balance error");
	}

	string testcase(void) {
		return string(R"(
		[
			[{"needcontext":true,"env":{"witness":["AQf4Mzu1YJrhz9f3aRkkwSm9n3qhXGSh4p"]}, "method":"test_oep4", "param":"", "expected":""}
    	    ]
		]
		)");
	}

};

ONTIO_DISPATCH( hello, (test_oep4)(testcase))
