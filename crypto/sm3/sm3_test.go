package sm3

import (
	"crypto/rand"
	"fmt"
	"io"
	"testing"
)

type sm3Test struct {
	out string
	in  string
}

var golden = []sm3Test{
	{"1ab21d8355cfa17f8e61194831e81a8f22bec8c728fefb747ed035eb5082aa2b", ""},
	{"623476ac18f65a2909e43c7fec61b49c7e764a91a18ccb82f1917a29c86c5e88", "a"},
	{"e07d8ee6e54586a459e30eb8d809e02194558e2b0b235a31f3226a3687faab88", "ab"},
	{"66c7f0f462eeedd9d1f2d46bdc10e4e24167c4875cf2f7a2297da02b8f4ba8e0", "abc"},
	{"82ec580fe6d36ae4f81cae3c73f4a5b3b5a09c943172dc9053c69fd8e18dca1e", "abcd"},
	{"afe4ccac5ab7d52bcae36373676215368baf52d3905e1fecbe369cc120e97628", "abcde"},
	{"5d60e23c9fe29b5e62517e144ad67541c6eb132c8926637b6393fe8d9b62b3bf", "abcdef"},
	{"08b7ee8f741bfb63907fcd0029ae3fd6403e6927b50ed9f04665b22eab81e9b7", "abcdefg"},
	{"1fe46fe782fa5618721cdf61de2e50c0639f4b26f6568f9c67b128f5610ced68", "abcdefgh"},
	{"0654f4e3ee1061cbad10a84879af8de6a1c6be9c6e928110a6400b17da1068db", "abcdefghi"},
	{"a30f4e801d2fdc7f2de4bee4d3f5d892b15f6d474a54f5bc96f01b035aa04345", "abcdefghij"},
	{"8a89bd24087ae6f9a3aae485bfa9ecd276f909a04b248eab1b4f9be2b24f0111", "Discard medicine more than two years old."},
	{"2bb6c53ad20eaf2552425f44e72d96d1b61e63310a1a30f4e5406a103619177d", "He who has a shady past knows that nice guys finish last."},
	{"5ecec640017afd77d00147ef42fdb8e7901f089a62c1888637917e89bb3a6532", "I wouldn't marry him with a ten foot pole."},
	{"26598310dfeea2787829ec21d88fbf9f17c9299adf23de49cfcf26030dbc0e35", "Free! Free!/A trip/to Mars/for 900/empty jars/Burma Shave"},
	{"c3555aaf32465c61f681e6dabcc0c95ac93e7c383b1c6eeb621a5ca0eb300508", "The days of the digital watch are numbered.  -Tom Stoppard"},
	{"6ed60c343b5edac968461b8298e20239c320bb59eb81862197d48fd9ac64ed0f", "Nepal premier won't resign."},
	{"199369d5ead103fb01ac6ac507c0c4a6e51d90a787f956dedea18ee9b97d62e0", "For every action there is an equal and opposite government program."},
	{"02e3288a2c5a160c5fae481c1e04dc793818a1dba2448203a82d35a5cd92f36b", "His money is twice tainted: 'taint yours and 'taint mine."},
	{"39aec069bd8643421292a8a94315e84257b925595fb2d23eec55a5edf13f173e", "There is no reason for any individual to have a computer in their home. -Ken Olsen, 1977"},
	{"436ed747ee335dc5f6481ffb8ef1ab8d0b20ecb3476257cbdab7c7715275b1f4", "It's a tiny change to the code and not completely disgusting. - Bob Manchek"},
	{"a84d947e889834860791b13088b664695fb27bf674e9237eaf142076b2390686", "size:  a.out:  bad magic"},
	{"2d64e74ebb40094f0293c6baa04dcb90dde74c109905bff5fd26219db42fe49c", "The major problem is with sendmail.  -Mark Horton"},
	{"4a6a9a28890af31c92863715ff6a919e991f75a99c47700e8bfec17cc89680d5", "Give me a rock, paper and scissors and I will move the world.  CCFestoon"},
	{"6b0899497712fb2b66417c3aaecb3ad23aa730bed6cc59e32020c0f9ea64cd0e", "If the enemy is within range, then so are you."},
	{"6db11897766b88775dda5dc41be73fc50dde13150834bd9bed746c4af79af6ce", "It's well we cannot hear the screams/That we create in others' dreams."},
	{"5a3d123c9457c143b9b2c478fd77926b5f29964448ee8a4e69b0592037b3e811", "You remind me of a TV show, but that's all right: I watch it anyway."},
	{"a015afd04c7a0a4e397d39f328f15972b6553683ef2a6ae861aa1cfc58a67ea6", "C is as portable as Stonehedge!!"},
	{"f165b6f7697a0ebeb508c93d7d5a92ad6ecd24b330994358fa4f39b5ad1ec72a", "Even if I could be Shakespeare, I think I should still choose to be Faraday. - A. Huxley"},
	{"afd8c1159ca40644c9a72b752410b9d076c489109bc0d47d6edc43d81cabf6e3", "The fugacity of a constituent in a mixture of gases at a given temperature is proportional to its mole fraction.  Lewis-Randall Rule"},
	{"f57a610c5a815810a8ec5ff04d43abd4105ce69d3ac64c0f69678d5a270918cd", "How can you write a big system without C++?  -Paul Glick"},
}

func TestGolden(t *testing.T) {
	for i := 0; i < len(golden); i++ {
		g := golden[i]
		s := fmt.Sprintf("%x", Sum([]byte(g.in)))
		if s != g.out {
			t.Fatalf("Sum function: sm3(%s) = %s want %s", g.in, s, g.out)
		}
		c := New()
		for j := 0; j < 3; j++ {
			if j < 2 {
				io.WriteString(c, g.in)
			} else {
				io.WriteString(c, g.in[0:len(g.in)/2])
				c.Sum(nil)
				io.WriteString(c, g.in[len(g.in)/2:])
			}
			s := fmt.Sprintf("%x", c.Sum(nil))
			if s != g.out {
				t.Fatalf("sm3[%d](%s) = %s want %s", j, g.in, s, g.out)
			}
			c.Reset()
		}
	}
}

func TestSize(t *testing.T) {
	c := New()
	if got := c.Size(); got != Size {
		t.Errorf("Size = %d; want %d", got, Size)
	}
}

func TestBlockSize(t *testing.T) {
	c := New()
	if got := c.BlockSize(); got != BlockSize {
		t.Errorf("BlockSize = %d want %d", got, BlockSize)
	}
}

func TestBlockGeneric(t *testing.T) {
	gen, asm := New().(*digest), New().(*digest)
	buf := make([]byte, BlockSize*20) // arbitrary factor
	rand.Read(buf)
	blockGeneric(gen, buf)
	block(asm, buf)
	if *gen != *asm {
		t.Error("block and blockGeneric resulted in different states")
	}
}

var bench = New()
var buf = make([]byte, 8192)

func benchmarkSize(b *testing.B, size int) {
	b.SetBytes(int64(size))
	sum := make([]byte, bench.Size())
	for i := 0; i < b.N; i++ {
		bench.Reset()
		bench.Write(buf[:size])
		bench.Sum(sum[:0])
	}
}

func BenchmarkHash8Bytes(b *testing.B) {
	benchmarkSize(b, 8)
}

func BenchmarkHash1K(b *testing.B) {
	benchmarkSize(b, 1024)
}

func BenchmarkHash8K(b *testing.B) {
	benchmarkSize(b, 8192)
}
