package service

import (
	"fmt"
	"sanyuktgolang/auth"
	"sanyuktgolang/domain"
	"sanyuktgolang/errs"
	"sanyuktgolang/logger"
	"sanyuktgolang/model"

	"github.com/dgrijalva/jwt-go"
)

type AuthService interface {
	Login(model.LoginRequest) (*model.Response, *errs.AppError)
	GenerateOtp(model.LoginRequest) (*model.Response, *errs.AppError)
	VerifyOtp(model.LoginRequest) (*model.Response, *errs.AppError)
	Verify(urlParams map[string]string) *errs.AppError
	Refresh(request model.RefreshTokenRequest) (*model.Response, *errs.AppError)
}

type DefaultAuthService struct {
	repo            domain.AuthRepository
	rolePermissions domain.RolePermissions
}

func (s DefaultAuthService) Refresh(request model.RefreshTokenRequest) (*model.Response, *errs.AppError) {
	if vErr := request.IsAccessTokenValid(); vErr != nil {
		if vErr.Errors == jwt.ValidationErrorExpired {
			// continue with the refresh token functionality
			var appErr *errs.AppError
			if appErr = s.repo.RefreshTokenExists(request.RefreshToken); appErr != nil {
				return nil, appErr
			}
			// generate a access token from refresh token.
			var accessToken string
			if accessToken, appErr = domain.NewAccessTokenFromRefreshToken(request.RefreshToken); appErr != nil {
				return nil, appErr
			}
			return &model.Response{Data: model.LoginResponse{AccessToken: accessToken}, Status: true, Message: "Success"}, nil
		}
		return nil, errs.NewAuthenticationError("invalid token")
	}
	return nil, errs.NewAuthenticationError("cannot generate a new access token until the current one expires")
}

func (s DefaultAuthService) Login(req model.LoginRequest) (*model.Response, *errs.AppError) {
	var appErr *errs.AppError
	var login *domain.Login

	if login, appErr = s.repo.FindBy(req.Username, req.Password); appErr != nil {
		return nil, appErr
	}

	claims := login.ClaimsForAccessToken()
	authToken := domain.NewAuthToken(claims)

	var accessToken, refreshToken string
	if accessToken, appErr = authToken.NewAccessToken(); appErr != nil {
		return nil, appErr
	}

	if refreshToken, appErr = s.repo.GenerateAndSaveRefreshTokenToStore(authToken); appErr != nil {
		return nil, appErr
	}

	return &model.Response{Data: model.LoginResponse{AccessToken: accessToken, RefreshToken: refreshToken}, Status: true, Message: "Success"}, nil
}

func (s DefaultAuthService) GenerateOtp(req model.LoginRequest) (*model.Response, *errs.AppError) {
	var appErr *errs.AppError
	var login *domain.Users

	if login, appErr = s.repo.FindByMobile(req.Mobile); appErr != nil {
		return nil, appErr
	}
	fmt.Println(login)
	return &model.Response{Message: "An OTP has been sent to your given mobile number", Status: true}, nil
}

func (s DefaultAuthService) VerifyOtp(req model.LoginRequest) (*model.Response, *errs.AppError) {
	var appErr *errs.AppError
	var login *domain.Users

	if login, appErr = s.repo.VerifyOtp(req.Mobile, req.Otp); appErr != nil {
		return nil, appErr
	}
	fmt.Println(login)

	token, err := auth.CreateToken(uint32(login.Id))
	if err != nil {
		return nil, errs.NewAuthenticationError("not able to create token")
	}
	return &model.Response{Data: model.LoginResponse{AccessToken: token, RefreshToken: token}, Status: true, Message: "success"}, nil
}

func (s DefaultAuthService) Verify(urlParams map[string]string) *errs.AppError {
	// convert the string token to JWT struct
	if jwtToken, err := jwtTokenFromString(urlParams["token"]); err != nil {
		return errs.NewAuthorizationError(err.Error())
	} else {
		/*
		   Checking the validity of the token, this verifies the expiry
		   time and the signature of the token
		*/
		if jwtToken.Valid {
			// type cast the token claims to jwt.MapClaims
			claims := jwtToken.Claims.(*domain.AccessTokenClaims)
			/* if Role if user then check if the account_id and customer_id
			   coming in the URL belongs to the same token
			*/
			if claims.IsUserRole() {
				if !claims.IsRequestVerifiedWithTokenClaims(urlParams) {
					return errs.NewAuthorizationError("request not verified with the token claims")
				}
			}
			// verify of the role is authorized to use the route
			isAuthorized := s.rolePermissions.IsAuthorizedFor(claims.Role, urlParams["routeName"])
			if !isAuthorized {
				return errs.NewAuthorizationError(fmt.Sprintf("%s role is not authorized", claims.Role))
			}
			return nil
		} else {
			return errs.NewAuthorizationError("Invalid token")
		}
	}
}

func jwtTokenFromString(tokenString string) (*jwt.Token, error) {
	token, err := jwt.ParseWithClaims(tokenString, &domain.AccessTokenClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(domain.HMAC_SAMPLE_SECRET), nil
	})
	if err != nil {
		logger.Error("Error while parsing token: " + err.Error())
		return nil, err
	}
	return token, nil
}

func NewLoginService(repo domain.AuthRepository, permissions domain.RolePermissions) DefaultAuthService {
	return DefaultAuthService{repo, permissions}
}
