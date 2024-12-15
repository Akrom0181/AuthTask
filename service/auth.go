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
	storedOTP, err := a.redis.Get(ctx, phoneNumber)
	if err != nil {
		a.log.Error("Error retrieving OTP from Redis", logger.Error(err))
		return errors.New("invalid or expired OTP")
	}

	if storedOTP != otp {
		a.log.Error("OTP does not match")
		return errors.New("invalid OTP")
	}

	resp, err := a.storage.User().UpdatePhoneNumber(ctx, userID, phoneNumber)
	if err != nil {
		a.log.Error("Error updating phone number in the database", logger.Error(err))
		return err
	}

	a.log.Info(resp)
	return nil
}

func (a authService) UserRegister(ctx context.Context, loginRequest models.UserRegisterRequest) (string, error) {
	fmt.Println("loginRequest.MobilePhone: ", loginRequest.MobilePhone)

	// Generate OTP
	otpCode := pkg.GenerateOTP()
	otpStr := fmt.Sprintf("%d", otpCode)

	identifier := pkg.GenerateIdentifier()

	msg := fmt.Sprintf("Ro‘yxatdan o‘tish uchun tasdiqlash kodi: %v  Identifikator: %v", otpCode, identifier)

	tempUser := struct {
		OTP         string `json:"otp"`
		Identifier  string `json:"identifier"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}{
		OTP:         otpStr,
		Identifier:  identifier,
		FirstName:   loginRequest.User.FirstName,
		LastName:    loginRequest.User.LastName,
		PhoneNumber: loginRequest.MobilePhone,
	}

	tempUserData, err := json.Marshal(tempUser)
	if err != nil {
		a.log.Error("error while marshalling temp user data", logger.Error(err))
		return "", err
	}

	err = a.redis.SetX(ctx, loginRequest.MobilePhone, tempUserData, time.Minute*3)
	if err != nil {
		a.log.Error("error while setting temp user data to redis", logger.Error(err))
		return "", err
	}

	a.log.Info("User registration initiated",
		logger.String("phone_number", loginRequest.MobilePhone),
		logger.String("identifier", identifier),
	)

	return msg, nil
}

func (a authService) UserRegisterConfirm(ctx context.Context, req models.UserRegisterConfRequest) (models.UserLoginResponse, error) {
	resp := models.UserLoginResponse{}

	otpData, err := a.redis.Get(ctx, req.MobilePhone)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			a.log.Error("OTP not found or expired in Redis", logger.String("mobile_phone", req.MobilePhone))
			return resp, fmt.Errorf(config.ErrOtpExpired)
		}
		a.log.Error("Error while getting OTP data from Redis", logger.Error(err))
		return resp, fmt.Errorf("error retrieving OTP data: %v", err)
	}

	otpDataStr, ok := otpData.(string)
	if !ok {
		a.log.Error("Invalid OTP data format", logger.String("type", fmt.Sprintf("%T", otpData)))
		return resp, fmt.Errorf(config.ErrOtpInvalidFormat)
	}

	var tempUser struct {
		OTP         string `json:"otp"`
		Identifier  string `json:"identifier"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}

	err = json.Unmarshal([]byte(otpDataStr), &tempUser)
	if err != nil {
		a.log.Error("Error unmarshalling OTP data", logger.Error(err))
		return resp, fmt.Errorf("failed to parse OTP data")
	}

	// Validate OTP
	if req.Otp != tempUser.OTP {
		a.log.Error("Incorrect OTP code", logger.String("mobile_phone", req.MobilePhone))
		return resp, fmt.Errorf(config.ErrOtpMismatch)
	}

	// Validate Identifier (if used)
	if req.Identifier != tempUser.Identifier {
		a.log.Error("Identifier mismatch", logger.String("mobile_phone", req.MobilePhone))
		return resp, fmt.Errorf("identifier mismatch")
	}

	// Create User
	user := models.User{
		FirstName:   tempUser.FirstName,
		LastName:    tempUser.LastName,
		PhoneNumber: tempUser.PhoneNumber,
	}

	id, err := a.storage.User().Create(ctx, &user)
	if err != nil {
		a.log.Error(config.ErrUserCreation, logger.Error(err), logger.String("mobile_phone", req.MobilePhone))
		return resp, fmt.Errorf(config.ErrUserCreation)
	}

	// Insert Device
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

	deviceID, err := a.storage.Device().Insert(ctx, &device)
	if err != nil {
		a.log.Error(config.ErrDeviceInsertion, logger.Error(err))
		rollbackErr := a.storage.User().Delete(ctx, id.ID)
		if rollbackErr != nil {
			a.log.Error("Error rolling back user creation", logger.Error(rollbackErr), logger.String("user_id", id.ID))
		}
		return resp, fmt.Errorf(config.ErrDeviceInsertion)
	}

	// Generate Tokens
	m := map[string]interface{}{
		"user_id":   id.ID,
		"user_role": config.USER_ROLE,
		"device_id": deviceID.ID,
	}

	accessToken, refreshToken, err := jwt.GenJWT(m)
	if err != nil {
		a.log.Error(config.ErrTokenGeneration, logger.Error(err))
		return resp, fmt.Errorf(config.ErrTokenGeneration)
	}

	resp.AccessToken = accessToken
	resp.RefreshToken = refreshToken

	return resp, nil
}

func (a authService) UserLoginSendOTP(ctx context.Context, loginRequest models.UserLoginRequest) (string, error) {
	fmt.Println(" loginRequest.Login: ", loginRequest.MobilePhone)

	otpCode := pkg.GenerateOTP()
	identifier := pkg.GenerateIdentifier()

	msg := fmt.Sprintf("login uchun tasdiqlash kodi: %v  Identifikator: %v", otpCode, identifier)

	tempUser := struct {
		OTP        int    `json:"otp"`
		Identifier string `json:"identifier"`
	}{
		OTP:        otpCode,
		Identifier: identifier,
	}

	tempUserData, err := json.Marshal(tempUser)
	if err != nil {
		a.log.Error("error while marshalling temp user data", logger.Error(err))
		return "", err
	}

	err = a.redis.SetX(ctx, loginRequest.MobilePhone, tempUserData, time.Minute*3)
	if err != nil {
		a.log.Error("error while setting otpCode to redis user login", logger.Error(err))
		return "", err
	}

	// err = pkg.SendSms(loginRequest.MobilePhone, msg)
	// if err != nil {
	// 	a.log.Error("error while sending otp code to user register", logger.Error(err))
	// 	return err
	// }

	return msg, nil
}

func (a authService) UserLoginByPhoneConfirm(ctx context.Context, req models.UserLoginPhoneConfirmRequest) (models.UserLoginResponse, error) {
	resp := models.UserLoginResponse{}

	storedOTP, err := a.redis.Get(ctx, req.MobilePhone)
	if err != nil {
		if errors.Is(err, redis.Nil) {
			a.log.Error("OTP code not found or expired", logger.Error(err))
			return resp, errors.New("OTP kod topilmadi yoki muddati tugagan")
		}
		a.log.Error("error while getting OTP code from redis", logger.Error(err))
		return resp, errors.New("tizim xatosi yuz berdi")
	}

	otpDataStr, ok := storedOTP.(string)
	if !ok {
		a.log.Error("Invalid OTP data format", logger.String("type", fmt.Sprintf("%T", otpDataStr)))
		return resp, fmt.Errorf(config.ErrOtpInvalidFormat)
	}

	var tempUser struct {
		OTP         string `json:"otp"`
		Identifier  string `json:"identifier"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}

	err = json.Unmarshal([]byte(otpDataStr), &tempUser)
	if err != nil {
		a.log.Error("Error unmarshalling OTP data", logger.Error(err))
		return resp, fmt.Errorf("failed to parse OTP data")
	}

	if req.SmsCode != tempUser.OTP {
		a.log.Error("incorrect OTP code", logger.Error(errors.New("OTP code mismatch")))
		return resp, errors.New("noto'g'ri OTP kod")
	}

	if req.Identifier != tempUser.Identifier {
		a.log.Error("identifier mismatch", logger.Error(errors.New("identifier mismatch")))
		return resp, errors.New("noto'g'ri identifikator")
	}

	err = a.redis.Del(ctx, req.MobilePhone)
	if err != nil {
		a.log.Error("error while deleting OTP from redis", logger.Error(err))
		return resp, err
	}

	id, err := a.storage.User().CheckPhoneNumberExist(ctx, req.MobilePhone)
	if err != nil {
		a.log.Error("error while getting user by phone number", logger.Error(err))
		return resp, err
	}

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

	deviceID, err := a.storage.Device().Insert(ctx, &device)
	if err != nil {
		a.log.Error("error while inserting device: " + err.Error())
		return resp, err
	}

	m := make(map[string]interface{})
	m["user_id"] = id.ID
	m["user_role"] = config.USER_ROLE
	m["device_id"] = deviceID.ID

	accessToken, refreshToken, err := jwt.GenJWT(m)
	if err != nil {
		a.log.Error("error while generating tokens for user register confirm", logger.Error(err))
		return resp, err
	}

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
