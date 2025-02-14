package client

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/tomnomnom/linkheader"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

const (
	baseURL = "https://api.freshbooks.com/auth/api/v1/businesses/"

	getTeamMembers = "/team_members"

	getBussinessID = "https://api.freshbooks.com/auth/api/v1/users/me"
)

type FreshBooksClient struct {
	client     *uhttp.BaseHttpClient
	businessID string
	token      string
}

func (f *FreshBooksClient) WithBearerToken(apiToken string) *FreshBooksClient {
	f.token = apiToken
	return f
}

func (f *FreshBooksClient) WithBusinessID(businessID string) *FreshBooksClient {
	f.businessID = businessID
	return f
}

func NewClient() *FreshBooksClient {
	return &FreshBooksClient{
		client:     &uhttp.BaseHttpClient{},
		token:      "",
		businessID: "",
	}
}

func (f *FreshBooksClient) GetBusinessID() string {
	return f.businessID
}

func (f *FreshBooksClient) SetBusinessID(bid int64) {
	f.businessID = strconv.FormatInt(bid, 10)
}

func (f *FreshBooksClient) GetToken() string {
	return f.token
}

func New(ctx context.Context) (*FreshBooksClient, error) {
	clientToken := ""
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	cli, err := uhttp.NewBaseHttpClientWithContext(context.Background(), httpClient)
	if err != nil {
		return nil, err
	}

	fbClient := FreshBooksClient{
		client:     cli,
		token:      clientToken,
		businessID: "",
	}

	return &fbClient, nil
}

// getListFromAPI sends a request to the Freshdesk API to receive a JSON with a list of entities.
func (f *FreshBooksClient) getListFromAPI(
	ctx context.Context,
	urlAddress string,
	res any,
	reqOpt ...ReqOpt,
) (string, annotations.Annotations, error) {
	header, annotation, err := f.doRequest(ctx, http.MethodGet, urlAddress, &res, nil, reqOpt...)

	if err != nil {
		return "", nil, err
	}

	var pageToken string
	pagingLinks := linkheader.Parse(header.Get("Link"))
	for _, link := range pagingLinks {
		if link.Rel == "next" {
			nextPageUrl, err := url.Parse(link.URL)
			if err != nil {
				return "", nil, err
			}
			pageToken = nextPageUrl.Query().Get("page")
			break
		}
	}

	return pageToken, annotation, nil
}

// ListTeamMembers Gets all the Team Members from FreshBooks and deserialized them into an Array.
func (f *FreshBooksClient) ListTeamMembers(ctx context.Context, opts PageOptions) ([]TeamMember, string, annotations.Annotations, error) {
	queryUrl, err := url.JoinPath(baseURL, f.GetBusinessID(), getTeamMembers)
	if err != nil {
		return nil, "", nil, err
	}

	var res Response
	nextPage, annotation, err := f.getListFromAPI(ctx, queryUrl, &res, WithPage(opts.Page), WithPageLimit(opts.PerPage))
	if err != nil {
		return nil, "", nil, err
	}

	return res.Response, nextPage, annotation, nil
}

func (f *FreshBooksClient) RequestBusinessID(ctx context.Context) (int64, error) {
	queryUrl := getBussinessID
	var response ResponseBID

	var opts []ReqOpt
	_, _, err := f.doRequest(ctx, http.MethodGet, queryUrl, &response, nil, opts...)
	if err != nil {
		return 0, err
	}

	if response.Response.BusinessMemberships == nil || len(response.Response.BusinessMemberships) == 0 {
		return 0, fmt.Errorf("business ID not found")
	}

	return response.Response.BusinessMemberships[0].Business.ID, nil
}

func (f *FreshBooksClient) doRequest(
	ctx context.Context,
	method string,
	endpointUrl string,
	res interface{},
	body interface{},
	reqOpts ...ReqOpt,
) (http.Header, annotations.Annotations, error) {
	var (
		resp *http.Response
		err  error
	)

	urlAddress, err := url.Parse(endpointUrl)
	if err != nil {
		return nil, nil, err
	}

	for _, o := range reqOpts {
		o(urlAddress)
	}
	req, err := f.client.NewRequest(
		ctx,
		method,
		urlAddress,
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithHeader("Authorization", "Bearer "+f.GetToken()),
	)
	if err != nil {
		return nil, nil, err
	}

	resp, err = f.client.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if res != nil {
		bodyContent, err := io.ReadAll(resp.Body)

		if err != nil {
			return nil, nil, err
		}
		err = json.Unmarshal(bodyContent, &res)
		if err != nil {
			return nil, nil, err
		}
	}

	annotation := annotations.Annotations{}
	return resp.Header, annotation, nil
}
