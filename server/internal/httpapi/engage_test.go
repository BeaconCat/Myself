package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestReactionsToggleAndEngage(t *testing.T) {
	e := newEnv(t)
	var notes struct {
		Items []struct {
			ID int64 `json:"id"`
		} `json:"items"`
	}
	e.call(http.MethodGet, "/api/v1/notes?pageSize=2", nil, &notes, http.StatusOK)
	id := notes.Items[0].ID
	body := `{"target":"note","id":` + itoa(id) + `,"kind":"like"}`
	// 无自定义头：拒绝（防跨站刷量）
	if res := e.do(http.MethodPost, "/api/v1/reactions", strings.NewReader(body), map[string]string{"Content-Type": "application/json"}); res.StatusCode != http.StatusForbidden {
		t.Fatalf("reaction without header: %d", res.StatusCode)
	}
	h := map[string]string{"Content-Type": "application/json", "X-Requested-With": "myself"}
	res := e.do(http.MethodPost, "/api/v1/reactions", strings.NewReader(body), h)
	var sum engageSummary
	json.NewDecoder(res.Body).Decode(&sum)
	if sum.Reactions["like"] != 1 || len(sum.Mine) != 1 {
		t.Fatalf("first like: %+v", sum)
	}
	var vid string
	for _, c := range res.Cookies() {
		if c.Name == visitorCookie {
			vid = c.Value
		}
	}
	if vid == "" {
		t.Fatal("visitor cookie not issued")
	}
	// 同一访客再点一次：取消
	h["Cookie"] = visitorCookie + "=" + vid
	res = e.do(http.MethodPost, "/api/v1/reactions", strings.NewReader(body), h)
	sum = engageSummary{}
	json.NewDecoder(res.Body).Decode(&sum)
	if sum.Reactions["like"] != 0 || len(sum.Mine) != 0 {
		t.Fatalf("toggle off: %+v", sum)
	}
	// 另一位访客点赞后，批量摘要可见
	e.do(http.MethodPost, "/api/v1/reactions", strings.NewReader(strings.Replace(body, "like", "spark", 1)), map[string]string{"Content-Type": "application/json", "X-Requested-With": "myself"})
	var all map[string]engageSummary
	e.call(http.MethodGet, "/api/v1/engage?target=note&ids="+itoa(id)+","+itoa(notes.Items[1].ID), nil, &all, http.StatusOK)
	if all[itoa(id)].Reactions["spark"] != 1 {
		t.Fatalf("engage: %+v", all)
	}
	// 单条随想
	var one struct {
		Note  struct{ ID int64 } `json:"note"`
		Older int64              `json:"older"`
	}
	e.call(http.MethodGet, "/api/v1/notes/"+itoa(id), nil, &one, http.StatusOK)
	if one.Note.ID != id || one.Older == 0 {
		t.Fatalf("note detail: %+v", one)
	}
	// 关闭回应
	e.enableUsers(map[string]any{"reactions": false})
	if res := e.do(http.MethodPost, "/api/v1/reactions", strings.NewReader(body), h); res.StatusCode != http.StatusForbidden {
		t.Fatalf("reactions off: %d", res.StatusCode)
	}
}
