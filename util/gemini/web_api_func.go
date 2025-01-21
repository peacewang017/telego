package gemini

import "fmt"

var WebAPIEndpoint = "/gemini/v1/gemini_api/gemini_api"

// Gemini http API 包装函数 - WebAPI
type WebAPIFuncStruct struct {
	server *GeminiServer
}

// 获取所有空间列表
func (fun *WebAPIFuncStruct) SpaceList(req *SpaceListRequest) (*SpaceListResponse, error) {
	resp, err := fun.server.httpAct(req)
	if err != nil {
		return nil, fmt.Errorf("WebAPIFunc.SpaceList: %w", err)
	}

	spaceListResp, ok := resp.(*SpaceListResponse)
	if !ok {
		return nil, fmt.Errorf("WebAPIFunc.SpaceList: unexpected response type: %T", resp)
	}

	return spaceListResp, nil
}

// 获取用户加入的空间列表
func (fun *WebAPIFuncStruct) UserJoinedSpace(req *UserJoinedSpaceRequest) (*UserJoinedSpaceResponse, error) {
	resp, err := fun.server.httpAct(req)
	if err != nil {
		return nil, fmt.Errorf("WebAPIFunc.UserJoinedSpace: %w", err)
	}

	userJoinedSpaceResp, ok := resp.(*UserJoinedSpaceResponse)
	if !ok {
		return nil, fmt.Errorf("WebAPIFunc.UserJoinedSpace: unexpected response type: %T", resp)
	}

	return userJoinedSpaceResp, nil
}
