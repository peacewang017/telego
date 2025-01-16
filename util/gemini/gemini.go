package gemini

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"strings"
)

type GeminiServer struct {
	BaseURL string

	UserAuthFunc *UserAuthFuncStruct
	WebAPIFunc   *WebAPIFuncStruct
}

func NewGeminiServer(BaseURL string) (*GeminiServer, error) {
	if !strings.HasPrefix(BaseURL, "http://") && !strings.HasPrefix(BaseURL, "https://") {
		return nil, fmt.Errorf("NewGeminiServer: invalid BaseURL: %s", BaseURL)
	}
	gServer := &GeminiServer{BaseURL: BaseURL}
	gServer.UserAuthFunc = &UserAuthFuncStruct{server: gServer}
	gServer.WebAPIFunc = &WebAPIFuncStruct{server: gServer}
	return gServer, nil
}

func (gServer *GeminiServer) httpAct(req Request) (Response, error) {
	// 跳过 TLS 验证
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// request -> httpRequest
	httpReq, err := req.ToHttpRequest(gServer.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("httpAct: %w", err)
	}

	// httpRequest -> httpResponse
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("httpAct: failed to send request: %w", err)
	}

	// httpResponse -> Response
	// // 反序列化 Body
	resp := req.GetEmptyResponse()
	err = resp.FromHttpRequest(httpResp)
	if err != nil {
		return nil, fmt.Errorf("httpAct: error get response: %w", err)
	}

	return resp, nil
}
