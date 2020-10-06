package model

import (
	"github.com/go-park-mail-ru/2020_2_Eternity/configs/config"
	"log"
)

type Pin struct {
	Id      int    `json:"id"`
	Title   string `json:"title"`
	Content string `json:"content"`
	ImgLink string `json:"img_link"`
	UserId  int    `json:"user_id"`
}

func (p *Pin) CreatePin() error {
	err := config.Db.QueryRow("insert into pins(title, content, img_link, user_id) values($1, $2, $3, $4) returning id",
		p.Title, p.Content, p.ImgLink, p.UserId).Scan(&p.Id)

	if err != nil {
		return err
	}

	return nil
}

func (p *Pin) GetPin() bool {
	row := config.Db.QueryRow("select id, title, content, img_link, user_id from pins where id=$1", p.Id)

	if err := row.Scan(&p.Id, &p.Title, &p.Content, &p.ImgLink, &p.UserId); err != nil {
		log.Print(err)
		return false
	}

	return true
}