package slackbot

type Store struct {
	sotd string
}

func NewStore() *Store {
	return &Store{
		sotd: "",
	}
}
