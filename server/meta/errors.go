package meta

import (
	"fmt"
)

// Error return error.
func (r *ResponseHeader) Error() error {
	return fmt.Errorf("%d:%s", r.Code, r.Msg)
}
