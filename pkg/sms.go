package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

type nopCloser struct {
	io.Reader
}

func SendSms(phone string, message string) error {
	client := &http.Client{}
	formData := url.Values{}
	formData.Add("mobile_phone", phone)
	formData.Add("message", message)
	formData.Add("from", "4546")
	encodedFormData := formData.Encode()
	formDataReader := strings.NewReader(encodedFormData)
	readerWithCloser := nopCloser{formDataReader}

	filepathh, _ := filepath.Abs("./auth.json")

	fmt.Println(filepathh)

	var token SmsToken
	jsonFile, err := os.Open(filepathh)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(jsonFile)
	_ = decoder.Decode(&token)

	req, err := http.NewRequest("POST", "https://notify.eskiz.uz/api/message/sms/send", readerWithCloser)
	if err != nil {
		fmt.Println("Error creating request:", err)
		return err
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Authorization", "Bearer "+token.Token)
	resp, err := client.Do(req)
	if err != nil {
		log.Println("Error sending request:  ", err.Error())
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 401 {
		err = tokenGenerate()
		if err != nil {
			return err
		}
		err = SendSms(phone, message)
		if err != nil {
			return err
		}
	}

	defer resp.Body.Close()

	if resp.StatusCode == 400 {
		return errors.New("bad request")
	}

	if resp.StatusCode == 200 {
		return nil
	} else {
		return err
	}
}

func tokenGenerate() error {
	fmt.Println("here token")
	client := &http.Client{}
	formData := url.Values{}
	formData.Set("email", "jalolovdavlatbek@gmail.com")
	formData.Set("password", "BFlM2obX8dF5N5WjcNmDA0I7G2DJequc87YCHBhw")
	encodedFormData := formData.Encode()
	formDataReader := strings.NewReader(encodedFormData)
	readerWithCloser := nopCloser{formDataReader}

	req, err := http.NewRequest("POST", "https://notify.eskiz.uz/api/auth/login", readerWithCloser)
	if err != nil {
		log.Println("Error sending request: ", err.Error())
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req) // send request and returns response
	if err != nil {
		fmt.Println("Error sending request:", err)
		return err
	}

	var responseData map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		log.Println("error decoding json:", err)
		return err
	}
	token := responseData["data"].(map[string]interface{})["token"].(string)

	newToken := SmsToken{
		Token: token,
	}

	jsonData, err := json.Marshal(newToken)
	if err != nil {
		return err
	}

	err = os.WriteFile("auth.json", jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

type SmsToken struct {
	Token string `json:"token"`
}
