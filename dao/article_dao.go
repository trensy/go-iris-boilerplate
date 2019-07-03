package dao

import (
	"gitee.com/trensy/duocaiCRM/models"
	"github.com/go-xorm/xorm"
	"log"
	"time"
)

type ArticleDao struct {
	engine *xorm.Engine
}

func NewArticleDao(engine *xorm.Engine) *ArticleDao{
	return &ArticleDao{
		engine:engine}
}

func (dao *ArticleDao) Get(id int) *models.Article {
	data := &models.Article{Id: id}
	ok,err :=dao.engine.Get(data)
	if ok && err == nil {
		return data
	}else{
		data.Id=0
		return data
	}
}

func (dao *ArticleDao) Insert(title string) *models.Article {
	article := &models.Article{
		Title:title,
		Content:title,
		Desc:title,
		IsDelete:0,
		CreatedAt:int(time.Now().Unix()),
		UpdatedAt:int(time.Now().Unix()),
	}
	_, err := dao.engine.Insert(article)
	if err != nil {
		log.Fatal("insert error: ", err)
	}
	return article
}

