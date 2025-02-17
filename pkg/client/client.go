package client

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"

	"github.com/conductorone/baton-sdk/pkg/annotations"
	"github.com/conductorone/baton-sdk/pkg/uhttp"
	"github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/tomnomnom/linkheader"
	"golang.org/x/oauth2"
)

const (
	getNewToken    = "https://api.freshbooks.com/auth/oauth/token" // #nosec G101
	getBussinessID = "https://api.freshbooks.com/auth/api/v1/users/me"

	baseURL        = "https://api.freshbooks.com/auth/api/v1/businesses/"
	getTeamMembers = "/team_members"
)

type FreshBooksClient struct {
	client      *uhttp.BaseHttpClient
	TokenSource oauth2.TokenSource
	Config      Config
}

type Config struct {
	businessID      string
	businessIDMutex sync.Mutex
}

type Option func(client *FreshBooksClient)

func WithBearerToken(apiToken string) Option {
	return func(client *FreshBooksClient) {
		client.SetToken(apiToken)
	}
}

// WithRefreshToken it receives a Refresh Token, Client ID and Client Secret from the platform to be able to renew the token when expired.
// The 3 arguments should be received when the connector is executed.
func WithRefreshToken(ctx context.Context, refreshToken, clientID, clientSecret string) Option {
	return func(client *FreshBooksClient) {
		token := &oauth2.Token{
			AccessToken:  "",
			RefreshToken: refreshToken,
			Expiry:       time.Now().Add(-1 * time.Second),
		}

		config := oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Endpoint: oauth2.Endpoint{
				TokenURL: getNewToken,
			},
		}
		tokenSource := oauth2.ReuseTokenSource(token, config.TokenSource(ctx, token))

		client.TokenSource = tokenSource
	}
}

func (f *FreshBooksClient) EnsureBusinessID(ctx context.Context) error {
	f.Config.businessIDMutex.Lock()
	defer f.Config.businessIDMutex.Unlock()

	if f.GetBusinessID() == "" {
		businessID, err := f.RequestBusinessID(ctx)
		if err != nil {
			return err
		}
		f.SetBusinessID(businessID)
	}

	return nil
}

func (f *FreshBooksClient) GetBusinessID() string {
	return f.Config.businessID
}

func (f *FreshBooksClient) SetBusinessID(bid int64) {
	f.Config.businessID = strconv.FormatInt(bid, 10)
}

func (f *FreshBooksClient) GetToken() (*oauth2.Token, error) {
	return f.TokenSource.Token()
}

func (f *FreshBooksClient) SetToken(token string) {
	f.TokenSource = oauth2.StaticTokenSource(&oauth2.Token{AccessToken: token})
}

func New(ctx context.Context, opts ...Option) (*FreshBooksClient, error) {
	httpClient, err := uhttp.NewClient(ctx, uhttp.WithLogger(true, ctxzap.Extract(ctx)))
	if err != nil {
		return nil, err
	}

	cli, err := uhttp.NewBaseHttpClientWithContext(context.Background(), httpClient)
	if err != nil {
		return nil, err
	}

	fbClient := FreshBooksClient{
		client: cli,
	}

	for _, o := range opts {
		o(&fbClient)
	}

	return &fbClient, nil
}

// getListFromAPI sends a request to the FreshBooks API to receive a JSON with a list of entities.
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
	var response ResponseBID
	var opts []ReqOpt
	queryUrl := getBussinessID

	_, _, err := f.doRequest(ctx, http.MethodGet, queryUrl, &response, nil, opts...)
	if err != nil {
		return 0, err
	}

	if len(response.Response.BusinessMemberships) == 0 {
		return 0, fmt.Errorf("business ID not found")
	}

	return response.Response.BusinessMemberships[0].Business.ID, nil
}

func (f *FreshBooksClient) doRequest(
	ctx context.Context,
	method string,
	endpointUrl string,
	res interface{},
	_ interface{},
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

	clientToken, err := f.GetToken()
	if err != nil {
		return nil, nil, err
	}

	req, err := f.client.NewRequest(
		ctx,
		method,
		urlAddress,
		uhttp.WithAcceptJSONHeader(),
		uhttp.WithContentTypeJSONHeader(),
		uhttp.WithHeader("Authorization", "Bearer "+clientToken.AccessToken),
	)
	if err != nil {
		return nil, nil, err
	}

	resp, err = f.client.Do(req)
	if err != nil {
		return nil, nil, err
	}
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
