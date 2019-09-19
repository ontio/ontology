#include<ontiolib/ontio.hpp>

using namespace ontio;
using std::string;

class hello: public contract {
	public:
	using contract::contract;
	int128_t add(int128_t a, int128_t b) {
		return a + b;
	}

	string testcase(void) {
		return string(R"(
		[
    	    [{"env":{"witness":[]}, "method":"add", "param":"int:1, int:2", "expected":"int:3"}
    	    ]
		]
		)");
	}
};

ONTIO_DISPATCH( hello, (testcase)(add))
