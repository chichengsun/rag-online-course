package auth

import "golang.org/x/crypto/bcrypt"

// HashPassword 使用 bcrypt 对明文密码进行哈希。
func HashPassword(raw string) (string, error) {
	data, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// ComparePassword 比对哈希密码与明文密码是否一致。
func ComparePassword(hash, raw string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(raw))
}
