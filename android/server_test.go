package candy

import (
	"testing"
)


func TestCandy(t *testing.T) {
    c := NewCandy("127.0.0.1:6543")
    if err := c.Start(); err != nil {
        println(err.Error())
        return
    }

    id, err := c.Register("abc", "123")
    if err != nil {
        println(err.Error())
        return
    }
    println(id)
}
