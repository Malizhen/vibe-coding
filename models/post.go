package models

import (
	"bufio"
	"bytes"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

type Post struct {
	Slug        string
	Title       string
	Content     string
	HTMLContent string
	Excerpt     string
	Date        time.Time
	Category    string
	Tags        []string
	ReadingTime int // 分钟
	Published   bool
}

// 解析 Markdown 文件的 frontmatter
func ParsePost(filename string) (*Post, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	post := &Post{
		Slug:      strings.TrimSuffix(filepath.Base(filename), ".md"),
		Published: true,
	}

	lines := strings.Split(string(content), "\n")
	inFrontmatter := false
	frontmatterEnd := 0

	for i, line := range lines {
		if i == 0 && strings.TrimSpace(line) == "---" {
			inFrontmatter = true
			continue
		}
		if inFrontmatter && strings.TrimSpace(line) == "---" {
			frontmatterEnd = i + 1
			break
		}
		if inFrontmatter {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				key := strings.TrimSpace(parts[0])
				value := strings.TrimSpace(parts[1])
				value = strings.Trim(value, "\"'")

				switch key {
				case "title":
					post.Title = value
				case "date":
					post.Date, _ = time.Parse("2006-01-02", value)
				case "category":
					post.Category = value
				case "tags":
					// 解析 [tag1, tag2] 格式
					value = strings.Trim(value, "[]")
					for _, tag := range strings.Split(value, ",") {
						tag = strings.TrimSpace(tag)
						tag = strings.Trim(tag, "\"'")
						if tag != "" {
							post.Tags = append(post.Tags, tag)
						}
					}
				case "published":
					post.Published = value == "true"
				}
			}
		}
	}

	// 获取正文内容
	if frontmatterEnd > 0 && frontmatterEnd < len(lines) {
		post.Content = strings.Join(lines[frontmatterEnd:], "\n")
	} else {
		post.Content = string(content)
	}

	post.Content = strings.TrimSpace(post.Content)
	post.HTMLContent = MarkdownToHTML(post.Content)
	post.Excerpt = generateExcerpt(post.Content, 150)
	post.ReadingTime = calculateReadingTime(post.Content)

	return post, nil
}

// 获取所有文章
func GetAllPosts() ([]*Post, error) {
	var posts []*Post

	files, err := filepath.Glob("content/posts/*.md")
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		post, err := ParsePost(file)
		if err != nil {
			continue
		}
		if post.Published {
			posts = append(posts, post)
		}
	}

	// 按日期倒序排列
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].Date.After(posts[j].Date)
	})

	return posts, nil
}

// 获取单篇文章
func GetPost(slug string) (*Post, error) {
	filename := filepath.Join("content/posts", slug+".md")
	return ParsePost(filename)
}

// 按分类获取文章
func GetPostsByCategory(category string) ([]*Post, error) {
	allPosts, err := GetAllPosts()
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, post := range allPosts {
		if strings.EqualFold(post.Category, category) {
			posts = append(posts, post)
		}
	}
	return posts, nil
}

// 按标签获取文章
func GetPostsByTag(tag string) ([]*Post, error) {
	allPosts, err := GetAllPosts()
	if err != nil {
		return nil, err
	}

	var posts []*Post
	for _, post := range allPosts {
		for _, t := range post.Tags {
			if strings.EqualFold(t, tag) {
				posts = append(posts, post)
				break
			}
		}
	}
	return posts, nil
}

// 获取所有分类
func GetAllCategories() ([]string, error) {
	posts, err := GetAllPosts()
	if err != nil {
		return nil, err
	}

	categoryMap := make(map[string]bool)
	for _, post := range posts {
		if post.Category != "" {
			categoryMap[post.Category] = true
		}
	}

	var categories []string
	for cat := range categoryMap {
		categories = append(categories, cat)
	}
	sort.Strings(categories)
	return categories, nil
}

// 获取所有标签
func GetAllTags() ([]string, error) {
	posts, err := GetAllPosts()
	if err != nil {
		return nil, err
	}

	tagMap := make(map[string]bool)
	for _, post := range posts {
		for _, tag := range post.Tags {
			tagMap[tag] = true
		}
	}

	var tags []string
	for tag := range tagMap {
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, nil
}

// 保存文章
func SavePost(post *Post) error {
	var buf bytes.Buffer
	buf.WriteString("---\n")
	buf.WriteString("title: \"" + post.Title + "\"\n")
	buf.WriteString("date: " + post.Date.Format("2006-01-02") + "\n")
	if post.Category != "" {
		buf.WriteString("category: " + post.Category + "\n")
	}
	if len(post.Tags) > 0 {
		buf.WriteString("tags: [")
		for i, tag := range post.Tags {
			if i > 0 {
				buf.WriteString(", ")
			}
			buf.WriteString("\"" + tag + "\"")
		}
		buf.WriteString("]\n")
	}
	buf.WriteString("published: ")
	if post.Published {
		buf.WriteString("true\n")
	} else {
		buf.WriteString("false\n")
	}
	buf.WriteString("---\n\n")
	buf.WriteString(post.Content)

	filename := filepath.Join("content/posts", post.Slug+".md")
	return os.WriteFile(filename, buf.Bytes(), 0644)
}

// 删除文章
func DeletePost(slug string) error {
	filename := filepath.Join("content/posts", slug+".md")
	return os.Remove(filename)
}

// 生成摘要
func generateExcerpt(content string, maxLen int) string {
	// 移除 Markdown 标记
	re := regexp.MustCompile(`[#*_\[\]()` + "`" + `]`)
	plain := re.ReplaceAllString(content, "")

	// 取第一段
	scanner := bufio.NewScanner(strings.NewReader(plain))
	var excerpt string
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" {
			excerpt = line
			break
		}
	}

	if utf8.RuneCountInString(excerpt) > maxLen {
		runes := []rune(excerpt)
		excerpt = string(runes[:maxLen]) + "..."
	}

	return excerpt
}

// 计算阅读时间
func calculateReadingTime(content string) int {
	words := len(strings.Fields(content))
	// 中文按字符计算，平均阅读速度约 400 字/分钟
	chars := utf8.RuneCountInString(content)
	// 混合计算
	minutes := (words + chars/2) / 400
	if minutes < 1 {
		minutes = 1
	}
	return minutes
}
