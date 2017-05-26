package sm3

func block(dig *digest, p []byte) {
	blockGeneric(dig, p)
}

func blockGeneric(dig *digest, p []byte) {
	var w [68]uint32
	var w1 [64]uint32
	var ss1, ss2, tt1, tt2 uint32

	h0, h1, h2, h3, h4, h5, h6, h7 := dig.h[0], dig.h[1], dig.h[2], dig.h[3], dig.h[4], dig.h[5], dig.h[6], dig.h[7]
	for len(p) >= chunk {
		for i := 0; i < 16; i++ {
			j := i * 4
			w[i] = uint32(p[j])<<24 | uint32(p[j+1])<<16 | uint32(p[j+2])<<8 | uint32(p[j+3])
		}
		for i := 16; i < 68; i++ {
			w[i] = sm3_p1(w[i-16]^w[i-9]^sm3_rotl(w[i-3], 15)) ^ sm3_rotl(w[i-13], 7) ^ w[i-6]
		}

		for i := 0; i < 64; i++ {
			w1[i] = w[i] ^ w[i+4]
		}

		a, b, c, d, e, f, g, h := h0, h1, h2, h3, h4, h5, h6, h7

		for j := 0; j < 64; j++ {
			ss1 = sm3_rotl(sm3_rotl(a, 12)+e+sm3_rotl(sm3_t(j), uint32(j)), 7)
			ss2 = ss1 ^ sm3_rotl(a, 12)
			tt1 = sm3_ff(a, b, c, j) + d + ss2 + w1[j]
			tt2 = sm3_gg(e, f, g, j) + h + ss1 + w[j]
			d = c
			c = sm3_rotl(b, 9)
			b = a
			a = tt1
			h = g
			g = sm3_rotl(f, 19)
			f = e
			e = sm3_p0(tt2)
		}

		h0 ^= a
		h1 ^= b
		h2 ^= c
		h3 ^= d
		h4 ^= e
		h5 ^= f
		h6 ^= g
		h7 ^= h

		p = p[chunk:]
	}

	dig.h[0], dig.h[1], dig.h[2], dig.h[3], dig.h[4], dig.h[5], dig.h[6], dig.h[7] = h0, h1, h2, h3, h4, h5, h6, h7
}

func sm3_t(j int) uint32 {
	if j >= 16 {
		return 0x7A879D8A
	} else {
		return 0x79CC4519
	}
}

func sm3_ff(x, y, z uint32, j int) uint32 {
	if j >= 16 {
		return ((x | y) & (x | z) & (y | z))
	} else {
		return x ^ y ^ z
	}

}

func sm3_gg(x, y, z uint32, j int) uint32 {
	if j >= 16 {
		return ((x & y) | ((^x) & z))
	} else {
		return x ^ y ^ z
	}
}

func sm3_rotl(x, n uint32) uint32 {
	return (x << (n % 32)) | (x >> (32 - (n % 32)))
}

func sm3_p0(x uint32) uint32 {
	return x ^ sm3_rotl(x, 9) ^ sm3_rotl(x, 17)
}

func sm3_p1(x uint32) uint32 {
	return x ^ sm3_rotl(x, 15) ^ sm3_rotl(x, 23)
}
