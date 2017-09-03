package serverplus

type DecisionTreeEntry struct {
	data [2]string
}

func (dte DecisionTreeEntry) Question() string {
	return dte.data[0]
}

func (dte DecisionTreeEntry) Answer() string {
	return dte.data[1]
}

type DecisionTree []DecisionTreeEntry

type EscalationInfo [][]string

func (ei EscalationInfo) Get(target string) (string, bool) {
	for _, entry := range ei {
		key, value := entry[0], entry[1]
		if target == key {
			return value, true
		}
	}
	return "", false
}

type ExportApiData struct {
	SubscriberID    string `form:"subscriber_id"`
	SubscriberName  string `form:"subscriber_name"`
	SubscriberEmail string `form:"subscriber_email"`
	SubscriberPhone string `form:"subscriber_phone"`

	Time    string `form:"time"`
	TimeUTC string `form:"time_utc"`

	// Enum: {resolved, pending, open, escalated}
	Status string `form:"status"`

	Notes   string `form:"notes"`
	Service string `form:"service"`
	Issue   string `form:"issue"`
	IssueID string `form:"issue_id"`

	DecisionTree   DecisionTree   `form:"dt_form"`
	EscalationInfo EscalationInfo `form:"escalation_info_form"`

	Staff string `form:"staff"`

	TextRepresentation string `form:"text_representation"`

	IncidentID       string `form:"incident_id"`
	TicketID         string `form:"ticket_id"`
	TicketExternalID string `form:"ticket_external_id"`
	CompanyID        string `form:"company_id"`

	Username string `form:"username"`
	Password string `form:"password"`
}

func (data ExportApiData) Verify(username, password string) bool {
	return data.Username == username && data.Password == password
}
