package models

import (
	"encoding/json"
	"os"
	"sort"
	"time"
)

type Comment struct {
	ID        string    `json:"id"`
	PostSlug  string    `json:"post_slug"`
	Author    string    `json:"author"`
	Email     string    `json:"email"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
	Approved  bool      `json:"approved"`
}

const commentsFile = "data/comments.json"

// 加载所有评论
func LoadComments() ([]Comment, error) {
	data, err := os.ReadFile(commentsFile)
	if err != nil {
		if os.IsNotExist(err) {
			return []Comment{}, nil
		}
		return nil, err
	}

	var comments []Comment
	if err := json.Unmarshal(data, &comments); err != nil {
		return nil, err
	}

	return comments, nil
}

// 保存所有评论
func SaveComments(comments []Comment) error {
	data, err := json.MarshalIndent(comments, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(commentsFile, data, 0644)
}

// 添加评论
func AddComment(comment Comment) error {
	comments, err := LoadComments()
	if err != nil {
		return err
	}

	comment.ID = time.Now().Format("20060102150405") + "-" + comment.PostSlug
	comment.CreatedAt = time.Now()
	comment.Approved = false

	comments = append(comments, comment)
	return SaveComments(comments)
}

// 获取文章的已审核评论
func GetApprovedComments(postSlug string) ([]Comment, error) {
	comments, err := LoadComments()
	if err != nil {
		return nil, err
	}

	var result []Comment
	for _, c := range comments {
		if c.PostSlug == postSlug && c.Approved {
			result = append(result, c)
		}
	}

	// 按时间排序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.Before(result[j].CreatedAt)
	})

	return result, nil
}

// 获取待审核评论
func GetPendingComments() ([]Comment, error) {
	comments, err := LoadComments()
	if err != nil {
		return nil, err
	}

	var result []Comment
	for _, c := range comments {
		if !c.Approved {
			result = append(result, c)
		}
	}

	// 按时间倒序
	sort.Slice(result, func(i, j int) bool {
		return result[i].CreatedAt.After(result[j].CreatedAt)
	})

	return result, nil
}

// 获取所有评论
func GetAllComments() ([]Comment, error) {
	comments, err := LoadComments()
	if err != nil {
		return nil, err
	}

	// 按时间倒序
	sort.Slice(comments, func(i, j int) bool {
		return comments[i].CreatedAt.After(comments[j].CreatedAt)
	})

	return comments, nil
}

// 审核评论
func ApproveComment(id string) error {
	comments, err := LoadComments()
	if err != nil {
		return err
	}

	for i := range comments {
		if comments[i].ID == id {
			comments[i].Approved = true
			break
		}
	}

	return SaveComments(comments)
}

// 删除评论
func DeleteComment(id string) error {
	comments, err := LoadComments()
	if err != nil {
		return err
	}

	var result []Comment
	for _, c := range comments {
		if c.ID != id {
			result = append(result, c)
		}
	}

	return SaveComments(result)
}
