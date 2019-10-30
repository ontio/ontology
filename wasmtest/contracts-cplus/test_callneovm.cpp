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
		address res;
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("bytearray", test_neo), res);

		check(res == test_neo, "get bytearray result wrong");
		notify_event(res);
		return res;
	}

	address callneo_vector(test_conext &tc) {
		vector<char> res;
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("bytearray", test_neo), res);

		//vector<uint8_t> com(test_neo.begin(), test_neo.end());
		address com;
		check(res.size()== 20, "vector should be 20 Bytes");
		std::copy(res.begin(), res.end(), com.begin());
		check(test_neo == com, "get bytearray result wrong");
		notify_event(com);
		return com;
	}

	string callneo_string(test_conext &tc) {
		string res;
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("string", "test_callneo"), res);
		check(res == string("test_callneo hello my name is steven"), "return error");
		notify_event(res);
		return res;
	}

	bool callneo_bool(test_conext &tc) {
		bool res;
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("boolean", std::tuple<bool, bool>(true, false)), res);
		check(res == true, "return error");
		notify_event(res);

		call_neo_contract(test_neo, pack_neoargs("boolean", std::tuple<bool, bool>(false, false)), res);
		check(res == false, "return error");
		notify_event(res);


		call_neo_contract(test_neo, pack_neoargs("boolean", std::tuple<bool, bool>(false, true)), res);
		check(res == false, "return error");
		notify_event(res);


		call_neo_contract(test_neo, pack_neoargs("boolean", std::tuple<bool, bool>(true, true)), res);
		check(res == true, "return error");
		notify_event(res);
		return res;
	}

	H256 callneo_h256(test_conext &tc) {
		H256 res;
		address test_neo = tc.addrmap["neo_contract.avm"];
		H256 cbhash = current_blockhash();
		call_neo_contract(test_neo, pack_neoargs("H256", cbhash), res);
		check(res == cbhash, "return error");
		notify_event(res);
		return cbhash;
	}

	int128_t callneo_int(test_conext &tc) {
		uint64_t res;
		address test_neo = tc.addrmap["neo_contract.avm"];

		call_neo_contract(test_neo, pack_neoargs("intype", 9), res);
		check(res = 9 + 0x101, "return error");
		notify_event(res);

		// here test overflow error. should be panic.
		//int32_t res_t = 0;
		//call_neo_contract(test_neo, pack_neoargs("intype", int128_t(std::numeric_limits<int64_t>::min())), res_t);

		return res;
	}

	void  callneo_tuples(test_conext &tc) {
		vector<string> addresspool = {"Ab1z3Sxy7ovn4AuScdmMh4PRMvcwCMzSNV", "AW6WWR5xuMpDU5HNoKcta5vm4PUAkByzB5", "Ady2fjRT42bXJBTnNrw1eDPmsVfZ9dHRAB", "AXUpHEB92DDycS7GKjknbF8ZAekskEXkP2", "ASJwxXuVYTJjs2GYVjoJR3W4fyoM7xMfWM", "AdvyrRVzqeR2dQBxB5WrGj7iGKCV93qDP8", "AJxvw1H9zXivoe8mg9toKDQoHDzCHmc4ir", "ATo1YJsYAaLsZKXdnPakgWmhgLdeWFxmFq"};
		vector<address> addr;
		for(auto si : addresspool) {
			addr.push_back(base58toaddress(si));
		}

		string s("hello world");
		address t = addr[0];
		bool b = false;
		int64_t a0= 22;
		int32_t a1= 33;
		int128_t a2= 44;
		H256 a3 = current_blockhash();
		std::vector<uint8_t> arg1{0x11,0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88, 0x99, 0xaa, 0xbb, 0xcc};

		//std::tuple<string, string, address, bool, int64_t, int32_t, int128_t, H256, vector<uint8_t>> tupleargs("hahahah", s, t, b, a0, a1, a2, a3, arg1);
		std::tuple<bool, string, string, address, H256, vector<uint8_t>> tupleargs(false, "hello", s, t, a3, arg1);
		std::tuple<bool, string, string, address, H256, vector<uint8_t>> res;

		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("std", tupleargs), res);
		check(res == tupleargs, "return error");
	}

	void callneo_vectorT(test_conext &tc) {
		vector<uint64_t> res;
		std::vector<uint64_t> arg1{1,0x112233,0x445566, 0x998877};
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("std", arg1), res);
		check(res == arg1, "return error");
		check(res[0] == 1, "xxx");
		check(res[1] == 0x112233, "xxx");
		check(res[2] == 0x445566, "xxx");
		check(res[3] == 0x998877, "xxx");
	}

	void callneo_list(test_conext &tc) {
		std::list<uint64_t> res;
		std::list<uint64_t> arg1{0,0x112233,0x445566, 0x998877};
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("std", arg1), res);
		check(res == arg1, "return error");
	}

	void callneo_deque(test_conext &tc) {
		std::deque<uint64_t> res;
		std::deque<uint64_t> arg1{1, 0x112233,0x445566, 0x998877};
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("std", arg1), res);
		//std::deque<uint64_t> arg2{0x112234,0x445566, 0x998875};
		check(res == arg1, "return error");
	}

	void callneo_pair(test_conext &tc) {
		H256 cbhash = current_blockhash();
		std::pair<int64_t, string> res;
		std::pair<int64_t, string> arg1 = {0, "steven hello"};
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("std", arg1), res);
		//std::deque<uint64_t> arg2{0x112234,0x445566, 0x998875};
		check(res == arg1, "return error");
	}

	void callneo_set(test_conext &tc) {
		std::set<int64_t> res;
		std::set<int64_t> arg1{0, -1234342,-83743, -9223372036854775807};
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("std", arg1), res);
		//std::set<uint64_t> arg2{0x112234,0x445566, 0x998875};
		check(res == arg1, "return error");
		notify_event(res);
	}

	void callneo_array1(test_conext &tc) {
		std::array<int32_t, 6> ttt4 = {0,1,2,-3,-1,5};
		std::array<int32_t, 6> res;
		address test_neo = tc.addrmap["neo_contract.avm"];
		call_neo_contract(test_neo, pack_neoargs("std", ttt4), res);
		//std::array<int32_t, 5> ttt5 = {1,2,3,4,5};

		//check(res == ttt4, "return error");
		notify_event(res);
	}

	string testcase(void) {
		return string(R"(
		[
				[{"needcontext":true,"env":{"witness":[]}, "method":"callneo_address", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_vector", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_string", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_bool", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_h256", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_tuples", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_vectorT", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_list", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_deque", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_pair", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_set", "param":"", "expected":""},
				{"needcontext":true,"env":{"witness":[]}, "method":"callneo_array1", "param":"", "expected":""}
				]
		]
		)");
	}
};

ONTIO_DISPATCH( hello,(callneo_h256)(callneo_string)(callneo_address)(callneo_vector)(callneo_bool)(callneo_int)(callneo_vectorT)(callneo_tuples)(callneo_list)(callneo_deque)(callneo_pair)(callneo_set)(callneo_array1)(testcase))
