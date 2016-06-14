package model

type Relation struct {
	Id          int64
	Userid      int64
	Otheruserid int64
	Status      RelationStatus
}

type RelationStatus int

const (
	RelationLike RelationStatus = iota
	RelationDislike
	RelationMatched
)

type RelationStatusDescription string

const (
	RelationLikeDescription    RelationStatusDescription = "liked"
	RelationDislikeDescription RelationStatusDescription = "disliked"
	RelationMatchedDescription RelationStatusDescription = "matched"
)

func (r RelationStatus) ToRelationStatusDescription() RelationStatusDescription {
	if r == RelationLike {
		return RelationLikeDescription
	} else if r == RelationDislike {
		return RelationDislikeDescription
	} else {
		return RelationMatchedDescription
	}
}
