package main

import (
	"flag"
	"fmt"
	"os"

	"driftGo/pkg/encryption"
)

func main() {
	key := flag.String("key", "", "Encryption key (min 32 characters). Defaults to ENCRYPTION_KEY env var.")
	encrypt := flag.String("encrypt", "", "Plaintext to encrypt")
	decrypt := flag.String("decrypt", "", "Ciphertext to decrypt")
	flag.Parse()

	if *key == "" {
		*key = os.Getenv("ENCRYPTION_KEY")
	}
	if *key == "" {
		fmt.Fprintln(os.Stderr, "Error: -key is required (or set ENCRYPTION_KEY env var)")
		flag.Usage()
		os.Exit(1)
	}

	if *encrypt == "" && *decrypt == "" {
		fmt.Fprintln(os.Stderr, "Error: provide -encrypt or -decrypt")
		flag.Usage()
		os.Exit(1)
	}

	enc, err := encryption.NewEncryptor(*key)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating encryptor: %v\n", err)
		os.Exit(1)
	}

	if *encrypt != "" {
		result, err := enc.Encrypt(*encrypt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Encryption error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	}

	if *decrypt != "" {
		result, err := enc.Decrypt(*decrypt)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Decryption error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(result)
	}
}
