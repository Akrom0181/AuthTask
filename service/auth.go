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
	otpStr := fmt.Sprintf("%d", otpCode)

	msg := fmt.Sprintf("ro‘yxatdan o‘tish uchun tasdiqlash kodi: %v", otpCode)

	tempUser := struct {
		OTP         string `json:"otp"`
		FirstName   string `json:"first_name"`
		LastName    string `json:"last_name"`
		PhoneNumber string `json:"phone_number"`
	}{
		OTP:         otpStr,
		FirstName:   loginRequest.User.FirstName,
		LastName:    loginRequest.User.LastName,
		PhoneNumber: loginRequest.MobilePhone,
	}

	tempUserData, err := json.Marshal(tempUser)
	if err != nil {
		a.log.Error("error while marshalling temp user data", logger.Error(err))
		return "", err
	}

	err = a.redis.SetX(ctx, loginRequest.MobilePhone, tempUserData, time.Minute*2)
	if err != nil {
		a.log.Error("error while setting temp user data to redis", logger.Error(err))
		return "", err
	}

	return msg, nil
}

func (a authService) UserRegisterConfirm(ctx context.Context, req models.UserRegisterConfRequest) (models.UserLoginResponse, error) {
	resp := models.UserLoginResponse{}

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

	user := models.User{
		FirstName:   tempUser.FirstName,
		LastName:    tempUser.LastName,
		PhoneNumber: tempUser.PhoneNumber,
	}

	id, err := a.storage.User().Create(ctx, &user)
	if err != nil {
		a.log.Error("error while creating user", logger.Error(err))
		return resp, fmt.Errorf("failed to create user: %v", err)
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
		a.log.Error("error while inserting device", logger.Error(err))

		rollbackErr := a.storage.User().Delete(ctx, id.ID)
		if rollbackErr != nil {
			a.log.Error("error while rolling back user creation", logger.Error(rollbackErr))
		}

		return resp, fmt.Errorf("failed to insert device: %v", err)
	}

	m := make(map[string]interface{})
	m["user_id"] = id.ID
	m["user_role"] = config.USER_ROLE
	m["device_id"] = deviceID.ID

	accessToken, refreshToken, err := jwt.GenJWT(m)
	if err != nil {
		a.log.Error("error while generating tokens for user register confirm", logger.Error(err))
		return resp, fmt.Errorf("failed to generate tokens: %v", err)
	}

	resp.AccessToken = accessToken
	resp.RefreshToken = refreshToken

	return resp, nil
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

	if req.SmsCode != storedOTP {
		a.log.Error("incorrect OTP code", logger.Error(errors.New("OTP code mismatch")))
		return resp, errors.New("noto'g'ri OTP kod")
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
