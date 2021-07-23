package internal

import (
	"bytes"
	"fmt"

	"net/http"
	"os"

	"strings"

	"github.com/GrooveCommunity/go-dispatcher/entity"
	"github.com/andygrunwald/go-jira"

	"time"

	gcp "github.com/GrooveCommunity/glib-cloud-storage/gcp"
)

type Customfield10366 struct {
	Value string `json:"value"`
}

type Fields struct {
	Customfield10366 Customfield10366 `json:"customfield_10366"`
}

type DataField struct {
	Fields Fields `json:"fields"`
}

type Issue struct {
	ID                 string `json:"id,omitempty"`
	Description        string `json:"description,omitempty"`
	Reporter           string `json:"reporter,omitempty"`
	CreatedDate        string `json:"created_date,omitempty"`
	Type               string `json:"type,omitempty"`
	Priority           string `json:"priority,omitempty"`
	ProductServiceDesk string `json:"priority,omitempty"`
}

type Response struct {
	Issues []Issue `json:"issues,omitempty"`
}

var rules []entity.Rule

func ForwardIssue(username, token, endpoint string) {

	rules = GetRules()

	tp := jira.BasicAuthTransport{
		Username: username, //usuário do jira
		Password: token,    //token de api
	}

	client, err := jira.NewClient(tp.Client(), strings.TrimSpace(endpoint))
	if err != nil {
		panic("\nError:" + err.Error())
	}

	for {

		fmt.Println("Regras consideradas: ", rules)

		for _, rule := range rules {

			for _, field := range rule.Forward.Input.Fields {

				content := ""
				attachment := ""

				if rule.Forward.Input.Content != "" {
					content = " and text ~ '" + rule.Forward.Input.Content + "'"
				}

				if rule.Forward.Input.HasAttachment {
					attachment = "and NOT attachments is EMPTY"
				}

				jql := "project = 'service desk' and type = incidente and status = 'AGUARDANDO SD' and '" + field.Name + "' = '" + field.Value + "'" + content + " " + attachment

				issuesJira, err := getAllIssues(client, jql)

				if err != nil && !(strings.HasPrefix(err.Error(), "No response returned")) {
					panic(err)
				}

				for _, v := range issuesJira {

					updateStatusIssue(&v, client, "Analisar - SD")
					updateStatusIssue(&v, client, "Acionar Squad")

					updateIssueCustomField(entity.Issue{KeyID: v.Key, Rule: rule})

					if err != nil {
						panic(err)
					}
				}

			}
		}

		fmt.Println("Aguardando um minuto")

		time.Sleep(1 * time.Minute)
	}
}

func UpdateRules(rule entity.Rule) {
	rules = append(rules, rule)
}

func getAllIssues(client *jira.Client, searchString string) ([]jira.Issue, error) {
	last := 0
	var issues []jira.Issue
	for {
		opt := &jira.SearchOptions{
			MaxResults: 1000, // Max results can go up to 1000
			StartAt:    last,
		}

		chunk, resp, err := client.Issue.Search(searchString, opt)
		if err != nil {
			return nil, err
		}

		total := resp.Total
		if issues == nil {
			issues = make([]jira.Issue, 0, total)
		}
		issues = append(issues, chunk...)
		last = resp.StartAt + len(chunk)
		if last >= total {
			return issues, nil
		}
	}

}

func updateStatusIssue(issue *jira.Issue, client *jira.Client, status string) {
	fmt.Println("Issue ID:" + issue.ID)

	var transitionID string
	possibleTransitions, _, err := client.Issue.GetTransitions(issue.ID)

	if err != nil {
		panic("\nError:" + err.Error())
	}

	for _, transitions := range possibleTransitions {
		if transitions.Name == status {
			transitionID = transitions.ID
			break
		}
	}

	_, errorTransition := client.Issue.DoTransition(issue.ID, transitionID)

	if errorTransition != nil {
		panic("\nError:" + errorTransition.Error())
	}

	fmt.Println("Status atualizado para " + status)

}

func updateIssueCustomField(issue entity.Issue) {
	host := os.Getenv("JIRA_ENDPOINT") + "/rest/api/2/issue/" + issue.KeyID

	data := "{\"fields\": {\"" + issue.Rule.Forward.Output.CustomFieldID + "\":{\"value\":\"" + issue.Rule.Forward.Output.CustomFieldValue + "\"}}}"

	req, err := http.NewRequest(http.MethodPut, host, bytes.NewBuffer([]byte(data)))
	req.SetBasicAuth(os.Getenv("JIRA_USERNAME"), os.Getenv("JIRA_TOKENAPI"))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}

	resp, err := client.Do(req)

	defer resp.Body.Close()

	if err != nil {
		panic(err)
	}

	gcp.WriteObject(issue, "forwarded-calls", issue.KeyID)

	fmt.Println("Issue " + issue.KeyID + " atualizada!")
}
