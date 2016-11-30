package ledger


// Store provides storage for State data
type StateStore interface {
	//TODO: define the state store func
	SaveState(*State) error
}


type State struct {
	//TODO: define the state struct
}
