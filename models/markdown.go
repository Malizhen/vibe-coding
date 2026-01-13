package models

import (
	"html"
	"regexp"
	"strings"
)

// 简单的 Markdown 转 HTML
func MarkdownToHTML(md string) string {
	lines := strings.Split(md, "\n")
	var result []string
	inCodeBlock := false
	inList := false
	listType := ""

	for i := 0; i < len(lines); i++ {
		line := lines[i]

		// 代码块
		if strings.HasPrefix(line, "```") {
			if inCodeBlock {
				result = append(result, "</code></pre>")
				inCodeBlock = false
			} else {
				lang := strings.TrimPrefix(line, "```")
				if lang != "" {
					result = append(result, "<pre><code class=\"language-"+html.EscapeString(lang)+"\">")
				} else {
					result = append(result, "<pre><code>")
				}
				inCodeBlock = true
			}
			continue
		}

		if inCodeBlock {
			result = append(result, html.EscapeString(line))
			continue
		}

		// 关闭列表
		trimmed := strings.TrimSpace(line)
		isListItem := strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") ||
			regexp.MustCompile(`^\d+\. `).MatchString(trimmed)

		if inList && !isListItem && trimmed != "" {
			if listType == "ul" {
				result = append(result, "</ul>")
			} else {
				result = append(result, "</ol>")
			}
			inList = false
		}

		// 标题
		if strings.HasPrefix(line, "# ") {
			result = append(result, "<h1>"+processInline(strings.TrimPrefix(line, "# "))+"</h1>")
			continue
		}
		if strings.HasPrefix(line, "## ") {
			result = append(result, "<h2>"+processInline(strings.TrimPrefix(line, "## "))+"</h2>")
			continue
		}
		if strings.HasPrefix(line, "### ") {
			result = append(result, "<h3>"+processInline(strings.TrimPrefix(line, "### "))+"</h3>")
			continue
		}
		if strings.HasPrefix(line, "#### ") {
			result = append(result, "<h4>"+processInline(strings.TrimPrefix(line, "#### "))+"</h4>")
			continue
		}

		// 水平线
		if trimmed == "---" || trimmed == "***" || trimmed == "___" {
			result = append(result, "<hr>")
			continue
		}

		// 引用
		if strings.HasPrefix(line, "> ") {
			result = append(result, "<blockquote>"+processInline(strings.TrimPrefix(line, "> "))+"</blockquote>")
			continue
		}

		// 无序列表
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") {
			if !inList || listType != "ul" {
				if inList {
					result = append(result, "</ol>")
				}
				result = append(result, "<ul>")
				inList = true
				listType = "ul"
			}
			content := strings.TrimPrefix(strings.TrimPrefix(trimmed, "- "), "* ")
			result = append(result, "<li>"+processInline(content)+"</li>")
			continue
		}

		// 有序列表
		if regexp.MustCompile(`^\d+\. `).MatchString(trimmed) {
			if !inList || listType != "ol" {
				if inList {
					result = append(result, "</ul>")
				}
				result = append(result, "<ol>")
				inList = true
				listType = "ol"
			}
			content := regexp.MustCompile(`^\d+\. `).ReplaceAllString(trimmed, "")
			result = append(result, "<li>"+processInline(content)+"</li>")
			continue
		}

		// 空行
		if trimmed == "" {
			continue
		}

		// 普通段落
		result = append(result, "<p>"+processInline(line)+"</p>")
	}

	// 关闭未关闭的标签
	if inCodeBlock {
		result = append(result, "</code></pre>")
	}
	if inList {
		if listType == "ul" {
			result = append(result, "</ul>")
		} else {
			result = append(result, "</ol>")
		}
	}

	return strings.Join(result, "\n")
}

// 处理行内元素
func processInline(text string) string {
	// 转义 HTML
	text = html.EscapeString(text)

	// 图片 ![alt](url)
	imgRe := regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)
	text = imgRe.ReplaceAllString(text, `<img src="$2" alt="$1" loading="lazy">`)

	// 链接 [text](url)
	linkRe := regexp.MustCompile(`\[([^\]]+)\]\(([^)]+)\)`)
	text = linkRe.ReplaceAllString(text, `<a href="$2" target="_blank" rel="noopener">$1</a>`)

	// 粗体 **text** 或 __text__
	boldRe := regexp.MustCompile(`\*\*([^*]+)\*\*|__([^_]+)__`)
	text = boldRe.ReplaceAllStringFunc(text, func(s string) string {
		s = strings.Trim(s, "*_")
		return "<strong>" + s + "</strong>"
	})

	// 斜体 *text* 或 _text_
	italicRe := regexp.MustCompile(`\*([^*]+)\*|_([^_]+)_`)
	text = italicRe.ReplaceAllStringFunc(text, func(s string) string {
		s = strings.Trim(s, "*_")
		return "<em>" + s + "</em>"
	})

	// 行内代码 `code`
	codeRe := regexp.MustCompile("`([^`]+)`")
	text = codeRe.ReplaceAllString(text, `<code>$1</code>`)

	return text
}
