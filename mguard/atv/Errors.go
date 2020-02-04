package atv

import "fmt"

// ErrNilReceiver is an error that is issued if a function is called with a nil receiver.
var ErrNilReceiver = fmt.Errorf("The receiver is nil")
