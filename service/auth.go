package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"task/api/models"
	"task/config"
	"task/pkg"
	"task/pkg/jwt"
	"task/pkg/logger"

	"task/storage"
	"time"

	"github.com/go-redis/redis"
)

type authService struct {
	storage storage.IStorage
	log     logger.LoggerI
	redis   storage.IRedisStorage
}

func NewAuthService(storage storage.IStorage, log logger.LoggerI, redis storage.IRedisStorage) authService {
	return authService{
		storage: storage,
		log:     log,
		redis:   redis,
	}
}

func (a authService) OTPForChangingNumber(ctx context.Context, loginRequest models.UserLoginRequest, id string) (string, error) {
	fmt.Println(" loginRequest.Login: ", loginRequest.MobilePhone)

	otpCode := pkg.GenerateOTP()

	msg := fmt.Sprintf("telefon raqamni yangilash uchun tasdiqlash kodi: %v", otpCode)

	err := a.redis.SetX(ctx, loginRequest.MobilePhone, otpCode, time.Minute*2)
	if err != nil {
		a.log.Error("error while setting otpCode to redis user for updating phone number", logger.Error(err))
		return "", err
	}

	return msg, nil
}

func (a authService) ConfirmOTPAndUpdatePhoneNumber(ctx context.Context, phoneNumber string, otp string, userID string) error {
	// Retrieve OTP from Redis
	storedOTP, err := a.redis.Get(ctx, phoneNumber)
	if err != nil {
		a.log.Error("Error retrieving OTP from Redis", logger.Error(err))
		return errors.New("Invalid or expired OTP")
	}

	// Check if OTP matches
	if storedOTP != otp {
		a.log.Error("OTP does not match")
		return errors.New("Invalid OTP")
	}

	// Update phone number in the database
	resp, err := a.storage.User().UpdatePhoneNumber(ctx, userID, phoneNumber)
	if err != nil {
		a.log.Error("Error updating phone number in the database", logger.Error(err))
		return err
	}

	// Successfully updated phone number
	a.log.Info(resp)
	return nil
}

func (a authService) UserLoginSendOTP(ctx context.Context, loginRequest models.UserLoginRequest) (string, error) {
	fmt.Println(" loginRequest.Login: ", loginRequest.MobilePhone)

	otpCode := pkg.GenerateOTP()

	msg := fmt.Sprintf("login uchun tasdiqlash kodi: %v", otpCode)

	err := a.redis.SetX(ctx, loginRequest.MobilePhone, otpCode, time.Minute*2)
	if err != nil {
		a.log.Error("error while setting otpCode to redis user register", logger.Error(err))
		return "", err
	}

	// err = pkg.SendSms(loginRequest.MobilePhone, msg)
	// if err != nil {
	// 	a.log.Error("error while sending otp code to user register", logger.Error(err))
	// 	return err
	// }
	return msg, nil
}

func (a authService) UserRegister(ctx context.Context, loginRequest models.UserRegisterRequest) (string, error) {
	fmt.Println("loginRequest.MobilePhone: ", loginRequest.MobilePhone)

	otpCode := pkg.GenerateOTP()
	otpStr := fmt.Sprintf("%d", otpCode) // Convert OTP to string for consistency

	msg := fmt.Sprintf("ro‘yxatdan o‘tish uchun tasdiqlash kodi: %v", otpCode)

	// Create a temporary user object to store in Redis
	tempUser := struct {
		OTP         string `json:"otp"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}{
		OTP:         otpStr, // Use the string OTP
		FirstName:   loginRequest.User.FirstName,
		LastName:    loginRequest.User.LastName,
		PhoneNumber: loginRequest.MobilePhone,
	}

	// Serialize to JSON
	tempUserData, err := json.Marshal(tempUser)
	if err != nil {
		a.log.Error("error while marshalling temp user data", logger.Error(err))
		return "", err
	}

	// Save to Redis with OTP as a string
	err = a.redis.SetX(ctx, loginRequest.MobilePhone, tempUserData, time.Minute*2)
	if err != nil {
		a.log.Error("error while setting temp user data to redis", logger.Error(err))
		return "", err
	}

	// Optionally send the OTP via SMS (if needed)
	// err = pkg.SendSms(loginRequest.MobilePhone, msg)
	// if err != nil {
	//     a.log.Error("error while sending otp code to user register", logger.Error(err))
	//     return "", err
	// }

	return msg, nil
}

// service/auth.go

func (a authService) UserRegisterConfirm(ctx context.Context, req models.UserRegisterConfRequest) (models.UserLoginResponse, error) {
	resp := models.UserLoginResponse{}

	// Retrieve OTP data from Redis
	otpData, err := a.redis.Get(ctx, req.MobilePhone)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			a.log.Error("otp not found in Redis", logger.String("mobile_phone", req.MobilePhone))
			return resp, fmt.Errorf("otp not found for phone number %s", req.MobilePhone)
		}
		a.log.Error("error while getting otp code for user register confirm", logger.Error(err))
		return resp, fmt.Errorf("error retrieving OTP data: %v", err)
	}

	otpDataStr, ok := otpData.(string)
	if !ok {
		a.log.Error("invalid OTP data format", logger.String("type", fmt.Sprintf("%T", otpData)))
		return resp, fmt.Errorf("invalid OTP data format")
	}

	var tempUser struct {
		OTP         string `json:"otp"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}
	err = json.Unmarshal([]byte(otpDataStr), &tempUser)
	if err != nil {
		a.log.Error("error while unmarshalling otp data", logger.Error(err))
		return resp, fmt.Errorf("failed to parse OTP data")
	}

	if req.Otp != tempUser.OTP {
		a.log.Error("incorrect otp code for user register confirm", logger.String("mobile_phone", req.MobilePhone))
		return resp, fmt.Errorf("incorrect otp code")
	}

	// Now create the user in the database
	user := models.User{
		FirstName:   tempUser.FirstName,
		LastName:    tempUser.LastName,
		PhoneNumber: tempUser.PhoneNumber,
	}

	// Insert the user into the database and get the user ID
	id, err := a.storage.User().Create(ctx, &user)
	if err != nil {
		a.log.Error("error while creating user", logger.Error(err))
		return resp, fmt.Errorf("failed to create user: %v", err)
	}

	// Now that the user is created, create the device
	device := models.Device{
		UserID:          id.ID, // Use the user ID here
		Name:            req.DeviceInfo.Name,
		NotificationKey: req.DeviceInfo.NotificationKey,
		Type:            req.DeviceInfo.Type,
		OsVersion:       req.DeviceInfo.OsVersion,
		AppVersion:      req.DeviceInfo.AppVersion,
		RememberMe:      req.DeviceInfo.RememberMe,
		AdId:            req.DeviceInfo.AdId,
	}

	// Insert the device into the database
	deviceID, err := a.storage.Device().Insert(ctx, &device)
	if err != nil {
		a.log.Error("error while inserting device", logger.Error(err))
		return resp, fmt.Errorf("failed to insert device: %v", err)
	}

	// Create JWT tokens for the user
	m := make(map[string]interface{})
	m["user_id"] = id.ID
	m["user_role"] = config.USER_ROLE
	m["device_id"] = deviceID.ID

	accessToken, refreshToken, err := jwt.GenJWT(m)
	if err != nil {
		a.log.Error("error while generating tokens for user register confirm", logger.Error(err))
		return resp, fmt.Errorf("failed to generate tokens: %v", err)
	}

	// Set the generated tokens in the response
	resp.AccessToken = accessToken
	resp.RefreshToken = refreshToken

	return resp, nil
}

func (a authService) UserLoginByPhoneConfirm(ctx context.Context, req models.UserLoginPhoneConfirmRequest) (models.UserLoginResponse, error) {
	resp := models.UserLoginResponse{}

	// Retrieve the OTP from Redis for the provided phone number
	storedOTP, err := a.redis.Get(ctx, req.MobilePhone)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			a.log.Error("OTP code not found or expired", logger.Error(err))
			return resp, errors.New("OTP kod topilmadi yoki muddati tugagan")
		}
		a.log.Error("error while getting OTP code from redis", logger.Error(err))
		return resp, errors.New("tizim xatosi yuz berdi")
	}

	// Validate the OTP provided by the user
	if req.SmsCode != storedOTP {
		a.log.Error("incorrect OTP code", logger.Error(errors.New("OTP code mismatch")))
		return resp, errors.New("noto'g'ri OTP kod")
	}

	// Delete OTP from Redis after successful verification
	err = a.redis.Del(ctx, req.MobilePhone)
	if err != nil {
		a.log.Error("error while deleting OTP from redis", logger.Error(err))
		return resp, err
	}

	// Check if the phone number exists and get the user ID
	id, err := a.storage.User().CheckPhoneNumberExist(ctx, req.MobilePhone)
	if err != nil {
		a.log.Error("error while getting user by phone number", logger.Error(err))
		return resp, err
	}

	// Create a device entry for the user (new device if not already registered)
	device := models.Device{
		UserID:          id.ID,
		Name:            req.DeviceInfo.Name,
		NotificationKey: req.DeviceInfo.NotificationKey,
		Type:            req.DeviceInfo.Type,
		OsVersion:       req.DeviceInfo.OsVersion,
		AppVersion:      req.DeviceInfo.AppVersion,
		RememberMe:      req.DeviceInfo.RememberMe,
		AdId:            req.DeviceInfo.AdId,
	}

	// Insert the device into the database and get the device ID
	deviceID, err := a.storage.Device().Insert(ctx, &device)
	if err != nil {
		a.log.Error("error while inserting device: " + err.Error())
		return resp, err
	}

	// Prepare the JWT payload
	m := make(map[string]interface{})
	m["user_id"] = id.ID
	m["user_role"] = config.USER_ROLE
	m["device_id"] = deviceID.ID // Add device_id to the payload

	// Generate the JWT access and refresh tokens
	accessToken, refreshToken, err := jwt.GenJWT(m)
	if err != nil {
		a.log.Error("error while generating tokens for user register confirm", logger.Error(err))
		return resp, err
	}

	// Set the generated tokens in the response
	resp.AccessToken = accessToken
	resp.RefreshToken = refreshToken

	return resp, nil
}

/*
func (a authService) UserLogin(ctx context.Context, loginRequest models.UserLoginRequest) (models.UserLoginResponse, error) {
	resp := models.UserLoginResponse{}
	fmt.Println(" loginRequest.Login: ", loginRequest.MobilePhone)
	user, err := a.storage.User().GetByLogin(ctx, loginRequest.MobilePhone)
	if err != nil {
		a.log.Error("error while getting user credentials by login", logger.Error(err))
		return models.UserLoginResponse{}, err
	}

	otpData, err := a.redis.Get(ctx, loginRequest.MobilePhone)
	if err != nil {
		a.log.Error("error while getting otp code for user login confirm", logger.Error(err))
		return resp, err
	}

	if otpData != loginRequest.Otp {
		a.log.Error("otp code mismatch", logger.Error(err))
		return resp, err
	}

	m := make(map[interface{}]interface{})

	m["user_id"] = user.ID
	m["user_role"] = config.USER_ROLE

	accessToken, refreshToken, err := jwt.GenJWT(m)
	if err != nil {
		a.log.Error("error while generating tokens for user login", logger.Error(err))
		return models.UserLoginResponse{}, err
	}

	return models.UserLoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}
*/
