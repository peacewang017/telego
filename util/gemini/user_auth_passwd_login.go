package gemini

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

var PasswdLoginEndpoint = "/user/login"

// // PasswdLoginRequest -> http.request
type PasswdLoginRequest struct {
	Header PasswdLoginRequestHeader
	Body   PasswdLoginRequestBody
}

type PasswdLoginRequestHeader struct {
	Authorization  string
	SpaceId        string
	TraceId        string
	AcceptLanguage string
}

type PasswdLoginRequestBody struct {
	UserName string `json:"userName"` // must have
	Password string `json:"password"` // must have
}

// // http.response -> PasswdLoginResponse
type PasswdLoginResponse struct {
	Body PasswdLoginResponseBody
}

type PasswdLoginResponseBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Token           string   `json:"token"`
		RefreshToken    string   `json:"refreshToken"`
		TokenExpireTime string   `json:"tokenExpireTime"`
		UserID          int      `json:"userId"`
		LoginMethod     string   `json:"loginMethod"`
		UserName        string   `json:"userName"`
		DisplayName     string   `json:"displayName"`
		Phone           string   `json:"phone"`
		Email           string   `json:"email"`
		PermissionList  []string `json:"permissionList"`
		Roles           []string `json:"roles"`
		RoleIDs         []int    `json:"roleIds"`
		AgreeSLA        int      `json:"agreeSLA"`
	} `json:"data"`
}

// implement Request

func (req *PasswdLoginRequest) ToHttpRequest(baseURL string) (*http.Request, error) {
	// fullURL
	fullURL := baseURL + UserAuthEndpoint + PasswdLoginEndpoint

	// Body
	if req.Body.Password == "" {
		return nil, fmt.Errorf("PasswdLoginRequest.httpRequest: necessary field empty: Body.Password")
	}

	if req.Body.UserName == "" {
		return nil, fmt.Errorf("PasswdLoginRequest.httpRequest: necessary field empty: Body.Username")
	}

	jsonBody, err := json.Marshal(req.Body)
	if err != nil {
		return nil, fmt.Errorf("PasswdLoginRequest.httpRequest: error converting to json body")
	}

	httpReq, err := http.NewRequest("POST", fullURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("PasswdLoginRequest.httpRequest: error creating new http request")
	}

	// Header
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authorization", req.Header.Authorization)
	httpReq.Header.Set("spaceId", req.Header.SpaceId)
	httpReq.Header.Set("traceId", req.Header.TraceId)
	httpReq.Header.Set("accept-language", req.Header.AcceptLanguage)

	return httpReq, nil
}

func (req *PasswdLoginRequest) GetEmptyResponse() Response {
	return &PasswdLoginResponse{}
}

// implement Response

func (resp *PasswdLoginResponse) FromHttpRequest(httpResp *http.Response) error {
	// Body
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("PasswdLoginRequest.FromHttpRequest: http response status %d", httpResp.StatusCode)
	}

	defer httpResp.Body.Close()
	decoder := json.NewDecoder(httpResp.Body)
	err := decoder.Decode(&resp.Body)
	if err != nil {
		return fmt.Errorf("PasswdLoginRequest.FromHttpRequest: error decoding body")
	}
	return nil
}
