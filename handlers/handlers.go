package handlers

import (
	"crypto/rand"
	"encoding/hex"
	"html/template"
	"myblog/models"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

type Handler struct {
	templates    *template.Template
	sessions     map[string]time.Time
	sessionMutex sync.RWMutex
	adminPass    string
}

func New() *Handler {
	funcMap := template.FuncMap{
		"safeHTML": func(s string) template.HTML {
			return template.HTML(s)
		},
		"formatDate": func(t time.Time) string {
			return t.Format("2006年01月02日")
		},
		"formatDateShort": func(t time.Time) string {
			return t.Format("2006-01-02")
		},
		"join": func(arr []string, sep string) string {
			return strings.Join(arr, sep)
		},
	}

	templates := template.Must(template.New("").Funcs(funcMap).ParseGlob("templates/*.html"))
	template.Must(templates.ParseGlob("templates/admin/*.html"))

	adminPass := os.Getenv("ADMIN_PASSWORD")
	if adminPass == "" {
		adminPass = "admin123" // 默认密码，生产环境请修改
	}

	return &Handler{
		templates: templates,
		sessions:  make(map[string]time.Time),
		adminPass: adminPass,
	}
}

// 首页
func (h *Handler) Home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}

	posts, _ := models.GetAllPosts()
	categories, _ := models.GetAllCategories()
	tags, _ := models.GetAllTags()

	data := map[string]interface{}{
		"Title":      "首页",
		"Posts":      posts,
		"Categories": categories,
		"Tags":       tags,
	}

	h.render(w, "home.html", data)
}

// 文章详情
func (h *Handler) Post(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/post/")
	if slug == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	post, err := models.GetPost(slug)
	if err != nil {
		http.NotFound(w, r)
		return
	}

	comments, _ := models.GetApprovedComments(slug)
	categories, _ := models.GetAllCategories()
	tags, _ := models.GetAllTags()

	data := map[string]interface{}{
		"Title":      post.Title,
		"Post":       post,
		"Comments":   comments,
		"Categories": categories,
		"Tags":       tags,
	}

	h.render(w, "post.html", data)
}

// 分类页面
func (h *Handler) Category(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimPrefix(r.URL.Path, "/category/")
	if category == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	posts, _ := models.GetPostsByCategory(category)
	categories, _ := models.GetAllCategories()
	tags, _ := models.GetAllTags()

	data := map[string]interface{}{
		"Title":      "分类: " + category,
		"Category":   category,
		"Posts":      posts,
		"Categories": categories,
		"Tags":       tags,
	}

	h.render(w, "category.html", data)
}

// 标签页面
func (h *Handler) Tag(w http.ResponseWriter, r *http.Request) {
	tag := strings.TrimPrefix(r.URL.Path, "/tag/")
	if tag == "" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	posts, _ := models.GetPostsByTag(tag)
	categories, _ := models.GetAllCategories()
	tags, _ := models.GetAllTags()

	data := map[string]interface{}{
		"Title":      "标签: " + tag,
		"Tag":        tag,
		"Posts":      posts,
		"Categories": categories,
		"Tags":       tags,
	}

	h.render(w, "tag.html", data)
}

// 关于页面
func (h *Handler) About(w http.ResponseWriter, r *http.Request) {
	categories, _ := models.GetAllCategories()
	tags, _ := models.GetAllTags()

	data := map[string]interface{}{
		"Title":      "关于",
		"Categories": categories,
		"Tags":       tags,
	}

	h.render(w, "about.html", data)
}

// 联系页面
func (h *Handler) Contact(w http.ResponseWriter, r *http.Request) {
	categories, _ := models.GetAllCategories()
	tags, _ := models.GetAllTags()

	message := ""
	if r.Method == http.MethodPost {
		// 处理表单提交
		name := r.FormValue("name")
		email := r.FormValue("email")
		content := r.FormValue("message")

		if name != "" && email != "" && content != "" {
			// 这里可以发送邮件或保存到文件
			// 简单起见，我们只显示成功消息
			message = "感谢您的留言！我会尽快回复。"
		}
	}

	data := map[string]interface{}{
		"Title":      "联系我",
		"Message":    message,
		"Categories": categories,
		"Tags":       tags,
	}

	h.render(w, "contact.html", data)
}

// 提交评论
func (h *Handler) SubmitComment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	postSlug := r.FormValue("post_slug")
	author := strings.TrimSpace(r.FormValue("author"))
	email := strings.TrimSpace(r.FormValue("email"))
	content := strings.TrimSpace(r.FormValue("content"))

	if postSlug == "" || author == "" || content == "" {
		http.Error(w, "缺少必填字段", http.StatusBadRequest)
		return
	}

	comment := models.Comment{
		PostSlug: postSlug,
		Author:   author,
		Email:    email,
		Content:  content,
	}

	if err := models.AddComment(comment); err != nil {
		http.Error(w, "保存评论失败", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/post/"+postSlug+"?comment=success", http.StatusFound)
}

// RSS 订阅
func (h *Handler) RSS(w http.ResponseWriter, r *http.Request) {
	posts, _ := models.GetAllPosts()

	w.Header().Set("Content-Type", "application/xml; charset=utf-8")

	siteURL := "http://localhost:8080"
	if host := r.Host; host != "" {
		if r.TLS != nil {
			siteURL = "https://" + host
		} else {
			siteURL = "http://" + host
		}
	}

	xml := `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0" xmlns:atom="http://www.w3.org/2005/Atom">
<channel>
<title>我的博客</title>
<link>` + siteURL + `</link>
<description>分享想法与见解</description>
<language>zh-CN</language>
<atom:link href="` + siteURL + `/rss" rel="self" type="application/rss+xml"/>
`

	for _, post := range posts {
		if len(posts) > 20 {
			break
		}
		xml += `<item>
<title>` + escapeXML(post.Title) + `</title>
<link>` + siteURL + `/post/` + post.Slug + `</link>
<description>` + escapeXML(post.Excerpt) + `</description>
<pubDate>` + post.Date.Format(time.RFC1123Z) + `</pubDate>
<guid>` + siteURL + `/post/` + post.Slug + `</guid>
</item>
`
	}

	xml += `</channel>
</rss>`

	w.Write([]byte(xml))
}

// 渲染模板
func (h *Handler) render(w http.ResponseWriter, name string, data interface{}) {
	if err := h.templates.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// 生成会话 ID
func generateSessionID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// XML 转义
func escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}
