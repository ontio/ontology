#include<ontiolib/ontio.hpp>
#include <regex>

using namespace ontio;
using std::string;
using std::regex;

class hello: public contract {
	public:
	using contract::contract;

	int128_t test_regex_match(string &text, string &pattern) {
		regex re(pattern);
        if (std::regex_match(text, re))
			return 1;
		else
			return 0;
	}

	int128_t test_regex_search(string &text, string &pattern) {
		regex re(pattern);
        bool ret = std::regex_search(text, re);
		if (ret) {
			return 1;
		}
		else {
			return 0;
		}
	}

	string test_regex_search2(string &text, string &pattern) {
		std::smatch results;
		std::regex re(pattern);


		std::regex_search(text, results, re);
		auto res = results[0];
		return res;
	}

	string test_regex_replace(string &text, string &pattern, string &replace) {
		std::regex re(pattern);
		string res = std::regex_replace(text, re, replace);
		return res;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"env":{"witness":[]}, "method":"test_regex_match", "param":"string:abcdefgijk,string:a/[bcd/]+ef./[^gh/]?ijk$", "expected":"int:1"},
			 {"env":{"witness":[]}, "method":"test_regex_match", "param":"string:babcdefgijk,string:a/[bcd/]+ef./[^gh/]?ijk$", "expected":"int:0"},
			 {"env":{"witness":[]}, "method":"test_regex_search", "param":"string:babcdefgijk,string:a/[bcd/]+ef./[^gh/]?ijk$", "expected":"int:1"},
			 {"env":{"witness":[]}, "method":"test_regex_search2", "param":"string:hello 2020 bye 2019,string:/[0-9/]{4}", "expected":"string:2020"},
			 {"env":{"witness":[]}, "method":"test_regex_replace", "param":"string:steven name steven,string:steven,string:dalvid", "expected":"string:dalvid name dalvid"},
			 {"env":{"witness":[]}, "method":"test_regex_search", "param":"string:\\,string:\\\\", "expected":"int:1"}
    	    ]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(test_regex_match)(test_regex_search)(test_regex_search2)(test_regex_replace))
