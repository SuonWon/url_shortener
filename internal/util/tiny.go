package util

import (
	"crypto/rand"
	"errors"
)

// Base62 without visually confusing chars kept intact for density.
// If you want to avoid lookalikes, replace with e.g. "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnopqrstuvwxyz23456789".
const alphabet = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

// RandCode returns a random Base62 string of length n using crypto/rand.
// It uses rejection sampling to avoid modulo bias.
func RandCode(n int) (string, error) {
	if n <= 0 {
		return "", errors.New("length must be > 0")
	}
	const base = 62
	out := make([]byte, n)

	i := 0
	for i < n {
		// Read a chunk of random bytes; size is flexible—16 is a decent batch.
		buf := make([]byte, 16)
		if _, err := rand.Read(buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			v := int(b & 0x3F) // take 6 bits -> range [0..63]
			if v >= base {
				// reject 62, 63 to avoid bias
				continue
			}
			out[i] = alphabet[v]
			i++
			if i == n {
				break
			}
		}
	}
	return string(out), nil
}
