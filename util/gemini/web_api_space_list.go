package gemini

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

var SpaceListEndpoint = "/admin/space/list"

var _ Request = &SpaceListRequest{}

type SpaceListRequest struct {
	Query  SpaceListRequestQuery
	Header SpaceListRequestHeader
}

type SpaceListRequestQuery struct {
	KeyWords  string
	PageNum   string
	PageSize  string
	SpaceName string
}

type SpaceListRequestHeader struct {
	Authorization  string // must have
	SpaceID        string
	TraceID        string
	AcceptLanguage string
}

var _ Response = &SpaceListResponse{}

type SpaceListResponse struct {
	Body SpaceListResponseBody
}

type SpaceListResponseBody struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ListCount int `json:"listCount"`
		SpaceList []struct {
			SpaceID          string `json:"spaceId"`
			SpaceName        string `json:"spaceName"`
			Description      string `json:"description"`
			OwnerUserID      int    `json:"ownerUserId"`
			OwnerUserName    string `json:"ownerUserName"`
			OwnerDisplayName string `json:"ownerDisplayName"`
			CreateTime       string `json:"createTime"`
			MemberCount      int    `json:"memberCnt"`
			SpaceQuotaID     int    `json:"spacequotaId"`
			ResourceLabel    string `json:"resourceLabel"`
		} `json:"spaceList"`
	} `json:"data"`
}

// implement Request

func (req *SpaceListRequest) ToHttpRequest(baseURL string) (*http.Request, error) {
	// Query
	queryParams := url.Values{}
	if req.Query.KeyWords != "" {
		queryParams.Add("keyWords", req.Query.KeyWords)
	}
	if req.Query.PageNum != "" {
		queryParams.Add("pageNum", req.Query.PageNum)
	}
	if req.Query.PageSize != "" {
		queryParams.Add("pageSize", req.Query.PageSize)
	}
	if req.Query.SpaceName != "" {
		queryParams.Add("spaceName", req.Query.SpaceName)
	}

	// fullURL
	fullURL := baseURL + WebAPIEndpoint + SpaceListEndpoint
	if queryString := queryParams.Encode(); queryString != "" {
		fullURL += "?" + queryString
	}

	// httpReq
	httpReq, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("SpaceListRequest.httpRequest: error creating new http request")
	}

	// Header
	if req.Header.Authorization == "" {
		return nil, fmt.Errorf("SpaceListRequest.httpRequest: necessary field empty: Header.Authorization")
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("authorization", req.Header.Authorization)
	httpReq.Header.Set("spaceId", req.Header.SpaceID)
	httpReq.Header.Set("traceId", req.Header.TraceID)
	httpReq.Header.Set("accept-language", req.Header.AcceptLanguage)

	return httpReq, nil
}

func (req *SpaceListRequest) GetEmptyResponse() Response {
	return &SpaceListResponse{}
}

// implement Response

func (resp *SpaceListResponse) FromHttpRequest(httpResp *http.Response) error {
	// Body
	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("SpaceListResponse.FromHttpRequest: http response status %d", httpResp.StatusCode)
	}

	defer httpResp.Body.Close()
	decoder := json.NewDecoder(httpResp.Body)
	err := decoder.Decode(&resp.Body)
	if err != nil {
		return fmt.Errorf("SpaceListResponse.FromHttpRequest: error decoding body")
	}
	return nil
}
