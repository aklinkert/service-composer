package servicecomposer

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"github.com/samber/do/v2"
	"github.com/stretchr/testify/suite"
)

type ComposerTestSuite struct {
	suite.Suite

	composer   *composer
	ctx        context.Context
	ctxCancel  context.CancelFunc
	testServer *testServer

	foundationModule *testModule[string]
	testModule1      *testModule[string]
	testModule2      *testModule[string]
}

func (s *ComposerTestSuite) SetupTest() {
	s.ctx, s.ctxCancel = context.WithCancel(context.Background())
	s.composer = New(s.ctx, slog.New(slog.NewTextHandler(io.Discard, nil)), "test").(*composer)

	s.testServer = &testServer{}

	s.foundationModule = newTestModule("foundation", nil, "foundation")
	s.testModule1 = newTestModule("testModule1", nil, "val1")
	s.testModule2 = newTestModule("testModule2", []string{"testModule1"}, "val2")
}

func TestEcosystemSuite(t *testing.T) {
	suite.Run(t, new(ComposerTestSuite))
}

func (s *ComposerTestSuite) cancelAfter10ms() {
	go func() {
		time.Sleep(10 * time.Millisecond)
		s.ctxCancel()
	}()
}

func (s *ComposerTestSuite) TestInjector() {
	s.Assert().NotNil(s.composer.Injector())
	s.Assert().IsType(&do.RootScope{}, s.composer.Injector())
}

func (s *ComposerTestSuite) TestStart_NoModules() {
	s.composer.MustAddModules(s.foundationModule)
	s.cancelAfter10ms()

	startErr := s.composer.Start()
	s.Require().NoErrorf(startErr, "eco.Start() should not return an error, but returned %v", startErr)

	s.Assert().True(s.foundationModule.initialized, "foundationModule should be initialized")
	s.Assert().True(s.foundationModule.started, "foundationModule should be started")
}

func (s *ComposerTestSuite) TestStart_TwoModules() {
	s.composer.MustAddModules(s.foundationModule, s.testModule1, s.testModule2)
	s.cancelAfter10ms()

	startErr := s.composer.Start()
	s.Require().NoErrorf(startErr, "eco.Start() should not return an error, but returned %v", startErr)

	s.Assert().True(s.foundationModule.initialized, "foundationModule should be initialized")
	s.Assert().True(s.foundationModule.started, "foundationModule should be started")

	s.Assert().True(s.testModule1.initialized, "testModule1 should be initialized")
	s.Assert().True(s.testModule1.started, "testModule1 should be started")

	s.Assert().True(s.testModule2.initialized, "testModule2 should be initialized")
	s.Assert().True(s.testModule2.started, "testModule2 should be started")
}

func (s *ComposerTestSuite) TestStart_TwoModules_MissingDependency() {
	missingDependencyModule := newTestModule("testModule1", []string{"missing"}, "val1")

	s.composer.MustAddModules(s.foundationModule, missingDependencyModule, s.testModule2)

	s.Assert().ErrorIs(s.composer.Start(), ErrModuleNotFound, "eco.Start() should return an error when a module is missing a dependency")
	s.Assert().EqualError(s.composer.Start(), "module not found: module testModule1 depends on missing, but it is not registered", "eco.Start() should return an error when a module is missing a dependency")
}
