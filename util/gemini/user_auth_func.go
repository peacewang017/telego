package gemini

import (
	"fmt"
)

var UserAuthEndpoint = "/gemini/v1/gemini_userauth"

// Gemini http API 包装函数 - UserAuthAPI
type UserAuthFuncStruct struct {
	server *GeminiServer
}

func (fun *UserAuthFuncStruct) PasswdLogin(req *PasswdLoginRequest) (*PasswdLoginResponse, error) {
	resp, err := fun.server.httpAct(req)
	if err != nil {
		return nil, fmt.Errorf("UserAuthFunc.PasswdLogin: %w", err)
	}

	passwdLoginResp, ok := resp.(*PasswdLoginResponse)
	if !ok {
		return nil, fmt.Errorf("UserAuthFunc.PasswdLogin: unexpected response type: %T", resp)
	}

	return passwdLoginResp, nil
}
