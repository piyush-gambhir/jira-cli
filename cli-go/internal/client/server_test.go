package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/piyush-gambhir/jira-cli/cli-go/internal/auth"
	"github.com/piyush-gambhir/jira-cli/cli-go/internal/config"
)

// serverClient is a Server/DC (PAT, REST API v2) client for site.
func serverClient(t *testing.T, site string) *Client {
	t.Helper()
	a, err := auth.New(config.Profile{Site: site, Token: "pat", AuthType: config.AuthPAT}, "", nil)
	if err != nil {
		t.Fatal(err)
	}
	return NewClient(a, "2", false, false)
}

type recorded struct {
	method, path, query, body string
}

// recorder serves canned JSON by "METHOD path" and records every request.
func recorder(t *testing.T, responses map[string]string) (*httptest.Server, *[]recorded) {
	t.Helper()
	var reqs []recorded
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		reqs = append(reqs, recorded{r.Method, r.URL.Path, r.URL.RawQuery, string(b)})
		resp, ok := responses[r.Method+" "+r.URL.Path]
		if !ok {
			http.Error(w, `{"errorMessages":["not found"]}`, http.StatusNotFound)
			return
		}
		if resp == "" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		_, _ = io.WriteString(w, resp)
	}))
	t.Cleanup(ts.Close)
	return ts, &reqs
}

func TestServerUsersUseUsernames(t *testing.T) {
	ts, reqs := recorder(t, map[string]string{
		"GET /rest/api/2/user/search":             `[{"name":"jane","key":"JIRAUSER10100","displayName":"Jane Doe"}]`,
		"GET /rest/api/2/myself":                  `{"name":"me","key":"JIRAUSER1"}`,
		"PUT /rest/api/2/issue/ABC-1/assignee":    "",
		"DELETE /rest/api/2/issue/ABC-1/watchers": "",
		"POST /rest/api/2/group/user":             "",
		"GET /rest/api/2/user":                    `{"name":"jane","groups":{"size":1,"items":[{"name":"devs"}]}}`,
	})
	c := serverClient(t, ts.URL)

	id, err := c.ResolveUser("jane@example.com")
	if err != nil || id != "jane" {
		t.Fatalf("ResolveUser = %q, %v; want the username jane", id, err)
	}
	if got := (*reqs)[0].query; got != "maxResults=2&username=jane%40example.com" {
		t.Fatalf("user search query = %q; want username=", got)
	}
	if me, err := c.ResolveUser("@me"); err != nil || me != "me" {
		t.Fatalf("ResolveUser(@me) = %q, %v; want me", me, err)
	}
	if err := c.AssignIssue("ABC-1", id); err != nil {
		t.Fatal(err)
	}
	if err := c.RemoveWatcher("ABC-1", id); err != nil {
		t.Fatal(err)
	}
	if err := c.AddGroupUser("devs", id); err != nil {
		t.Fatal(err)
	}
	groups, err := c.UserGroups(id)
	if err != nil || len(groups) != 1 || groups[0].Name != "devs" {
		t.Fatalf("UserGroups = %#v, %v; want [devs]", groups, err)
	}
	if _, err := c.BulkUsers([]string{"jane"}); err == nil || !strings.Contains(err.Error(), "only available on Jira Cloud") {
		t.Fatalf("BulkUsers on Server/DC: err = %v; want a Cloud-only error", err)
	}

	want := []recorded{
		{"PUT", "/rest/api/2/issue/ABC-1/assignee", "", `{"name":"jane"}`},
		{"DELETE", "/rest/api/2/issue/ABC-1/watchers", "username=jane", ""},
		{"POST", "/rest/api/2/group/user", "groupname=devs", `{"name":"jane"}`},
		{"GET", "/rest/api/2/user", "expand=groups&username=jane", ""},
	}
	got := (*reqs)[2:]
	if len(got) != len(want) {
		t.Fatalf("requests = %#v", got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("request %d = %#v; want %#v", i, got[i], want[i])
		}
	}
}

func TestCloudUsersUseAccountIDs(t *testing.T) {
	ts, reqs := recorder(t, map[string]string{
		"GET /rest/api/3/user/search":          `[{"accountId":"5b10ac","displayName":"Jane Doe"}]`,
		"PUT /rest/api/3/issue/ABC-1/assignee": "",
	})
	c := testClient(t, ts.URL)
	id, err := c.ResolveUser("jane")
	if err != nil || id != "5b10ac" {
		t.Fatalf("ResolveUser = %q, %v; want the accountId", id, err)
	}
	if err := c.AssignIssue("ABC-1", id); err != nil {
		t.Fatal(err)
	}
	if q := (*reqs)[0].query; q != "maxResults=2&query=jane" {
		t.Fatalf("user search query = %q; want query=", q)
	}
	if b := (*reqs)[1].body; b != `{"accountId":"5b10ac"}` {
		t.Fatalf("assign body = %s; want accountId", b)
	}
}

func TestServerProjectListFiltersLocally(t *testing.T) {
	ts, reqs := recorder(t, map[string]string{
		"GET /rest/api/2/project": `[{"key":"WEB","name":"Website"},{"key":"MOB","name":"Mobile"},{"key":"MOBX","name":"Mobile Next"}]`,
	})
	c := serverClient(t, ts.URL)
	got, err := c.ListProjects("mob", 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Key != "MOB" {
		t.Fatalf("ListProjects(mob, 1) = %#v; want [MOB]", got)
	}
	if all, err := c.ListProjects("", 0); err != nil || len(all) != 3 {
		t.Fatalf("ListProjects(\"\", 0) = %d projects, %v; want 3", len(all), err)
	}
	if (*reqs)[0].query != "" {
		t.Fatalf("Server/DC project list sent query %q", (*reqs)[0].query)
	}
}

func TestServerCreateMetaReadsValues(t *testing.T) {
	ts, _ := recorder(t, map[string]string{
		"GET /rest/api/2/issue/createmeta/ABC/issuetypes":   `{"start":0,"size":1,"last":true,"values":[{"id":"1","name":"Bug"}]}`,
		"GET /rest/api/2/issue/createmeta/ABC/issuetypes/1": `{"start":0,"size":1,"last":true,"values":[{"fieldId":"summary","name":"Summary","required":true}]}`,
	})
	c := serverClient(t, ts.URL)
	types, err := c.CreateMetaIssueTypes("ABC")
	if err != nil || len(types) != 1 || types[0].Name != "Bug" {
		t.Fatalf("CreateMetaIssueTypes = %#v, %v; want [Bug]", types, err)
	}
	fields, err := c.CreateMetaFields("ABC", "1")
	if err != nil || !fields["summary"].Required {
		t.Fatalf("CreateMetaFields = %#v, %v; want a required summary", fields, err)
	}
}

func TestServerAttachmentDownloadUsesContentURL(t *testing.T) {
	var gotAuth, gotAccept string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.EscapedPath() {
		case "/jira/rest/api/2/attachment/10000":
			// Jira's own base URL may name another host; the path is what counts.
			_, _ = io.WriteString(w, `{"id":10000,"filename":"a b.txt","content":"http://jira.internal:8080/jira/secure/attachment/10000/a%20b.txt"}`)
		case "/jira/secure/attachment/10000/a%20b.txt":
			gotAuth, gotAccept = r.Header.Get("Authorization"), r.Header.Get("Accept")
			_, _ = io.WriteString(w, "file bytes")
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	c := serverClient(t, ts.URL+"/jira")
	data, name, err := c.DownloadAttachment("10000")
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "file bytes" || name != "a b.txt" {
		t.Fatalf("DownloadAttachment = %q, %q", data, name)
	}
	if gotAuth != "Bearer pat" || gotAccept != "*/*" {
		t.Fatalf("download headers: Authorization=%q Accept=%q", gotAuth, gotAccept)
	}
}

func TestServerAttachmentDownloadRejectsForeignPath(t *testing.T) {
	ts, _ := recorder(t, map[string]string{
		"GET /jira/rest/api/2/attachment/1": `{"id":"1","filename":"x","content":"http://evil.example/other/x"}`,
	})
	c := serverClient(t, ts.URL+"/jira")
	if _, _, err := c.DownloadAttachment("1"); err == nil || !strings.Contains(err.Error(), "outside the configured site") {
		t.Fatalf("err = %v; want an outside-the-site error", err)
	}
}

func TestBulkUsersFollowsPages(t *testing.T) {
	var starts []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query()["accountId"]
		start := 0
		if s := r.URL.Query().Get("startAt"); s != "" {
			start = atoiOrZero(s)
		}
		starts = append(starts, r.URL.Query().Get("startAt"))
		end := min(start+10, len(ids)) // Jira pages at 10 users by default
		var page struct {
			Values []User `json:"values"`
			IsLast bool   `json:"isLast"`
		}
		for _, id := range ids[start:end] {
			page.Values = append(page.Values, User{AccountID: id})
		}
		page.IsLast = end == len(ids)
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer ts.Close()

	ids := make([]string, 11)
	for i := range ids {
		ids[i] = "acct-" + string(rune('a'+i))
	}
	users, err := testClient(t, ts.URL).BulkUsers(ids)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 11 {
		t.Fatalf("BulkUsers returned %d users; want 11 (startAt values %v)", len(users), starts)
	}
}

func atoiOrZero(s string) int {
	n := 0
	for _, r := range s {
		n = n*10 + int(r-'0')
	}
	return n
}

func TestSearchKeepsRequestedCustomFields(t *testing.T) {
	ts, _ := recorder(t, map[string]string{
		"POST /rest/api/3/search/jql": `{"isLast":true,"issues":[{"key":"ABC-1","fields":{"summary":"S","status":{"name":"Done"},"customfield_10010":{"value":"High"},"components":[{"name":"api"}],"customfield_10020":null}}]}`,
	})
	issues, err := testClient(t, ts.URL).SearchIssues("project = ABC", []string{"summary", "status", "customfield_10010", "components", "customfield_10020"}, 10, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(issues) != 1 || issues[0].Fields.Summary != "S" || statusOf(issues[0]) != "Done" {
		t.Fatalf("typed fields not decoded: %#v", issues)
	}
	b, err := json.Marshal(issues[0])
	if err != nil {
		t.Fatal(err)
	}
	var out struct {
		Fields map[string]json.RawMessage `json:"fields"`
	}
	if err := json.Unmarshal(b, &out); err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{
		"summary":           `"S"`,
		"customfield_10010": `{"value":"High"}`,
		"components":        `[{"name":"api"}]`,
		"customfield_10020": `null`,
	} {
		if got := string(out.Fields[k]); got != want {
			t.Errorf("fields.%s = %s; want %s (output %s)", k, got, want, b)
		}
	}
}

func statusOf(i Issue) string {
	if i.Fields.Status == nil {
		return ""
	}
	return i.Fields.Status.Name
}
