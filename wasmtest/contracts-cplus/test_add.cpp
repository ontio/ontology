#include<ontiolib/ontio.hpp>

using namespace ontio;
using std::string;

class hello: public contract {
	public:
	using contract::contract;
	int64_t add(int64_t a, int64_t b) {
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
