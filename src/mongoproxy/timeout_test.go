package mongoproxy

import (
	. "launchpad.net/gocheck"
	"time"
)

type TimeoutSuite struct{}

var _ = Suite(&TimeoutSuite{})

func (s *TimeoutSuite) TestTimeoutFires(c *C) {
	block := func() (interface{}, error) {
		time.Sleep(10 * time.Second)
		return nil, nil
	}

	_, err := TimeoutIn(block, 10*time.Millisecond)
	c.Check(IsTimeout(err), Equals, true)
}

func (s *TimeoutSuite) TestTimeoutDoesNotFireInSimpleCase(c *C) {
	block := func() (interface{}, error) {
		return 5, nil
	}

	value, err := TimeoutIn(block, 30*time.Millisecond)
	c.Check(value, Equals, 5)
	c.Check(err, IsNil)
}
