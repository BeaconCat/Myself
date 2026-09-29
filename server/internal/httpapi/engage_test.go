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

// 评论点赞：已公开的评论可以点，待审的不行；列表带喜欢数与本人是否已点；删除评论时回应一并清理。
func TestCommentLikes(t *testing.T) {
	e := newEnv(t)
	e.enableUsers(map[string]any{"enabled": true, "comments": map[string]any{"enabled": true, "anonymous": true, "moderation": "all"}})
	h := map[string]string{"Content-Type": "application/json", "X-Requested-With": "myself"}

	// Demo 留言墙示例是已公开的访客留言
	var list struct {
		Items []struct {
			ID    int64 `json:"id"`
			Likes int   `json:"likes"`
			Liked bool  `json:"liked"`
		} `json:"items"`
	}
	e.call(http.MethodGet, "/api/v1/comments?target=guestbook&key=", nil, &list, http.StatusOK)
	if len(list.Items) == 0 {
		t.Fatal("no demo guestbook comments")
	}
	id := list.Items[0].ID
	res := e.do(http.MethodPost, "/api/v1/reactions", strings.NewReader(`{"target":"comment","id":`+itoa(id)+`,"kind":"like"}`), h)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("like comment: %d", res.StatusCode)
	}
	var vid string
	for _, c := range res.Cookies() {
		if c.Name == visitorCookie {
			vid = c.Value
		}
	}
	// 同一访客再取列表：计数 1 且标记已点
	req, _ := http.NewRequest(http.MethodGet, e.srv.URL+"/api/v1/comments?target=guestbook&key=", nil)
	req.AddCookie(&http.Cookie{Name: visitorCookie, Value: vid})
	lr, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	list.Items = nil
	json.NewDecoder(lr.Body).Decode(&list)
	lr.Body.Close()
	for _, it := range list.Items {
		if it.ID == id && (it.Likes != 1 || !it.Liked) {
			t.Fatalf("list likes: %+v", it)
		}
	}

	// 待审留言不能点赞
	// 以访客身份发（不带管理员令牌），审核规则为全部审核 → 待审
	var posted struct{ ID int64 }
	pr := e.do(http.MethodPost, "/api/v1/comments", strings.NewReader(`{"target":"guestbook","key":"","body":"待审","guestName":"路人"}`), h)
	json.NewDecoder(pr.Body).Decode(&posted)
	pr.Body.Close()
	if pr.StatusCode != http.StatusCreated || posted.ID == 0 {
		t.Fatalf("guest comment: %d", pr.StatusCode)
	}
	if res := e.do(http.MethodPost, "/api/v1/reactions", strings.NewReader(`{"target":"comment","id":`+itoa(posted.ID)+`,"kind":"like"}`), h); res.StatusCode != http.StatusNotFound {
		t.Fatalf("like pending comment: %d", res.StatusCode)
	}

	// 删除评论：回应一并清理
	e.call(http.MethodDelete, "/api/v1/admin/comments/"+itoa(id), nil, nil, http.StatusOK)
	var n int
	e.server.DB.QueryRow(`SELECT COUNT(*) FROM reactions WHERE target = 'comment' AND target_id = ?`, id).Scan(&n)
	if n != 0 {
		t.Fatalf("reactions left after delete: %d", n)
	}
}
