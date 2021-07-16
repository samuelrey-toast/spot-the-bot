package cmd

import (
	"testing"

	"github.com/samuelrey/spot-the-bot/framework"
	"github.com/stretchr/testify/suite"
)

type UserQueueSuite struct{ suite.Suite }

func (suite UserQueueSuite) TestHeadPop() {
	q := SimpleUserQueue{}

	// empty queue
	actual := q.Head()
	suite.Require().Nil(actual)

	actual = q.Pop()
	suite.Require().Nil(actual)

	// nonempty queue
	au := framework.MessageUser{ID: "amethyst#4422", Username: "amethyst"}
	ou := framework.MessageUser{ID: "osh#1219", Username: "osh"}
	q = SimpleUserQueue{queue: []framework.MessageUser{au, ou}}

	actual = q.Head()
	suite.Require().Equal(&au, actual)

	actual = q.Pop()
	suite.Require().Equal(&au, actual)
	suite.Require().Equal(1, q.Length())

	actual = q.Head()
	suite.Require().Equal(&ou, actual)
}

func (suite UserQueueSuite) TestPushRemove() {
	q := SimpleUserQueue{}
	mu := framework.MessageUser{ID: "amethyst#4422", Username: "amethyst"}

	q.Push(mu)
	actual := q.Head()
	suite.Require().Equal(&mu, actual)

	q.Remove(mu)
	actual = q.Head()
	suite.Require().Nil(actual)
}

func TestUserQueue(t *testing.T) {
	suite.Run(t, new(UserQueueSuite))
}
