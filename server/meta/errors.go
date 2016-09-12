package meta

import (
	"fmt"
)

// Error return error.
func (r *ResponseHeader) Error() error {
	if r == nil {
		return nil
	}
	return fmt.Errorf("%d:%s", r.Code, r.Msg)
}
