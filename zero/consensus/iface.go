package consensus

import (
	"github.com/dece-cash/go-dece/decedb"
)

type DB interface {
	CurrentTri() decedb.Tri
	GlobalGetter() decedb.Getter
}

type CItem interface {
	CopyTo() (ret CItem)
	CopyFrom(CItem)
}

type PItem interface {
	CItem
	Id() (ret []byte)
	State() (ret []byte)
}
