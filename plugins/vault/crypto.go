package vault

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

const (
	// PBKDF2 迭代次数（100,000 次是 OWASP 推荐值）
	pbkdf2Iterations = 100000
	// 密钥长度（AES-256 = 32 字节）
	keyLength = 32
	// Salt 长度
	saltLength = 32
	// Nonce 长度（GCM 推荐值）
	nonceLength = 12
)

var (
	ErrInvalidPassword     = errors.New("invalid master password")
	ErrDecryptionFailed    = errors.New("decryption failed")
	ErrInvalidCiphertext   = errors.New("invalid ciphertext")
	ErrKeyDerivationFailed = errors.New("key derivation failed")
)

// GenerateSalt 生成随机 salt
func GenerateSalt() (string, error) {
	salt := make([]byte, saltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(salt), nil
}

// DeriveKey 使用 PBKDF2 从主密码派生加密密钥
func DeriveKey(password, salt string) []byte {
	saltBytes, _ := base64.StdEncoding.DecodeString(salt)
	return pbkdf2.Key([]byte(password), saltBytes, pbkdf2Iterations, keyLength, sha256.New)
}

// Encrypt 使用 AES-256-GCM 加密数据
// 返回 base64 编码的 (nonce + ciphertext)
func Encrypt(plaintext []byte, key []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	// 生成随机 nonce
	nonce := make([]byte, nonceLength)
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}

	// 加密数据
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	// 将 nonce 附加到密文前面
	result := append(nonce, ciphertext...)
	return base64.StdEncoding.EncodeToString(result), nil
}

// Decrypt 使用 AES-256-GCM 解密数据
// 输入为 base64 编码的 (nonce + ciphertext)
func Decrypt(encrypted string, key []byte) ([]byte, error) {
	data, err := base64.StdEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, ErrInvalidCiphertext
	}

	if len(data) < nonceLength {
		return nil, ErrInvalidCiphertext
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	// 提取 nonce 和密文
	nonce := data[:nonceLength]
	ciphertext := data[nonceLength:]

	// 解密数据
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, ErrDecryptionFailed
	}

	return plaintext, nil
}

// HashPassword 创建密码验证哈希
// 用于快速验证主密码是否正确，无需解密整个文件
func HashPassword(password, salt string) string {
	key := DeriveKey(password, salt)
	// 再做一次哈希来验证
	hash := sha256.Sum256(key)
	return base64.StdEncoding.EncodeToString(hash[:])
}

// VerifyPassword 验证主密码
func VerifyPassword(password, salt, verificationHash string) bool {
	computedHash := HashPassword(password, salt)
	return computedHash == verificationHash
}

// GenerateRandomID 生成随机 ID
func GenerateRandomID() (string, error) {
	id := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, id); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(id), nil
}
