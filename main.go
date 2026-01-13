package main

import (
	"log"
	"myblog/handlers"
	"net/http"
	"os"
)

func main() {
	// 初始化处理器
	h := handlers.New()

	// 静态文件
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	// 公开路由
	http.HandleFunc("/", h.Home)
	http.HandleFunc("/post/", h.Post)
	http.HandleFunc("/category/", h.Category)
	http.HandleFunc("/tag/", h.Tag)
	http.HandleFunc("/about", h.About)
	http.HandleFunc("/contact", h.Contact)
	http.HandleFunc("/rss", h.RSS)

	// 评论
	http.HandleFunc("/api/comment", h.SubmitComment)

	// 管理后台
	http.HandleFunc("/admin", h.AdminLogin)
	http.HandleFunc("/admin/dashboard", h.AdminDashboard)
	http.HandleFunc("/admin/posts", h.AdminPosts)
	http.HandleFunc("/admin/posts/new", h.AdminNewPost)
	http.HandleFunc("/admin/posts/edit/", h.AdminEditPost)
	http.HandleFunc("/admin/posts/delete/", h.AdminDeletePost)
	http.HandleFunc("/admin/comments", h.AdminComments)
	http.HandleFunc("/admin/comments/approve/", h.AdminApproveComment)
	http.HandleFunc("/admin/comments/delete/", h.AdminDeleteComment)
	http.HandleFunc("/admin/logout", h.AdminLogout)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("博客服务器启动在 http://localhost:%s", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
