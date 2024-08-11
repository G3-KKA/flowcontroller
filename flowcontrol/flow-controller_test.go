package flowcontrol

import (
	"context"
	"testing"
	"time"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/suite"
)

type FlowControllerTestSuite struct {
	suite.Suite
}

func TestController(t *testing.T) {
	suite.Run(t, new(FlowControllerTestSuite))
}

// ========================

func (t *FlowControllerTestSuite) SetupSuite() {
	viper.Set("WORKSPACE", "..")
	viper.Set("CONFIG_FILE", viper.GetString("WORKSPACE")+"/config.yaml")
}

func (t *FlowControllerTestSuite) BeforeTest(suiteName, testName string) {
	switch testName {
	}
}

// ========================

func (t *FlowControllerTestSuite) Test_stopAndDrain() {
	timer := time.NewTimer(time.Nanosecond)
	time.Sleep(10 * time.Millisecond)
	stopAndDrain(timer)
}

// Разделить на несколько тестов, очень много разных сценариев
func (t *FlowControllerTestSuite) Test_FlowController() {
	flow, err := New(context.TODO())
	t.NoError(err)
	mock_managebale := NewMockManageable(t.T())
	flag := false
	gracefullshutdowncalled := func() bool {
		return flag

	}
	go t.Eventually(gracefullshutdowncalled, 1*time.Second, 100*time.Millisecond)

	mock_managebale.On("ServiceManager").Return(ServiceManager(func(tunnel ControlTunnel) error {
		for {
			msg, replyF := tunnel.ReadMessage()
			if msg == GracefulShutdown {
				replyF(OK)
				flag = true
			}
		}
	}))
	sid, err := flow.Register(mock_managebale)
	t.NoError(err)
	md, err := flow.Metadata(sid)
	t.NoError(err)
	t.NotNil(md)
	t.NotNil(md.logger)
	t.NotZero(md)
	ctrl, ok := flow.(*controller)
	t.True(ok)
	t.NotNil(ctrl)
	sslice, runlock := ctrl.broadcast.GetRead()
	t.NotNil(sslice)
	t.Len(sslice, 2) // one channel for controller itself and one for mocked service
	for _, msgchan := range sslice {
		msgchan <- internalMsg{
			data: GracefulShutdown,
		}
	}
	runlock()
	time.Sleep(2 * time.Second)
}
