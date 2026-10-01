package httpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"sync"
	"testing"
	"time"

	"myself/server/internal/config"
	"myself/server/internal/store"
)

func TestScheduledPublicationLifecycle(t *testing.T) {
	e := newEnv(t)
	due := time.Now().UTC().Truncate(time.Second).Add(time.Hour)
	post := map[string]any{"slug": "queued-proof", "title": "未来私密文章", "contentMd": "未来正文", "tags": []string{"future-only-tag"}, "status": "scheduled", "publishAt": due.Format(time.RFC3339)}
	note := map[string]any{"contentMd": "未来私密随想", "status": "scheduled", "publishAt": due.Format(time.RFC3339)}
	var p, n struct {
		ID                int64 `json:"id"`
		Status, PublishAt string
	}
	e.call("POST", "/api/v1/admin/posts", post, &p, 201)
	e.call("POST", "/api/v1/admin/notes", note, &n, 201)
	if p.Status != "scheduled" || n.Status != "scheduled" {
		t.Fatal(p, n)
	}
	token := e.token
	e.token = ""
	e.call("GET", "/api/v1/posts/queued-proof", nil, nil, 404)
	e.call("GET", fmt.Sprintf("/api/v1/notes/%d", n.ID), nil, nil, 404)
	e.call("POST", fmt.Sprintf("/api/v1/notes/%d/view", n.ID), nil, nil, 404)
	for _, url := range []string{"/api/v1/posts?q=未来私密", "/api/v1/notes?q=未来私密&all=1", "/api/v1/hero", "/api/v1/tags", "/feed"} {
		r := e.do("GET", url, nil, nil)
		data, _ := io.ReadAll(r.Body)
		r.Body.Close()
		if strings.Contains(string(data), "未来私密") || strings.Contains(string(data), "future-only-tag") {
			t.Fatalf("queued content leaked into %s: %s", url, data)
		}
	}
	e.call("POST", "/api/v1/reactions", map[string]any{"target": "note", "id": n.ID, "kind": "like"}, nil, 404)
	e.token = token
	// Editing content without publication fields must retain the queue.
	e.call("PUT", fmt.Sprintf("/api/v1/admin/notes/%d", n.ID), map[string]any{"contentMd": "已修改的未来随想"}, nil, 200)
	var queued store.Note
	e.call("GET", fmt.Sprintf("/api/v1/admin/notes/%d", n.ID), nil, &queued, 200)
	if queued.Status != "scheduled" || queued.PublishAt != due.Format(time.RFC3339) {
		t.Fatalf("lost schedule: %+v", queued)
	}
	if err := e.server.publishDue(context.Background(), due.Add(-time.Second)); err != nil {
		t.Fatal(err)
	}
	row, _ := e.server.DB.GetPost(`WHERE id = ?`, p.ID)
	if row.Status != "scheduled" {
		t.Fatal("published early")
	}
	// Maintenance must pause background writes.
	e.server.maintenance.Store(true)
	if err := e.server.publishDue(context.Background(), due.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	e.server.maintenance.Store(false)
	row, _ = e.server.DB.GetPost(`WHERE id = ?`, p.ID)
	if row.Status != "scheduled" {
		t.Fatal("published during maintenance")
	}
	// A late/restarted worker catches up exactly once and retains the intended timestamp.
	if err := e.server.publishDue(context.Background(), due.Add(time.Hour)); err != nil {
		t.Fatal(err)
	}
	row, _ = e.server.DB.GetPost(`WHERE id = ?`, p.ID)
	publishedAt := row.UpdatedAt
	if row.Status != "published" || row.PublishAt != "" || row.CreatedAt != due.Format("2006-01-02 15:04:05") {
		t.Fatalf("bad publication: %+v", row)
	}
	if err := e.server.publishDue(context.Background(), due.Add(2*time.Hour)); err != nil {
		t.Fatal(err)
	}
	row, _ = e.server.DB.GetPost(`WHERE id = ?`, p.ID)
	if row.UpdatedAt != publishedAt {
		t.Fatal("publication repeated")
	}
	e.token = ""
	e.call("GET", "/api/v1/posts/queued-proof", nil, nil, 200)
	var detail struct {
		Note store.Note `json:"note"`
	}
	e.call("GET", fmt.Sprintf("/api/v1/notes/%d", n.ID), nil, &detail, 200)
	if detail.Note.Status != "published" || detail.Note.CreatedAt != due.Format("2006-01-02 15:04:05") {
		t.Fatalf("bad note: %+v", detail.Note)
	}
}

func TestRescheduleCancelAndValidation(t *testing.T) {
	e := newEnv(t)
	due := time.Now().UTC().Truncate(time.Second).Add(time.Hour)
	var n struct{ ID int64 }
	body := map[string]any{"contentMd": "保留正文", "status": "scheduled", "publishAt": due.Format(time.RFC3339)}
	e.call("POST", "/api/v1/admin/notes", body, &n, 201)
	url := fmt.Sprintf("/api/v1/admin/notes/%d", n.ID)
	for _, bad := range []string{"", "not-a-date", time.Now().Add(-time.Hour).Format(time.RFC3339)} {
		body["publishAt"] = bad
		e.call("PUT", url, body, nil, 400)
	}
	body["publishAt"] = due.Add(time.Hour).Format(time.RFC3339)
	e.call("PUT", url, body, nil, 200)
	if err := e.server.publishDue(context.Background(), due); err != nil {
		t.Fatal(err)
	}
	var note store.Note
	note = store.Note{}
	e.call("GET", url, nil, &note, 200)
	if note.Status != "scheduled" {
		t.Fatal("old due time still active")
	}
	body["status"] = "draft"
	body["publishAt"] = nil
	e.call("PUT", url, body, nil, 200)
	if err := e.server.publishDue(context.Background(), due.Add(4*time.Hour)); err != nil {
		t.Fatal(err)
	}
	note = store.Note{}
	e.call("GET", url, nil, &note, 200)
	if note.Status != "draft" || note.PublishAt != "" || note.ContentMd != "保留正文" {
		t.Fatalf("cancel discarded content: %+v", note)
	}
	// Local picker values are interpreted in the configured timezone.
	if _, err := e.server.Config.Save(config.Map{"timezone": "Asia/Hong_Kong"}); err != nil {
		t.Fatal(err)
	}
	body["status"] = "scheduled"
	body["publishAt"] = "2099-01-02T09:30:00"
	e.call("PUT", url, body, nil, 200)
	note = store.Note{}
	e.call("GET", url, nil, &note, 200)
	if note.PublishAt != "2099-01-02T01:30:00Z" {
		t.Fatal(note.PublishAt)
	}
	body["status"] = "published"
	e.call("PUT", url, body, nil, 200)
	note = store.Note{}
	e.call("GET", url, nil, &note, 200)
	if note.Status != "published" || note.PublishAt != "" {
		t.Fatal(note)
	}
	// Invalid daylight-saving wall times are rejected rather than silently shifted.
	e.server.Config.Save(config.Map{"timezone": "America/New_York"})
	body["status"] = "scheduled"
	body["publishAt"] = "2099-03-08T02:30:00"
	if _, ok := e.server.parsePublishTime("2027-03-14T02:30:00"); ok {
		t.Fatal("DST gap accepted")
	}
}

func TestNoteViewsPersistAndDeduplicate(t *testing.T) {
	e := newEnv(t)
	var n struct{ ID int64 }
	e.call("POST", "/api/v1/admin/notes", map[string]any{"contentMd": "浏览计数"}, &n, 201)
	e.token = ""
	url := fmt.Sprintf("/api/v1/notes/%d/view", n.ID)
	r := e.do("POST", url, nil, nil)
	var counted struct{ Views int64 }
	json.NewDecoder(r.Body).Decode(&counted)
	r.Body.Close()
	if r.StatusCode != 200 || counted.Views != 1 {
		t.Fatalf("first view: %d %+v", r.StatusCode, counted)
	}
	cookie := ""
	for _, c := range r.Cookies() {
		if c.Name == visitorCookie {
			cookie = c.Name + "=" + c.Value
		}
	}
	if cookie == "" {
		t.Fatal("missing signed visitor identity")
	}
	var wg sync.WaitGroup
	for range 8 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			res := e.do("POST", url, nil, map[string]string{"Cookie": cookie})
			io.Copy(io.Discard, res.Body)
			res.Body.Close()
		}()
	}
	wg.Wait()
	var views int64
	e.server.DB.QueryRow("SELECT views FROM notes WHERE id = ?", n.ID).Scan(&views)
	if views != 1 {
		t.Fatalf("refreshes counted repeatedly: %d", views)
	}
	r = e.do("POST", url, nil, nil)
	r.Body.Close()
	e.server.DB.QueryRow("SELECT views FROM notes WHERE id = ?", n.ID).Scan(&views)
	if views != 2 {
		t.Fatal(views)
	}
	if _, err := e.server.DB.Exec("UPDATE notes SET hidden = 1 WHERE id = ?", n.ID); err != nil {
		t.Fatal(err)
	}
	e.call("POST", url, nil, nil, 404)
	e.server.DB.QueryRow("SELECT views FROM notes WHERE id = ?", n.ID).Scan(&views)
	if views != 2 {
		t.Fatal("private view counted")
	}
}

func TestScheduledAuthorPermissionsAndHiddenState(t *testing.T) {
	e := newEnv(t)
	e.enableUsers(map[string]any{"enabled": true, "authors": map[string]any{"enabled": true, "directPublish": false}})
	var invite struct{ URL string }
	e.call("POST", "/api/v1/admin/invites", map[string]any{"role": "author", "email": "schedule@example.com"}, &invite, 201)
	code := invite.URL[strings.Index(invite.URL, "code=")+5:]
	token := e.signup(map[string]any{"email": "schedule@example.com", "name": "排期作者", "password": "author-pass-1", "invite": code})
	due := time.Now().UTC().Truncate(time.Second).Add(time.Hour)
	body := map[string]any{"slug": "author-scheduled", "title": "作者定时", "contentMd": "正文", "status": "scheduled", "publishAt": due.Format(time.RFC3339)}
	status, created := e.as(token, "POST", "/api/v1/admin/posts", body)
	if status != 201 {
		t.Fatalf("create %d", status)
	}
	id := int64(created["id"].(float64))
	post, _ := e.server.DB.GetPost(`WHERE id = ?`, id)
	if post.Status != "draft" || post.PublishAt != "" {
		t.Fatal("author bypassed review")
	}
	e.enableUsers(map[string]any{"authors": map[string]any{"directPublish": true}})
	if status, _ := e.as(token, "PUT", "/api/v1/admin/posts/"+itoa(id), body); status != 200 {
		t.Fatal(status)
	}
	e.enableUsers(map[string]any{"authors": map[string]any{"directPublish": false}})
	if err := e.server.publishDue(context.Background(), due.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	post, _ = e.server.DB.GetPost(`WHERE id = ?`, id)
	if post.Status != "draft" || post.PublishAt != "" {
		t.Fatal("revoked author still published")
	}
	var note struct{ ID int64 }
	e.call("POST", "/api/v1/admin/notes", map[string]any{"contentMd": "隐藏的定时内容", "status": "scheduled", "publishAt": due.Format(time.RFC3339)}, &note, 201)
	e.call("POST", "/api/v1/admin/notes/batch", map[string]any{"ids": []int64{note.ID}, "action": "hide"}, nil, 200)
	if err := e.server.publishDue(context.Background(), due.Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	notes, _ := e.server.DB.QueryNotes(`WHERE id = ?`, true, note.ID)
	if notes[0].Status != "published" || !notes[0].Hidden {
		t.Fatal("hidden state lost")
	}
	e.token = ""
	e.call("GET", "/api/v1/notes/"+itoa(note.ID), nil, nil, 404)
}

func TestExternalSchedulingScope(t *testing.T) {
	e := newEnv(t)
	for _, scope := range []string{"contrib", "full"} {
		var key struct{ Key string }
		e.call("POST", "/api/v1/admin/apikeys", map[string]string{"name": scope, "scope": scope}, &key, 201)
		headers := map[string]string{"Content-Type": "application/json", "X-Api-Key": key.Key}
		due := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
		raw, _ := json.Marshal(map[string]any{"slug": "scope-" + scope, "title": "Scope", "contentMd": "Body", "status": "scheduled", "publishAt": due})
		response := e.do("POST", "/api/v1/ext/posts", strings.NewReader(string(raw)), headers)
		var post store.Post
		json.NewDecoder(response.Body).Decode(&post)
		response.Body.Close()
		if response.StatusCode != 201 {
			t.Fatal(response.StatusCode)
		}
		want := "draft"
		if scope == "full" {
			want = "scheduled"
		}
		if post.Status != want {
			t.Fatalf("%s: %+v", scope, post)
		}
		if scope == "contrib" && post.PublishAt != "" {
			t.Fatal("contribution key queued a post")
		}
		raw, _ = json.Marshal(map[string]any{"contentMd": "Scheduled note", "status": "scheduled", "publishAt": due})
		response = e.do("POST", "/api/v1/ext/notes", strings.NewReader(string(raw)), headers)
		if scope == "contrib" {
			if response.StatusCode != 403 {
				t.Fatal(response.StatusCode)
			}
			response.Body.Close()
			continue
		}
		var note store.Note
		json.NewDecoder(response.Body).Decode(&note)
		response.Body.Close()
		if response.StatusCode != 201 || note.Status != "scheduled" {
			t.Fatal(note)
		}
		response = e.do("GET", "/api/v1/ext/notes/"+itoa(note.ID), nil, headers)
		if response.StatusCode != 200 {
			t.Fatal(response.StatusCode)
		}
		response.Body.Close()
	}
}
