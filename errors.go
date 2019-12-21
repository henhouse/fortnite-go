package fortnitego

type Error struct{ e string }

func (e *Error) Error() string {
	return e.e
}

// ErrNotFound is returned when we receive a 404 when attempting to query a player.
var ErrNotFound = Error{"Character not found."}
