package dao

import (
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/model"

	"fmt"
)

type RelationDao struct {
}

func (r *RelationDao) CreateRelationSchema(conf *config.Config) error {
	c := NewPostgreConnector(conf)
	_, err := c.DB.Exec("CREATE TABLE relations (id bigserial PRIMARY key , userid bigint, otheruserid bigint, status smallint)")
	return err
}

func (r *RelationDao) AddOrUpdateRelation(conf *config.Config, relation *model.Relation) (bool, error) {
	c := NewPostgreConnector(conf)
	b, err := c.DB.Model(relation).Where("userid=? and otheruserid=?", relation.Userid, relation.Otheruserid).SelectOrCreate()
	return b, err
}

func (r *RelationDao) GetRelationByUserIdPairs(conf *config.Config, userId int64, otherUserId int64) *model.Relation {
	c := NewPostgreConnector(conf)
	relation := &model.Relation{}
	err := c.DB.Model(relation).Where("userid=? and otheruserid=?", userId, otherUserId).Select()
	if err != nil {
		fmt.Sprintf("Fail to get relation by user id %d and other user id %d, error: %s\n", userId, otherUserId, err.Error())
		return nil
	}
	return relation
}

func (r *RelationDao) UpdateRelation(conf *config.Config, relation *model.Relation) error {
	c := NewPostgreConnector(conf)
	_, err := c.DB.Model(relation).Set("status=?", relation.Status).Where("id=?", relation.Id).Update()
	return err
}

func (r *RelationDao) GetAllRelationsByUserId(conf *config.Config, userId int64) []model.Relation {
	c := NewPostgreConnector(conf)
	var relations []model.Relation
	_, err := c.DB.Query(&relations, `SELECT * FROM relations where userid = ? `, userId)
	if err != nil {
		fmt.Sprintf("Fail to get all relations by userId, error: %s\n", err.Error())
		return nil
	}
	return relations
}
