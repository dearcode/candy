package meta

import (
	"fmt"
)

// Error return error.
func (r *ResponseHeader) Error() string {
	return fmt.Sprintf("%d:%s", r.Code, r.Msg)
}
