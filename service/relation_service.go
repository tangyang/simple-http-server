package service

import (
	"github.com/tangyang/simple-http-server/config"
	"github.com/tangyang/simple-http-server/dao"
	"github.com/tangyang/simple-http-server/model"

	"fmt"
)

var relationDao *dao.RelationDao = &dao.RelationDao{}

type RelationService struct {
}

func (*RelationService) AddRelation(conf *config.Config, relation *model.Relation) (bool, error) {
	existRelation := relationDao.GetRelationByUserIdPairs(conf, relation.Otheruserid, relation.Userid)
	if existRelation != nil && existRelation.Status == model.RelationLike && relation.Status == model.RelationLike {
		fmt.Sprintf("User %d is already liked by %d \n", relation.Userid, relation.Otheruserid)
		relation.Status = model.RelationMatched
		b, err := relationDao.AddOrUpdateRelation(conf, relation)
		if b {
			existRelation.Status = model.RelationMatched
			err = relationDao.UpdateRelation(conf, existRelation)
		}
		return b, err
	} else {
		return relationDao.AddOrUpdateRelation(conf, relation)
	}
}

func (r *RelationService) GetRelations(conf *config.Config, userId int64) []model.Relation {
	return relationDao.GetAllRelationsByUserId(conf, userId)
}
