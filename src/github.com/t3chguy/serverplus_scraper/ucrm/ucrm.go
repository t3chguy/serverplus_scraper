package ucrm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
)

type Client struct {
	url    string
	appKey string
	Client *http.Client
}

func (cli *Client) BuildURL(urlPath ...string) string {
	baseUrl, _ := url.Parse(cli.url)
	parts := []string{baseUrl.Path, "api", "v1.0"}
	parts = append(parts, urlPath...)
	baseUrl.Path = path.Join(parts...)
	return baseUrl.String()
}

type RespError struct {
	ErrCode string `json:"code"`
	Err     string `json:"message"`
}

func (e RespError) Error() string {
	return e.ErrCode + ": " + e.Err
}

type HTTPError struct {
	WrappedError error
	Message      string
	Code         int
}

func (e HTTPError) Error() string {
	var wrappedErrMsg string
	if e.WrappedError != nil {
		wrappedErrMsg = e.WrappedError.Error()
	}
	return fmt.Sprintf("msg=%s code=%d wrapped=%s", e.Message, e.Code, wrappedErrMsg)
}

func (cli *Client) MakeRequest(method string, httpURL string, reqBody interface{}, resBody interface{}) ([]byte, error) {
	var req *http.Request

	var err error
	if reqBody != nil {
		var jsonStr []byte
		jsonStr, err = json.Marshal(reqBody)
		fmt.Println(string(jsonStr))
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, httpURL, bytes.NewBuffer(jsonStr))
	} else {
		req, err = http.NewRequest(method, httpURL, nil)
	}

	if err != nil {
		return nil, err
	}

	req.Header.Set("X-Auth-App-Key", cli.appKey)
	req.Header.Set("Content-Type", "application/json")

	res, err := cli.Client.Do(req)
	if res != nil {
		defer res.Body.Close()
	}
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(res.Body)
	if err == nil && len(contents) == 0 {
		return []byte(res.Header.Get("Location")), nil
	}
	if res.StatusCode/100 != 2 { // not 2xx
		var wrap error
		var respErr RespError
		if _ = json.Unmarshal(contents, &respErr); respErr.ErrCode != "" {
			wrap = respErr
		}

		// If we failed to decode as RespError, don't just drop the HTTP body, include it in the
		// HTTP error instead (e.g proxy errors which return HTML).
		msg := "Failed to " + method + " JSON to " + req.URL.Path
		if wrap == nil {
			msg = msg + ": " + string(contents)
		}

		return contents, HTTPError{
			Code:         res.StatusCode,
			Message:      msg,
			WrappedError: wrap,
		}
	}
	if err != nil {
		return nil, err
	}

	if resBody != nil {
		if err = json.Unmarshal(contents, &resBody); err != nil {
			return nil, err
		}
	}

	return contents, nil
}

type Comment struct {
	UserID    *string `json:"userId,omitempty"`
	TicketID  int     `json:"ticketId"`
	Body      string  `json:"body"`
	Public    *bool   `json:"public,omitempty"`
	CreatedAt *string `json:"createdAt,omitempty"`
}

type TicketsCreateRequest struct {
	Subject        string  `json:"subject"`
	ClientID       int     `json:"clientId"`
	AssignedUserID *int    `json:"assignedUserId,omitempty"`
	CreateAt       *string `json:"createdAt,omitempty"`

	// Enum 0=New 1=Open 2=Pending 3=Solved
	Status *int `json:"status,omitempty"`

	Comments []Comment `json:"comments"`
}
type TicketsCreateResponse struct {
	TicketID int
}

func (cli *Client) TicketsCreate(request *TicketsCreateRequest) (resp *TicketsCreateResponse, err error) {
	urlPath := cli.BuildURL("ticketing", "tickets")
	var location []byte
	location, err = cli.MakeRequest("POST", urlPath, request, nil)
	if err == nil {
		parts := strings.SplitN(string(location), "/ticketing/tickets/", 2)
		if len(parts) < 2 {
			return
		}
		if num, err := strconv.Atoi(parts[1]); err == nil {
			resp = &TicketsCreateResponse{
				TicketID: num,
			}
		}
	}
	return
}

type ClientResponseContact struct {
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	IsBilling bool   `json:"isBilling"`
	IsContact bool   `json:"isContact"`
	ID        int    `json:"id"`
	ClientID  int    `json:"clientId"`
}

type ClientResponseAttributes struct {
	Value             string `json:"value"`
	CustomAttributeID int    `json:"customAttributeId"`
	ID                int    `json:"id"`
	ClientID          int    `json:"clientId"`
	Name              string `json:"name"`
	Key               string `json:"name"`
}

type ClientsResponseClient struct {
	UserIdent      string `json:"userIdent"`
	OrganizationID int    `json:"organizationId"`

	// Enum 1=Residential 2=Company
	ClientType int `json:"clientType"`

	CompanyName               *string `json:"companyName"`
	CompanyRegistrationNumber *string `json:"companyRegistrationNumber"`
	CompanyTaxID              *string `json:"companyTaxId"`
	CompanyWebsite            *string `json:"companyWebsite"`
	CompanyContactFirstName   *string `json:"companyContactFirstName"`
	CompanyContactLastName    *string `json:"companyContactLastName"`

	FirstName *string `json:"firstName"`
	LastName  *string `json:"lastName"`

	Street1   string `json:"street1"`
	Street2   string `json:"street2"`
	City      string `json:"city"`
	CountryID int    `json:"countryId"`
	StateID   int    `json:"stateId"`
	ZipCode   string `json:"zipCode"`

	//InvoiceAddressSameAsContact bool `json:"invoiceAddressSameAsContact"`
	//InvoiceStreet1 string `json:"invoiceStreet1"`
	//InvoiceStreet2 string `json:"invoiceStreet2"`
	//InvoiceCity string `json:"invoiceCity"`
	//InvoiceCountryID int `json:"invoiceCountryId"`
	//InvoiceStateID int `json:"invoiceStateId"`
	//InvoiceZipCode string `json:"invoiceZipCode"`
	//SendInvoiceByPost bool `json:"sendInvoiceByPost"`
	//InvoiceMaturityDays int `json:"invoiceMaturityDays"`
	//StopServiceDue bool `json:"stopServiceDue"`
	//StopServiceDueDays int `json:"stopServiceDueDays"`
	//Tax1ID *int `json:"tax1Id"`
	//Tax2ID *int `json:"tax2Id"`
	//Tax3ID *int `json:"tax3Id"`
	//RegistrationDate string `json:"registrationDate"`
	//PreviousISP string `json:"previousIsp"`

	Note       string                     `json:"note"`
	Username   *string                    `json:"username"`
	ID         int                        `json:"id"`
	Contacts   []ClientResponseContact    `json:"contacts"`
	Attributes []ClientResponseAttributes `json:"attributes"`
}

type ClientsResponse []ClientsResponseClient

func (cli *Client) GetClients() (resp ClientsResponse, err error) {
	urlPath := cli.BuildURL("clients")
	_, err = cli.MakeRequest("GET", urlPath, nil, &resp)
	return
}

func (cli *Client) GetClientByEmail(email string) (resp *ClientsResponseClient, found bool) {
	clients, err := cli.GetClients()
	if err != nil {
		return
	}

	for _, client := range clients {
		for _, clientContact := range client.Contacts {
			if clientContact.Email == email {
				return &client, true
			}
		}
	}
	return
}

func NewClient(url, appKey string) *Client {
	return &Client{
		url,
		appKey,
		http.DefaultClient,
	}
}
