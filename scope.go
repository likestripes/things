package things

import (
	"github.com/likestripes/pacific"
)

type Scope struct {
	Context     pacific.Context
	Ancestors   []pacific.Ancestor
	PersonId    int64
	PersonIdStr string
}

func ScopeToPerson(person_id int64) (ancestors []pacific.Ancestor) {
	if person_id == 0 {
		return
	}
	ancestor := pacific.Ancestor{
		Kind:   "Person",
		KeyInt: person_id,
	}
	return []pacific.Ancestor{ancestor}
}
