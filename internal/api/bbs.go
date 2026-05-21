package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type bbsPost struct {
	ID        int        `json:"id"`
	Title     string     `json:"title"`
	Content   string     `json:"content"`
	Author    string     `json:"author"`
	Tags      []string   `json:"tags,omitempty"`
	CreatedAt string     `json:"created_at"`
	UpdatedAt string     `json:"updated_at"`
	Replies   []bbsReply `json:"replies,omitempty"`
}

type bbsReply struct {
	ID        int    `json:"id"`
	Author    string `json:"author"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
}

type bbsState struct {
	BoardKey  string    `json:"board_key"`
	NextID    int       `json:"next_id"`
	NextReply int       `json:"next_reply_id"`
	Posts     []bbsPost `json:"posts"`
}

var bbsMu sync.Mutex

func (s *Server) bbsDir() string  { return filepath.Join(s.CfgStore.Root, "data") }
func (s *Server) bbsPath() string { return filepath.Join(s.bbsDir(), "bbs.json") }

func (s *Server) loadBBS() (bbsState, error) {
	st := bbsState{BoardKey: "ga-team", NextID: 1, NextReply: 1, Posts: []bbsPost{}}
	data, err := os.ReadFile(s.bbsPath())
	if err != nil {
		if os.IsNotExist(err) {
			return st, nil
		}
		return st, err
	}
	if len(strings.TrimSpace(string(data))) == 0 {
		return st, nil
	}
	if err := json.Unmarshal(data, &st); err != nil {
		return st, err
	}
	if st.BoardKey == "" {
		st.BoardKey = "ga-team"
	}
	if st.NextID <= 0 {
		st.NextID = 1
	}
	if st.NextReply <= 0 {
		st.NextReply = 1
	}
	if st.Posts == nil {
		st.Posts = []bbsPost{}
	}
	for _, p := range st.Posts {
		if p.ID >= st.NextID {
			st.NextID = p.ID + 1
		}
		for _, rp := range p.Replies {
			if rp.ID >= st.NextReply {
				st.NextReply = rp.ID + 1
			}
		}
	}
	return st, nil
}

func (s *Server) saveBBS(st bbsState) error {
	if err := os.MkdirAll(s.bbsDir(), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.bbsPath(), data, 0644)
}

func bbsClientKey(r *http.Request) string {
	if k := r.Header.Get("X-API-Key"); k != "" {
		return k
	}
	return r.URL.Query().Get("key")
}

func bbsAllowed(st bbsState, r *http.Request) bool {
	k := strings.TrimSpace(st.BoardKey)
	return k == "" || bbsClientKey(r) == k
}

func (s *Server) bbsStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		bad(w, 405, "method not allowed")
		return
	}
	bbsMu.Lock()
	defer bbsMu.Unlock()
	st, err := s.loadBBS()
	if err != nil {
		bad(w, 500, err.Error())
		return
	}
	baseURL := fmt.Sprintf("http://%s:%d", s.CfgStore.Cfg.Host, s.CfgStore.Cfg.Port)
	writeJSON(w, map[string]any{"enabled": true, "base_url": baseURL, "board_key": st.BoardKey, "posts": len(st.Posts), "path": s.bbsPath(), "readme": baseURL + "/readme?key=" + st.BoardKey})
}

func (s *Server) bbsPosts(w http.ResponseWriter, r *http.Request)       { s.bbsPostsCore(w, r, true) }
func (s *Server) bbsPostsCompat(w http.ResponseWriter, r *http.Request) { s.bbsPostsCore(w, r, false) }

func (s *Server) bbsPostsCore(w http.ResponseWriter, r *http.Request, admin bool) {
	bbsMu.Lock()
	defer bbsMu.Unlock()
	st, err := s.loadBBS()
	if err != nil {
		bad(w, 500, err.Error())
		return
	}
	if !admin && !bbsAllowed(st, r) {
		bad(w, 403, "invalid board key")
		return
	}
	if r.Method == http.MethodGet {
		limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
		if limit <= 0 || limit > 200 {
			limit = 50
		}
		posts := append([]bbsPost(nil), st.Posts...)
		sort.Slice(posts, func(i, j int) bool { return posts[i].ID > posts[j].ID })
		if len(posts) > limit {
			posts = posts[:limit]
		}
		writeJSON(w, posts)
		return
	}
	if r.Method == http.MethodPost {
		var p struct {
			Title, Content, Author string
			Tags                   []string `json:"tags"`
		}
		if err := decode(r, &p); err != nil {
			bad(w, 400, err.Error())
			return
		}
		if strings.TrimSpace(p.Title) == "" || strings.TrimSpace(p.Content) == "" {
			bad(w, 400, "title and content required")
			return
		}
		now := time.Now().Format(time.RFC3339)
		post := bbsPost{ID: st.NextID, Title: strings.TrimSpace(p.Title), Content: p.Content, Author: strings.TrimSpace(p.Author), Tags: p.Tags, CreatedAt: now, UpdatedAt: now, Replies: []bbsReply{}}
		if post.Author == "" {
			post.Author = "admin"
		}
		st.NextID++
		st.Posts = append(st.Posts, post)
		if err := s.saveBBS(st); err != nil {
			bad(w, 500, err.Error())
			return
		}
		writeJSON(w, post)
		return
	}
	bad(w, 405, "method not allowed")
}

func (s *Server) bbsPost(w http.ResponseWriter, r *http.Request)       { s.bbsPostCore(w, r, true) }
func (s *Server) bbsPostCompat(w http.ResponseWriter, r *http.Request) { s.bbsPostCore(w, r, false) }

func (s *Server) bbsPostCore(w http.ResponseWriter, r *http.Request, admin bool) {
	if r.Method != http.MethodGet {
		bad(w, 405, "method not allowed")
		return
	}
	bbsMu.Lock()
	defer bbsMu.Unlock()
	st, err := s.loadBBS()
	if err != nil {
		bad(w, 500, err.Error())
		return
	}
	if !admin && !bbsAllowed(st, r) {
		bad(w, 403, "invalid board key")
		return
	}
	id, _ := strconv.Atoi(r.URL.Query().Get("id"))
	if id <= 0 {
		bad(w, 400, "id required")
		return
	}
	for _, p := range st.Posts {
		if p.ID == id {
			writeJSON(w, p)
			return
		}
	}
	bad(w, 404, "post not found")
}

func (s *Server) bbsReply(w http.ResponseWriter, r *http.Request)       { s.bbsReplyCore(w, r, true) }
func (s *Server) bbsReplyCompat(w http.ResponseWriter, r *http.Request) { s.bbsReplyCore(w, r, false) }

func (s *Server) bbsReplyCore(w http.ResponseWriter, r *http.Request, admin bool) {
	if r.Method != http.MethodPost {
		bad(w, 405, "method not allowed")
		return
	}
	bbsMu.Lock()
	defer bbsMu.Unlock()
	st, err := s.loadBBS()
	if err != nil {
		bad(w, 500, err.Error())
		return
	}
	if !admin && !bbsAllowed(st, r) {
		bad(w, 403, "invalid board key")
		return
	}
	var req struct {
		PostID  int    `json:"post_id"`
		ID      int    `json:"id"`
		Author  string `json:"author"`
		Content string `json:"content"`
	}
	if err := decode(r, &req); err != nil {
		bad(w, 400, err.Error())
		return
	}
	if req.PostID == 0 {
		req.PostID = req.ID
	}
	if req.PostID <= 0 || strings.TrimSpace(req.Content) == "" {
		bad(w, 400, "post_id and content required")
		return
	}
	for i := range st.Posts {
		if st.Posts[i].ID == req.PostID {
			if strings.TrimSpace(req.Author) == "" {
				req.Author = "agent"
			}
			rep := bbsReply{ID: st.NextReply, Author: strings.TrimSpace(req.Author), Content: req.Content, CreatedAt: time.Now().Format(time.RFC3339)}
			st.NextReply++
			st.Posts[i].Replies = append(st.Posts[i].Replies, rep)
			st.Posts[i].UpdatedAt = rep.CreatedAt
			if err := s.saveBBS(st); err != nil {
				bad(w, 500, err.Error())
				return
			}
			writeJSON(w, rep)
			return
		}
	}
	bad(w, 404, "post not found")
}

func (s *Server) bbsReadme(w http.ResponseWriter, r *http.Request) { s.bbsReadmeCore(w, r, true) }
func (s *Server) bbsReadmeCompat(w http.ResponseWriter, r *http.Request) {
	s.bbsReadmeCore(w, r, false)
}

func (s *Server) bbsReadmeCore(w http.ResponseWriter, r *http.Request, admin bool) {
	bbsMu.Lock()
	st, err := s.loadBBS()
	bbsMu.Unlock()
	if err != nil {
		bad(w, 500, err.Error())
		return
	}
	if !admin && !bbsAllowed(st, r) {
		bad(w, 403, "invalid board key")
		return
	}
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	_, _ = w.Write([]byte("GenericAgent Admin Built-in BBS\n\nGET /posts?limit=10&key=BOARD_KEY  list newest posts\nGET /post?id=1&key=BOARD_KEY       read one post with replies\nPOST /reply?key=BOARD_KEY           JSON {post_id, author, content}\nPOST /posts?key=BOARD_KEY           JSON {title, content, author, tags}\n\nGA worker: set reflect/agent_team_setting.json base_url to this Admin URL and board_key to the shown key.\n"))
}
