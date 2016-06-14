package to

import (
	"github.com/tangyang/simple-http-server/model"
)

type RelationTo struct {
	UserId int64
	State  string
	Type   string
}

const (
	relationType = "relationship"
)

func NewRelationTo(relation *model.Relation) *RelationTo {
	relationDescription := string(relation.Status.ToRelationStatusDescription())
	return &RelationTo{UserId: relation.Otheruserid, State: relationDescription, Type: relationType}
}

func NewRelationToArray(relations []model.Relation) []RelationTo {
	if relations != nil {
		size := len(relations)
		var result = []RelationTo{}
		for i := 0; i < size; i++ {
			relationDescription := string(relations[i].Status.ToRelationStatusDescription())
			result = append(result, RelationTo{UserId: relations[i].Otheruserid, State: relationDescription, Type: relationType})
		}
		return result
	} else {
		return nil
	}
}
