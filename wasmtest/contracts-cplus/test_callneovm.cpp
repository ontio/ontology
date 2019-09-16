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
	address callneo_address(test_conext &tc) {
		NeoByteArray res;
		NeoByteArray arg0;
		NeoAddress arg1;
		string s("bytearray");
		address test_neo = tc.addrmap["neo_contract.avm"];
		arg0.resize(s.size());
		std::copy(s.begin(), s.end(), arg0.begin());
		arg1.value = test_neo;

		auto neoargs = serialize_args_forneo(arg0, arg1);
		call_neo_contract(test_neo, neoargs, res);
		address res_addr;
		std::copy(res.begin(), res.end(), res_addr.begin());
		check(res_addr == test_neo, "get bytearray result wrong");
		return res_addr;
	}

	bool callneo_bool(test_conext &tc) {
		address test_neo = tc.addrmap["neo_contract.avm"];
		string s("boolean");
		NeoByteArray arg0;
		arg0.resize(s.size());
		std::copy(s.begin(), s.end(), arg0.begin());
		NeoBoolean arg1{true};
		NeoBoolean res{false};
		auto neoargs = serialize_args_forneo(arg0, arg1);
		call_neo_contract(test_neo, neoargs, res);
		check(res.value == arg1.value, "get bool result wrong");
		return res.value;
	}

	void callneo_intype(test_conext &tc) {
		address test_neo = tc.addrmap["neo_contract.avm"];
		string s("intype");
		NeoByteArray arg0;
		arg0.resize(s.size());
		std::copy(s.begin(), s.end(), arg0.begin());

		NeoInt<int64_t>  arg1{0x349872};
		NeoInt<int64_t>  res;
		auto neoargs = serialize_args_forneo(arg0, arg1);
		call_neo_contract(test_neo, neoargs, res);
		check(res.value == arg1.value + 0x101, "get int64_t result wrong");

		vector<uint8_t>  argt{0x83,0x71,0x29,0x72,0x46,0x89,0x65,0x73,0x26,0x41,0x82,0x68,0x88};
		NeoInt<int128_t>  arg2;
		// must init to zero. or the later data will get random.
		arg2.value = 0;
		memcpy(((uint8_t*)&(arg2.value)),argt.data(), argt.size());
		NeoInt<int128_t>  res2;
		neoargs = serialize_args_forneo(arg0, arg2);
		call_neo_contract(test_neo, neoargs, res2);
		/*if use negtive value. the return value will unkown. int128_t to negtive value may only take 64bit or 32 bit.*/
		check(res2.value == arg2.value + 0x101, "get int128_t result wrong");
	}

	void callneo_H256(test_conext &tc) {
		address test_neo = tc.addrmap["neo_contract.avm"];
		string s("H256");
		NeoByteArray arg0;
		arg0.resize(s.size());
		std::copy(s.begin(), s.end(), arg0.begin());

		NeoH256 arg1;
		arg1.value = current_blockhash();
		NeoByteArray resl;
		auto neoargs = serialize_args_forneo(arg0, arg1);
		call_neo_contract(test_neo, neoargs, resl);
		NeoH256 res;
		std::copy(resl.begin(), resl.end(), res.value.begin());
		check(res.value == arg1.value, "get wrong h256");
	}

	void callneo_listype(test_conext &tc) {
		NeoInt<int64_t>  arg1{0x349872};
		NeoInt<int64_t>  arg3{0x349872};
		NeoByteArray arg0;
		string s("neolist");
		address test_neo = tc.addrmap["neo_contract.avm"];
		arg0.resize(s.size());
		std::copy(s.begin(), s.end(), arg0.begin());

		NeoList<NeoInt<int64_t>, NeoByteArray, NeoInt<int64_t>> args;
		std::get<0>(args.value) = arg1;
		std::get<1>(args.value) = arg0;
		std::get<2>(args.value) = arg3;
		auto neoargs = serialize_args_forneo(arg0, args);

		NeoList<NeoInt<int64_t>, NeoByteArray, NeoInt<int64_t>> res;
		call_neo_contract(test_neo, neoargs, res);
		check(std::get<0>(res.value).value == 0x349873, "list get first result error");
		check(std::get<1>(res.value) == arg0, "list get result second error");
		check(std::get<2>(res.value).value == 0x349871, "list get third result error");
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"needenv":true,"env":{"witness":[]}, "method":"callneo_address", "param":"", "expected":""},
			{"needenv":true,"env":{"witness":[]}, "method":"callneo_bool", "param":"", "expected":""},
			{"needenv":true,"env":{"witness":[]}, "method":"callneo_intype", "param":"", "expected":""},
			{"needenv":true,"env":{"witness":[]}, "method":"callneo_H256", "param":"", "expected":""},
			{"needenv":true,"env":{"witness":[]}, "method":"callneo_listype", "param":"", "expected":""}
    	    ]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(callneo_address)(callneo_bool)(callneo_intype)(callneo_H256)(callneo_listype))
