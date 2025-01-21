package gemini

import (
	"encoding/json"
	"fmt"
	"net/http"
)

var UserJoinedSpaceEndpoint = "/admin/user/join/list"

var _ Request = &UserJoinedSpaceRequest{}

type UserJoinedSpaceRequest struct {
	Header UserJoinedSpaceRequestHeader
}

type UserJoinedSpaceRequestHeader struct {
	Authorization  string `json:"authorization"`
	SpaceId        string `json:"spaceId"`
	TraceId        string `json:"traceId"`
	AcceptLanguage string `json:"accept-language"`
}

var _ Response = &UserJoinedSpaceResponse{}

type UserJoinedSpaceResponse struct {
	Body UserJoinedSpaceResponseBody
}

type UserJoinedSpaceResponseBody struct {
	Code int    `json:"code"` // 0代表成功，非0代表失败
	Msg  string `json:"msg"`  // 返回信息
	Data struct {
		SpaceList []struct {
			SpaceId     string `json:"spaceId"`     // 空间ID
			SpaceName   string `json:"spaceName"`   // 空间名称
			Description string `json:"description"` // 空间描述
		} `json:"spaceList"`
	} `json:"data"`
}

// implement Request

func (req *UserJoinedSpaceRequest) ToHttpRequest(baseURL string) (*http.Request, error) {
	// fullURL
	fullURL := baseURL + WebAPIEndpoint + UserJoinedSpaceEndpoint

	// httpReq
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("UserJoinedSpaceRequest.ToHttpRequest: error creating new http request")
	}

	// Header
	if req.Header.Authorization == "" {
		return nil, fmt.Errorf("UserJoinedSpaceRequest.ToHttpRequest: necessary field empty: Header.Authorization")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authorization", req.Header.Authorization)
	httpReq.Header.Set("spaceId", req.Header.SpaceId)
	httpReq.Header.Set("traceId", req.Header.TraceId)
	httpReq.Header.Set("accept-language", req.Header.AcceptLanguage)

	return httpReq, nil
}

func (req *UserJoinedSpaceRequest) GetEmptyResponse() Response {
	return &UserJoinedSpaceResponse{}
}

func (resp *UserJoinedSpaceResponse) FromHttpRequest(httpResp *http.Response) error {
	// Body
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("UserJoinedSpaceResponse.FromHttpRequest: http response status %d", httpResp.StatusCode)
	}

	defer httpResp.Body.Close()
	decoder := json.NewDecoder(httpResp.Body)
	err := decoder.Decode(&resp.Body)
	if err != nil {
		return fmt.Errorf("UserJoinedSpaceResponse.FromHttpRequest: error decoding body")
	}
	return nil
}
