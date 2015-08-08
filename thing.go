package things

import (
	"encoding/json"
	"github.com/likestripes/pacific"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type Thing struct {
	Scope Scope `datastore:"-" sql:"-" json:"-"`

	Value string `sql:"type:text" datastore:",noindex" json:"-"`

	Map        map[string]interface{} `datastore:"-" sql:"-"`
	Tags       map[string]Tag         `datastore:"-" sql:"-"`
	Dispatched map[string]interface{} `datastore:"-" sql:"-" json:"-"`
	Mock       bool                   `datastore:"-" sql:"-" json:"-"`
	Sharing    int                    `datastore:"-" sql:"-" json:"-"`

	ThingId     string
	Created     time.Time `sql:"-"`
	Updated     time.Time `sql:"-"`
	PersonId    int64
	PersonIdStr string
	Status      int  `json:"-"`
	IsNew       bool `datastore:"-" sql:"-" json:"-"`
}

func (scope Scope) Things(tag_ar ...[]string) (things []Thing) {

	for _, tags := range tag_ar {
		shares := scope.Shares(tags)
		for _, share := range shares {
			if notIn(things, share.ObjectId) {
				things = append(things, scope.Thing(share.ObjectId))
			}
		}
	}

	return things
}

func notIn(ar []Thing, thing_id string) bool {
	for _, thing := range ar {
		if thing.ThingId == thing_id {
			return false
		}
	}
	return true
}

func (scope Scope) Thing(thing_id string) (thing Thing) {
	if thing_id != "" {
		query := pacific.Query{
			Context:   scope.Context,
			Kind:      "Thing",
			KeyString: thing_id,
		}
		err := query.Get(&thing)
		scope.Context.Infof(thing_id)
		if err != nil {
			scope.Context.Errorf(err.Error())
			thing = thing.New(thing_id)
		} else {
			thing.Map = make(map[string]interface{})
			properties := make(map[string]interface{})
			json.Unmarshal([]byte(thing.Value), &properties)

			for key, value := range properties {
				thing.Map[key] = value
			}
		}

	} else {
		thing_id = scope.PersonIdStr + "/" + newColor() //TODO: default id scheme
		thing = thing.New(thing_id)
	}

	if thing.Tags == nil {
		thing.Tags = make(map[string]Tag)
	}

	thing.Scope = scope
	return thing
}

func (thing Thing) New(thing_id string) Thing {

	thing = Thing{
		ThingId: thing_id,
		Map:     make(map[string]interface{}),
		IsNew:   true,
		Created: time.Now(),
		Updated: time.Now(),
	}
	thing.Map = make(map[string]interface{})
	return thing
}

func (thing Thing) Share(person_id int64, person_id_str, tag string) {

	scope := thing.Scope

	share_tag := Share{
		Kind:     "SharedTag",
		Scope:    scope,
		ParentId: person_id_str,
		PersonId: person_id,
		ObjectId: tag,
	}

	share_tag.Save()

	ancestor := pacific.Ancestor{
		Context:    scope.Context,
		Kind:       "SharedTag",
		KeyString:  tag,
		PrimaryKey: "parent_id",
	}
	scope.Ancestors = append(scope.Ancestors, ancestor)

	share_thing := Share{
		Kind:     "SharedThing",
		Scope:    scope,
		ParentId: tag,
		ObjectId: thing.ThingId,
	}
	share_thing.Save()
}

func (thing *Thing) Save() bool {

	if thing.Map == nil {
		thing.Map = make(map[string]interface{})
	}
	if thing.Tags == nil {
		thing.Tags = make(map[string]Tag)
	}

	scope := thing.Scope
	existing := scope.Thing(thing.ThingId)
	if existing.Status > 0 {
		for key, item := range existing.Map {
			if thing.Map[key] == nil {
				thing.Map[key] = item
			}
		}
	}

	value_json, _ := json.Marshal(thing.Map)
	thing.Value = string(value_json)

	thing.Updated = time.Now()

	query := pacific.Query{
		Context:   scope.Context,
		Kind:      "Thing",
		KeyString: thing.ThingId,
	}
	query.Put(thing)

	if _, ok := thing.Tags[thing.ThingId]; !ok {
		thing.Tags[thing.ThingId] = Tag{
			TagId: thing.ThingId,
		}
	}

	for tag, _ := range thing.Tags {
		thing.Share(thing.PersonId, thing.PersonIdStr, tag)
	}

	return true
}

func (thing *Thing) TagsFromString(sep string, chunks ...interface{}) {
	var tags []string

	for _, chunk_int := range chunks {
		if chunk_int != nil {
			chunk := chunk_int.(string)
			if sep != "" {
				tags = strings.Split(chunk, sep)
			} else {
				tags = []string{chunk}
			}

			for _, tag_name := range tags {
				if _, ok := thing.Tags[tag_name]; !ok && tag_name != "" {
					scope := thing.Scope
					tag := scope.Tag(tag_name)
					thing.Tags[tag.TagId] = tag
				}
			}
		}
	}
	return
}

func newColor() string {
	red := strconv.FormatInt(int64(rand.Intn(255)), 16)
	blue := strconv.FormatInt(int64(rand.Intn(255)), 16)
	green := strconv.FormatInt(int64(rand.Intn(255)), 16)
	color := string(red + blue + green)
	if len(color) < 6 {
		color = color + strings.Repeat("0", 6-len(color))
	}
	return color
}
