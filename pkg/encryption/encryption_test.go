package encryption

import (
	"testing"
)

func TestEncryptionDecryption(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	testData := "test-access-token-12345"

	encrypted, err := encryptor.Encrypt(testData)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	if encrypted == testData {
		t.Fatal("Encrypted data should not be the same as original data")
	}

	decrypted, err := encryptor.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	if decrypted != testData {
		t.Fatalf("Decrypted data '%s' does not match original '%s'", decrypted, testData)
	}
}

func TestEncryptionKeyValidation(t *testing.T) {
	invalidKey := "short"
	_, err := NewEncryptor(invalidKey)
	if err != ErrInvalidKeyLength {
		t.Fatalf("Expected ErrInvalidKeyLength, got %v", err)
	}
}

func TestEmptyStringHandling(t *testing.T) {
	key := "12345678901234567890123456789012"
	encryptor, err := NewEncryptor(key)
	if err != nil {
		t.Fatalf("Failed to create encryptor: %v", err)
	}

	encrypted, err := encryptor.Encrypt("")
	if err != nil {
		t.Fatalf("Failed to encrypt empty string: %v", err)
	}
	if encrypted != "" {
		t.Fatal("Empty string should encrypt to empty string")
	}

	decrypted, err := encryptor.Decrypt("")
	if err != nil {
		t.Fatalf("Failed to decrypt empty string: %v", err)
	}
	if decrypted != "" {
		t.Fatal("Empty string should decrypt to empty string")
	}
}
