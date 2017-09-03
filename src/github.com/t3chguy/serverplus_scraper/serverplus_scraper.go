package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/pkg/errors"
	"github.com/t3chguy/serverplus_scraper/serverplus"
	"github.com/t3chguy/serverplus_scraper/ucrm"
	"net/http"
	"os"
	"strconv"
	"time"
)

const serverplusTimeString = "2006-01-02 15:04:05"
const ucrmTimeString = "02.01.06 15.04"

type response struct {
	TicketID   string `xml:"ticket_id"`
	IncidentID int    `xml:"incident_id"`
}

type MailGunEmail struct {
	Recipient string `form:"recipient"`
	Sender    string `form:"sender"`
	From      string `form:"from"`
	Subject   string `form:"subject"`
	BodyPlain string `form:"body-plain"`

	//StrippedText      *string `form:"stripped-text"`
	//StrippedSignature *string `form:"stripped-signature"`
	//BodyHTML          *string `form:"body-html"`
	//StrippedHTML      *string `form:"stripped-html"`

	AttachmentCount int    `form:"attachment-count"`
	Timestamp       int    `form:"timestamp"`
	Token           string `form:"token"`
	Signature       string `form:"signature"`

	//MessageHeaders string `form:"message-headers"`
	//ContentIDMap string `form:"content-id-map"`
}

func (email MailGunEmail) Verify(apiKey string) bool {
	mac := hmac.New(sha256.New, []byte(apiKey))
	mac.Write([]byte(strconv.Itoa(email.Timestamp) + email.Token))
	expectedMAC := mac.Sum(nil)

	return hex.EncodeToString(expectedMAC) == email.Signature
}

type Config struct {
	UCRMURL            string
	UCRMAppKey         string
	UCRMCatchall       int
	Port               string
	MGAPIKey           string
	ServerPlusUsername string
	ServerPlusPassword string
}

var config Config

func main() {
	ucrmCatchAll, err := strconv.Atoi(os.Getenv("UCRMCATCHALL"))
	if err != nil {
		panic(err)
	}

	config = Config{
		os.Getenv("UCRMURL"),
		os.Getenv("UCRMAPPKEY"),
		ucrmCatchAll,
		os.Getenv("PORT"),
		os.Getenv("MGAPIKEY"),
		os.Getenv("SERVERPLUSUSERNAME"),
		os.Getenv("SERVERPLUSPASSWORD"),
	}

	ucrmClient := ucrm.NewClient(config.UCRMURL, config.UCRMAppKey)

	r := gin.Default()
	r.POST("/mailgun", func(c *gin.Context) {
		var req MailGunEmail
		err := c.Bind(&req)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if !req.Verify(config.MGAPIKey) {
			c.AbortWithError(http.StatusForbidden, errors.New("Unable to Verify payload."))
			return
		}

		clientID := config.UCRMCatchall
		if client, found := ucrmClient.GetClientByEmail(req.Sender); found {
			clientID = client.ID
		}

		subject := fmt.Sprintf("%s [Ticker from Email]", req.Subject)

		resp, err := ucrmClient.TicketsCreate(&ucrm.TicketsCreateRequest{
			Subject:  subject,
			ClientID: clientID,
			Comments: []ucrm.Comment{
				{
					TicketID: 0,
					Body: fmt.Sprintf("Sent by: %s\n", req.From) +
						fmt.Sprintf("-------------------------------------\n\n%s", req.BodyPlain),
				},
			},
		})

		if err == nil && resp != nil {
			c.JSON(200, response{"", resp.TicketID})
			return
		}

		c.Status(http.StatusInternalServerError)
		c.Error(err)
	})

	r.POST("/serverplus", func(c *gin.Context) {
		var req serverplus.ExportApiData
		err := c.MustBindWith(&req, binding.FormPost)
		if err != nil {
			c.AbortWithError(http.StatusInternalServerError, err)
			return
		}
		if !req.Verify(config.ServerPlusUsername, config.ServerPlusPassword) {
			c.AbortWithError(http.StatusForbidden, errors.New("Unable to Verify payload."))
			return
		}

		createdTime, err := time.Parse(serverplusTimeString, req.Time)
		usCentralTime := createdTime.Add(-6 * time.Hour).Format(ucrmTimeString)
		subject := fmt.Sprintf("%s, %s, %s [Ticket from ServerPlus]", req.TicketID, usCentralTime, req.Service)

		resp, err := ucrmClient.TicketsCreate(&ucrm.TicketsCreateRequest{
			Subject:  subject,
			ClientID: config.UCRMCatchall,
			Comments: []ucrm.Comment{
				{
					TicketID: 0,
					Body:     req.TextRepresentation,
				},
			},
		})

		if err == nil && resp != nil {
			c.XML(200, response{"", resp.TicketID})
			return
		}

		c.Status(http.StatusInternalServerError)
		c.Error(err)
	})
	r.Run(":" + config.Port)
}
