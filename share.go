package things

import (
	"github.com/likestripes/pacific"
	"time"
)

type Share struct {
	Kind     string `datastore:"-" sql:"-"`
	ObjectId string
	ParentId string `pacific_parent:"parent_id"`
	PersonId int64  `pacific_parent:"person_id"`
	OriginId string `pacific_parent:"origin_id"`
	Status   int
	Created  time.Time `sql:"-"`
	Updated  time.Time `sql:"-"`
	Scope    Scope     `datastore:"-" sql:"-"`
}

func (scope Scope) Shares(tags []string) (shares []Share) {
	shares_by_tag := map[string][]Share{}
	for _, tag := range tags {
		shared_things := scope.SharedThings(tag)
		if len(shared_things) == 0 {
			return
		}
		for _, share := range shared_things {
			if shareInArray(shares_by_tag[tag], share.ObjectId) == false {
				shares_by_tag[tag] = append(shares_by_tag[tag], share)
			}
		}
	}
	for _, share := range shares_by_tag[tags[0]] {
		count := 1
		for _, tag := range tags[1:] {
			if shareInArray(shares_by_tag[tag], share.ObjectId) {
				count = count + 1
			}
		}
		if count == len(tags) && shareInArray(shares, share.ObjectId) == false {
			shares = append(shares, share)
		}
	}

	return shares
}

func shareInArray(haystack []Share, needle string) bool {
	for _, item := range haystack {
		if item.ObjectId != "" && item.ObjectId == needle {
			return true
		}
	}
	return false
}

func (scope Scope) SharedThings(tag string) (shared_things []Share) {
	ancestor := pacific.Ancestor{
		Context:    scope.Context,
		Kind:       "SharedTag",
		KeyString:  tag,
		PrimaryKey: "parent_id",
	}
	scope.Ancestors = append(scope.Ancestors, ancestor)

	query := pacific.Query{
		Context:    scope.Context,
		Ancestors:  scope.Ancestors,
		Kind:       "SharedThing",
		PrimaryKey: "object_id",
	}
	query.GetAll(&shared_things)
	return
}

func (share Share) Save() {
	scope := share.Scope
	share.PersonId = scope.PersonId
	query := pacific.Query{
		Context:    scope.Context,
		Ancestors:  scope.Ancestors,
		Kind:       share.Kind,
		KeyString:  share.ObjectId,
		PrimaryKey: "object_id",
	}
	query.Put(&share)
	return
}
