package ed25519

import (
	"bytes"
	"testing"
)

func TestComputeSharedSecret(t *testing.T) {
	a, A, _ := GenerateKey()
	b, B, _ := GenerateKey()

	secretA, err := ComputeSharedSecret(a, B)
	if err != nil {
		t.Fatal(err)
	}

	secretB, err := ComputeSharedSecret(b, A)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(secretA, secretB) {
		t.Fatal("aB is not equal to bA")
	}
}

func TestSignVerify(t *testing.T) {
	priv, pub, _ := GenerateKey()

	messages := []string{
		"Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.",
		"Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur.",
		"Sed ut perspiciatis unde omnis iste natus error sit voluptatem accusantium doloremque laudantium, totam rem aperiam, eaque ipsa quae ab illo inventore veritatis et quasi architecto beatae vitae dicta sunt explicabo.",
	}

	for _, msg := range messages {
		sig, err := Sign(priv, pub, []byte(msg))
		if err != nil {
			t.Fatal(err)
		}

		if err := Verify(pub, []byte(msg), sig); err != nil {
			t.Fatal(err)
		}
	}
}
