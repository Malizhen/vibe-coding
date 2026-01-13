package handlers

import (
	"myblog/models"
	"net/http"
	"strings"
	"time"
)

// 管理员登录
func (h *Handler) AdminLogin(w http.ResponseWriter, r *http.Request) {
	if h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
		return
	}

	errorMsg := ""
	if r.Method == http.MethodPost {
		password := r.FormValue("password")
		if password == h.adminPass {
			sessionID := generateSessionID()
			h.sessionMutex.Lock()
			h.sessions[sessionID] = time.Now().Add(24 * time.Hour)
			h.sessionMutex.Unlock()

			http.SetCookie(w, &http.Cookie{
				Name:     "session",
				Value:    sessionID,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   86400,
			})

			http.Redirect(w, r, "/admin/dashboard", http.StatusFound)
			return
		}
		errorMsg = "密码错误"
	}

	h.render(w, "admin_login.html", map[string]interface{}{
		"Title": "管理员登录",
		"Error": errorMsg,
	})
}

// 管理员登出
func (h *Handler) AdminLogout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err == nil {
		h.sessionMutex.Lock()
		delete(h.sessions, cookie.Value)
		h.sessionMutex.Unlock()
	}

	http.SetCookie(w, &http.Cookie{
		Name:   "session",
		Value:  "",
		Path:   "/",
		MaxAge: -1,
	})

	http.Redirect(w, r, "/admin", http.StatusFound)
}

// 仪表盘
func (h *Handler) AdminDashboard(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	posts, _ := models.GetAllPosts()
	pendingComments, _ := models.GetPendingComments()

	h.render(w, "admin_dashboard.html", map[string]interface{}{
		"Title":           "管理后台",
		"PostCount":       len(posts),
		"PendingComments": len(pendingComments),
	})
}

// 文章管理
func (h *Handler) AdminPosts(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	posts, _ := models.GetAllPosts()

	h.render(w, "admin_posts.html", map[string]interface{}{
		"Title": "文章管理",
		"Posts": posts,
	})
}

// 新建文章
func (h *Handler) AdminNewPost(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	if r.Method == http.MethodPost {
		post := &models.Post{
			Slug:      r.FormValue("slug"),
			Title:     r.FormValue("title"),
			Content:   r.FormValue("content"),
			Category:  r.FormValue("category"),
			Date:      time.Now(),
			Published: r.FormValue("published") == "on",
		}

		// 解析标签
		tags := r.FormValue("tags")
		for _, tag := range strings.Split(tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				post.Tags = append(post.Tags, tag)
			}
		}

		// 生成 slug
		if post.Slug == "" {
			post.Slug = generateSlug(post.Title)
		}

		if err := models.SavePost(post); err != nil {
			h.render(w, "admin_edit.html", map[string]interface{}{
				"Title": "新建文章",
				"Post":  post,
				"Error": "保存失败: " + err.Error(),
			})
			return
		}

		http.Redirect(w, r, "/admin/posts", http.StatusFound)
		return
	}

	h.render(w, "admin_edit.html", map[string]interface{}{
		"Title": "新建文章",
		"Post":  &models.Post{Date: time.Now(), Published: true},
		"IsNew": true,
	})
}

// 编辑文章
func (h *Handler) AdminEditPost(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/admin/posts/edit/")

	if r.Method == http.MethodPost {
		post := &models.Post{
			Slug:      slug,
			Title:     r.FormValue("title"),
			Content:   r.FormValue("content"),
			Category:  r.FormValue("category"),
			Published: r.FormValue("published") == "on",
		}

		// 解析日期
		if dateStr := r.FormValue("date"); dateStr != "" {
			post.Date, _ = time.Parse("2006-01-02", dateStr)
		} else {
			post.Date = time.Now()
		}

		// 解析标签
		tags := r.FormValue("tags")
		for _, tag := range strings.Split(tags, ",") {
			tag = strings.TrimSpace(tag)
			if tag != "" {
				post.Tags = append(post.Tags, tag)
			}
		}

		if err := models.SavePost(post); err != nil {
			h.render(w, "admin_edit.html", map[string]interface{}{
				"Title": "编辑文章",
				"Post":  post,
				"Error": "保存失败: " + err.Error(),
			})
			return
		}

		http.Redirect(w, r, "/admin/posts", http.StatusFound)
		return
	}

	post, err := models.GetPost(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	h.render(w, "admin_edit.html", map[string]interface{}{
		"Title": "编辑文章",
		"Post":  post,
	})
}

// 删除文章
func (h *Handler) AdminDeletePost(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	slug := strings.TrimPrefix(r.URL.Path, "/admin/posts/delete/")
	models.DeletePost(slug)

	http.Redirect(w, r, "/admin/posts", http.StatusFound)
}

// 评论管理
func (h *Handler) AdminComments(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	comments, _ := models.GetAllComments()

	h.render(w, "admin_comments.html", map[string]interface{}{
		"Title":    "评论管理",
		"Comments": comments,
	})
}

// 审核评论
func (h *Handler) AdminApproveComment(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/comments/approve/")
	models.ApproveComment(id)

	http.Redirect(w, r, "/admin/comments", http.StatusFound)
}

// 删除评论
func (h *Handler) AdminDeleteComment(w http.ResponseWriter, r *http.Request) {
	if !h.isAuthenticated(r) {
		http.Redirect(w, r, "/admin", http.StatusFound)
		return
	}

	id := strings.TrimPrefix(r.URL.Path, "/admin/comments/delete/")
	models.DeleteComment(id)

	http.Redirect(w, r, "/admin/comments", http.StatusFound)
}

// 检查是否已认证
func (h *Handler) isAuthenticated(r *http.Request) bool {
	cookie, err := r.Cookie("session")
	if err != nil {
		return false
	}

	h.sessionMutex.RLock()
	expiry, exists := h.sessions[cookie.Value]
	h.sessionMutex.RUnlock()

	if !exists || time.Now().After(expiry) {
		return false
	}

	return true
}

// 生成 slug
func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	// 移除特殊字符，保留中文
	return slug
}
