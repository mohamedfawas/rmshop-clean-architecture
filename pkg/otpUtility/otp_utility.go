package otputil

import (
	"crypto/rand" // provides functions to generate cryptographically secure random numbers
	"log"
	"math/big" //used to handle the random number generation for selecting digits
)

func GenerateOTP(length int) (string, error) {
	const digits = "0123456789" //all possible cahracters that can be used in the OTP
	result := make([]byte, length)

	//A loop runs length times to generate each digit of the OTP
	for i := 0; i < length; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits)))) //generates a cryptographically secure random integer num between 0 and 9
		if err != nil {
			log.Printf("error generating each digit of otp : %v", err)
			return "", err
		}
		result[i] = digits[num.Int64()] //randomly generated number num is used as an index to pick a digit from the digits string
	}
	return string(result), nil
}
