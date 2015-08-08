package things

import (
	"github.com/likestripes/pacific"
	"strings"
	"time"
)

type Tag struct {
	Scope   Scope  `datastore:"-" sql:"-" json:"-"`
	Claimed bool   `json:"-"`
	Policy  string `json:"-"`
	Owner   string `json:"-"`

	TagId    string
	ObjectId string
	OriginId string
	ParentId string
	Status   int
	Created  time.Time
	Updated  time.Time
	PersonId int64
	Things   []string `json:"-"`
}

func (scope Scope) Tags() (tags []Tag) {

	query := pacific.Query{
		Context:   scope.Context,
		Kind:      "SharedTag",
		Ancestors: scope.Ancestors,
	}

	err := query.GetAll(&tags)
	scope.Context.Infof("TAGS: " + scope.PersonIdStr)
	if err != nil {
		scope.Context.Errorf(err.Error())
	}

	results := []Tag{}

	for _, tag := range tags {
		if tag.ObjectId != scope.PersonIdStr {
			for _, ancestor := range scope.Ancestors {
				if ancestor.Kind == "Origin" && ancestor.KeyString != tag.ObjectId {
					tag.ObjectId = strings.Replace(tag.ObjectId, scope.PersonIdStr+"/"+ancestor.KeyString+"/", "", -1)
					results = append(results, tag)
				}
			}
		}
	}

	return uniq(results)
}

func (scope Scope) Tag(tag_id string) Tag {
	return Tag{
		Scope:    scope,
		TagId:    tag_id,
		ObjectId: tag_id,
		PersonId: scope.PersonId,
	}
}

func uniq(tags []Tag) (results []Tag) {

	for _, tag := range tags {
		unique := true
		for _, result := range results {
			if result.ObjectId == tag.ObjectId {
				unique = false
			}
		}
		if unique {
			results = append(results, tag)
		}
	}
	return results
}

//TODO: Tags...
func (tag Tag) Save() {

}

func (tag Tag) Share(person_id int64) {

}
