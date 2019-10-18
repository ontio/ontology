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
	address owner;
	key kn = make_key(string("owner"));
	public:
	using contract::contract;
	uint128_t test_native_ont(string &method, address &from, address &to, asset &amount, test_conext &tc) {
		if (method == "balanceOf") {
			asset balance = ont::balanceof(tc.admin);
			check(balance == 1000000000, "init balance wrong");
		} else if (method == "transfer") {
			/*keep admin alway initbalance.*/
			check(ont::transfer(tc.admin, to, amount), "transfer failed");
			check(ont::balanceof(to) == amount, "transfer amount wrong");
			check(ont::transfer(to, tc.admin, amount), "transfer failed");
			check(ont::balanceof(to) == 0, "transfer amount wrong");
		} else if (method == "approve") {
			/*keep admin alway initbalance.*/
			check(ont::approve(tc.admin, from, amount),"approve failed");
			check(ont::allowance(tc.admin, from) == amount, "allowance amount wrong");
			check(ont::transferfrom(from, tc.admin, to, amount),"transferfrom failed");
			check(ont::allowance(tc.admin, from) == 0, "allowance amount wrong");
			check(ont::balanceof(to) == amount, "transfer amount wrong");
			check(ont::transfer(to, tc.admin, amount), "transfer failed");
			check(ont::balanceof(to) == 0, "transfer amount wrong");
			check(ont::balanceof(from) == 0, "transfer amount wrong");
		}

		return 1;
	}

	int128_t balanceOf(address &from) {
		asset balance = ont::balanceof(from);
		int64_t t = int64_t(balance.amount);
		printf("balanceOf is : %lld", t);
		return balance.amount;
	}

	int128_t testranfer(address& from, address &to, asset &amount) {
		check(ont::transfer(from, to, amount), "transfer failed");
		return 1;
	}

	int128_t transfertoowner(address &from, asset &amount) {
		check(storage_get(kn, owner), "get owner key failed");
		check(ont::transfer(from, owner, amount), "transfer failed");
		return 1;
	}

	int128_t setadmin(test_conext &tc) {
		storage_put(kn, tc.admin);
		check(storage_get(kn, owner), "get owner key failed");
		check(owner == tc.admin, "storage failed");
		return 1;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"needcontext":true, "method":"test_native_ont", "param":"string:balanceOf,address:Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT,address:Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV,int:1000", "expected":"int:1"},
    	    {"env":{"witness":["Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT","Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV"]}, "needcontext":true, "method":"test_native_ont", "param":"string:transfer,address:Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT,address:Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV,int:1000", "expected":"int:1"},
    	    {"env":{"witness":["Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT","Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV"]}, "needcontext":true, "method":"test_native_ont", "param":"string:approve,address:Ad4pjz2bqep4RhQrUAzMuZJkBC3qJ1tZuT,address:Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV,int:1000", "expected":"int:1"}
    	    ]
		]
		)");
	}

};

ONTIO_DISPATCH( hello,(testcase)(test_native_ont)(balanceOf)(testranfer)(transfertoowner)(setadmin))
