package artifact

type Artifact interface {
	SetKey(string)
	SetName(string)
	SetParent(string)
	GetCollection() string
}
