package check

import (
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"time"
)

func ValidateCarYear(year int) error {
	if year <= 0 || year > time.Now().Year()+1 {
		return errors.New("year is not valid")
	}
	return nil
}

func ValidateEmailAddress(email string) error {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

	if !emailRegex.MatchString(email) {
		return fmt.Errorf("email address %s is not valid", email)
	}

	return nil
}

func ValidatePhoneNumber(phoneNumber string) error {
	starting := `^\+998\d{9}$`
	correct, err := regexp.MatchString(starting, phoneNumber)
	if err != nil {
		return fmt.Errorf("error while validation: %v", err)
	}
	if !correct {
		return fmt.Errorf("your phone number %s is invalid", phoneNumber)
	}
	return nil
}

func IsValidPassword(password string) error {
	regex := regexp.MustCompile(`^.{8,}$`)
	if !regex.MatchString(password) {
		return errors.New("password must be at least 8 characters long")
	}
	return nil
}

func CheckEmailExists(email string) error {
	// Your Hunter.io API Key (you need to replace this with your real API key)
	apiKey := "8c39fac60c281c47d47cbb11a6af5b78451c4534"
	url := fmt.Sprintf("https://api.hunter.io/v2/email-verifier?email=%s&api_key=%s", email, apiKey)

	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	fmt.Println(string(body))

	return nil
}
