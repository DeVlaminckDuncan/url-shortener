package shortener

import (
	"crypto/sha256"
	"fmt"
	"math/big"
	"os"

	"github.com/itchyny/base58-go"
)

func generateSha256Bytes(input string) []byte {
	algorithm := sha256.New()
	algorithm.Write([]byte(input))

	return algorithm.Sum(nil)
}

func generateBase58String(bytes []byte) string {
	encoding := base58.BitcoinEncoding
	encoded, err := encoding.Encode(bytes)

	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	return string(encoded)
}

func GenerateShortUrl(initialLink string, userId string) string {
	urlHashBytes := generateSha256Bytes(initialLink + userId)
	generatedNumber := new(big.Int).SetBytes(urlHashBytes).Uint64()

	result := generateBase58String([]byte(fmt.Sprintf("%d", generatedNumber)))

	return result[:8]
}
