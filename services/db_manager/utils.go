package dbmanager

import (
	"crypto/rand"
	"fmt"
	"math/big"

	"github.com/docker/docker/pkg/namesgenerator"
)

func generateCryptoRandInt(min, max int) (int, error) {
	rangeSize := big.NewInt(int64(max - min + 1))

	n, err := rand.Int(rand.Reader, rangeSize)
	if err != nil {
		return 0, err
	}

	return int(n.Int64()) + min, nil
}

func GenerateSubdomainAndPort() (string, string, string) {
	name := namesgenerator.GetRandomName(10)
	port, err := generateCryptoRandInt(20000, 40000)
	if err != nil {
		return "", "0", ""
	}

	fullName := name + "." + appHost

	return name, fmt.Sprintf("%d", port), fullName
}
