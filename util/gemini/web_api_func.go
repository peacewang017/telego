package gemini

import "fmt"

var WebAPIEndpoint = "/gemini/v1/gemini_api/gemini_api"

type WebAPIFuncStruct struct {
	server *GeminiServer
}

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
