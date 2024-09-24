package storage

import (
	"GoJob/db"
	"GoJob/info"
	"GoJob/xlog"
)

type JumpitPost interface {
	StorePost(post info.JumpitPost) error
	GetPost(id string) (post info.JumpitPost, err error)
}

type SqliteJumpitPost struct {
	*db.Sqlite
}

func (s *SqliteJumpitPost) StorePost(post info.JumpitPost) error {
	data := make(map[string]interface{})
	data["name"] = post.Name
	data["description"] = post.Description
	err := s.InsertData("jumpit", data)
	if err != nil {
		xlog.Logger.Error(err)
		return err
	}
	return nil
}

func (s *SqliteJumpitPost) GetPost(id string) (post info.JumpitPost, err error) {
	return
}
