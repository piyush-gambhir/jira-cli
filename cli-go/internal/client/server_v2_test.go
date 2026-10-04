package client

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestServerCloudOnlyCallsFailBeforeAnyRequest(t *testing.T) {
	ts, reqs := recorder(t, nil)
	c := serverClient(t, ts.URL)
	calls := map[string]func() error{
		"filter list":   func() error { _, err := c.ListMyFilters(); return err },
		"filter search": func() error { _, err := c.SearchFilters("x", 10); return err },
		"group members --group-id": func() error {
			_, err := c.GroupMembers("devs", "abc-123", 10, false)
			return err
		},
		"label list":                    func() error { _, err := c.ListLabels(10); return err },
		"jql parse":                     func() error { _, err := c.JQLParse([]string{"project = A"}, ""); return err },
		"permission permitted-projects": func() error { _, err := c.PermittedProjects([]string{"BROWSE_PROJECTS"}); return err },
		"webhook list":                  func() error { _, err := c.ListWebhooks(10); return err },
		"webhook register": func() error {
			_, err := c.RegisterWebhooks("https://hook.example", []WebhookRegistration{{JqlFilter: "project = A", Events: []string{"jira:issue_created"}}})
			return err
		},
		"webhook delete":  func() error { return c.DeleteWebhooks([]int{1}) },
		"webhook refresh": func() error { _, err := c.RefreshWebhooks([]int{1}); return err },
	}
	for name, call := range calls {
		if err := call(); err == nil || !strings.Contains(err.Error(), "only available on Jira Cloud") {
			t.Errorf("%s on Server/DC: err = %v; want a Cloud-only error", name, err)
		}
	}
	if len(*reqs) != 0 {
		t.Fatalf("Cloud-only calls sent requests: %#v", *reqs)
	}
}

func TestServerFilterFavouriteUsesV1Resource(t *testing.T) {
	type call struct{ method, path, token, body string }
	var calls []call
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b, _ := io.ReadAll(r.Body)
		calls = append(calls, call{r.Method, r.URL.Path, r.Header.Get("X-Atlassian-Token"), string(b)})
		switch r.Method + " " + r.URL.Path {
		case "PUT /rest/api/1.0/filters/10/favourite", "DELETE /rest/api/1.0/filters/10/favourite":
			w.WriteHeader(http.StatusNoContent)
		case "GET /rest/api/2/filter/10":
			_, _ = io.WriteString(w, `{"id":"10","name":"F","favourite":true}`)
		case "PUT /rest/api/2/filter/10":
			_, _ = io.WriteString(w, `{"id":"10","name":"New"}`)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()
	c := serverClient(t, ts.URL)

	if f, err := c.FavouriteFilter("10"); err != nil || !f.Favourite {
		t.Fatalf("FavouriteFilter = %#v, %v", f, err)
	}
	if _, err := c.UnfavouriteFilter("10"); err != nil {
		t.Fatal(err)
	}
	// REST v2 ignores "favourite" on update, so it goes to the 1.0 resource.
	if _, err := c.UpdateFilter("10", map[string]any{"name": "New", "favourite": false}); err != nil {
		t.Fatal(err)
	}
	want := []call{
		{"PUT", "/rest/api/1.0/filters/10/favourite", "no-check", ""},
		{"GET", "/rest/api/2/filter/10", "", ""},
		{"DELETE", "/rest/api/1.0/filters/10/favourite", "no-check", ""},
		{"GET", "/rest/api/2/filter/10", "", ""},
		{"PUT", "/rest/api/2/filter/10", "", `{"name":"New"}`},
		{"DELETE", "/rest/api/1.0/filters/10/favourite", "no-check", ""},
		{"GET", "/rest/api/2/filter/10", "", ""},
	}
	if len(calls) != len(want) {
		t.Fatalf("requests = %#v", calls)
	}
	for i := range want {
		if calls[i] != want[i] {
			t.Errorf("request %d = %#v; want %#v", i, calls[i], want[i])
		}
	}
}

func TestServerGroupListUsesGroupPicker(t *testing.T) {
	ts, reqs := recorder(t, map[string]string{
		"GET /rest/api/2/groups/picker": `{"header":"Showing 2 of 30 matching groups","total":30,"groups":[{"name":"devs"},{"name":"ops"}]}`,
	})
	groups, total, err := serverClient(t, ts.URL).ListGroups(50)
	if err != nil || len(groups) != 2 || groups[0].Name != "devs" || total != 30 {
		t.Fatalf("ListGroups = %#v, %d, %v; want devs, ops of 30", groups, total, err)
	}
	if q := (*reqs)[0].query; q != "maxResults=50" {
		t.Fatalf("group picker query = %q", q)
	}
}

func TestServerDashboardSearchPagesAndFiltersLocally(t *testing.T) {
	all := []Dashboard{{ID: "1", Name: "Team Alpha"}, {ID: "2", Name: "Ops"}, {ID: "3", Name: "Release team"}}
	var queries []string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/rest/api/2/dashboard" {
			http.NotFound(w, r)
			return
		}
		queries = append(queries, r.URL.RawQuery)
		start := atoiOrZero(r.URL.Query().Get("startAt"))
		end := min(start+2, len(all)) // the server caps maxResults at 2
		_ = json.NewEncoder(w).Encode(map[string]any{
			"startAt": start, "maxResults": 2, "total": len(all), "dashboards": all[start:end],
		})
	}))
	defer ts.Close()

	got, err := serverClient(t, ts.URL).SearchDashboards("TEAM", 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0].ID != "1" || got[1].ID != "3" {
		t.Fatalf("SearchDashboards = %#v; want dashboards 1 and 3", got)
	}
	if want := []string{"maxResults=100&startAt=0", "maxResults=2&startAt=2"}; strings.Join(queries, " ") != strings.Join(want, " ") {
		t.Fatalf("dashboard queries = %q; want %q", queries, want)
	}
}

func TestServerIssueTypesForProjectReadTheProject(t *testing.T) {
	ts, _ := recorder(t, map[string]string{
		"GET /rest/api/2/project/10001": `{"id":"10001","key":"ABC","issueTypes":[{"id":"1","name":"Bug"},{"id":"3","name":"Task"}]}`,
	})
	types, err := serverClient(t, ts.URL).ListIssueTypesForProject("10001")
	if err != nil || len(types) != 2 || types[1].Name != "Task" {
		t.Fatalf("ListIssueTypesForProject = %#v, %v; want Bug, Task", types, err)
	}
}

func TestServerChangelogExpandsTheIssue(t *testing.T) {
	ts, reqs := recorder(t, map[string]string{
		"GET /rest/api/2/issue/ABC-1": `{"key":"ABC-1","changelog":{"startAt":0,"maxResults":3,"total":3,"histories":[{"id":"1"},{"id":"2","items":[{"field":"status","fromString":"To Do","toString":"Done"}]},{"id":"3"}]}}`,
	})
	c := serverClient(t, ts.URL)
	got, err := c.GetChangelog("ABC-1", 1, 1)
	if err != nil || len(got) != 1 || got[0].ID != "2" || got[0].Items[0].ToString != "Done" {
		t.Fatalf("GetChangelog(1, 1) = %#v, %v; want history 2", got, err)
	}
	if all, err := c.GetChangelog("ABC-1", 0, 0); err != nil || len(all) != 3 {
		t.Fatalf("GetChangelog(0, 0) = %d entries, %v; want 3", len(all), err)
	}
	if q := (*reqs)[0].query; q != "expand=changelog&fields=summary" {
		t.Fatalf("changelog query = %q", q)
	}
}

func TestServerUserListUsesUserList(t *testing.T) {
	ts, reqs := recorder(t, map[string]string{
		"GET /rest/api/2/user/list": `{"isLast":true,"values":[{"name":"jane","displayName":"Jane"}]}`,
	})
	users, err := serverClient(t, ts.URL).ListAllUsers(25)
	if err != nil || len(users) != 1 || users[0].Name != "jane" {
		t.Fatalf("ListAllUsers = %#v, %v; want jane", users, err)
	}
	if q := (*reqs)[0].query; q != "maxResults=25" {
		t.Fatalf("user list query = %q", q)
	}

	old, _ := recorder(t, nil) // a Data Center release without user/list
	if _, err := serverClient(t, old.URL).ListAllUsers(25); err == nil || !strings.Contains(err.Error(), "jira user search") {
		t.Fatalf("ListAllUsers on an older release: err = %v; want a pointer to user search", err)
	}
}

func TestServerProjectArchiveAndRestoreUsePut(t *testing.T) {
	ts, reqs := recorder(t, map[string]string{
		"PUT /rest/api/2/project/ABC/archive": "",
		"PUT /rest/api/2/project/ABC/restore": "",
	})
	c := serverClient(t, ts.URL)
	if err := c.ArchiveProject("ABC"); err != nil {
		t.Fatal(err)
	}
	if err := c.RestoreProject("ABC"); err != nil {
		t.Fatal(err)
	}
	if len(*reqs) != 2 {
		t.Fatalf("requests = %#v", *reqs)
	}
}
