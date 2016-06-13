package model

type Relation struct {
	UserId      int64
	OtherUserId int64
	State       RelationStatus
}

type RelationStatus int

const (
	RelationLike RelationStatus = iota
	RelationDislike
	RelationMatched
)
