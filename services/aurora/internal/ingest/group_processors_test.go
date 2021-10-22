//lint:file-ignore U1001 Ignore all unused code, staticcheck doesn't understand testify/suite
package ingest

import (
	"errors"
	"testing"

	"github.com/diamnet/go/ingest/io"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

var _ auroraChangeProcessor = (*mockAuroraChangeProcessor)(nil)

type mockAuroraChangeProcessor struct {
	mock.Mock
}

func (m *mockAuroraChangeProcessor) ProcessChange(change io.Change) error {
	args := m.Called(change)
	return args.Error(0)
}

func (m *mockAuroraChangeProcessor) Commit() error {
	args := m.Called()
	return args.Error(0)
}

var _ auroraTransactionProcessor = (*mockAuroraTransactionProcessor)(nil)

type mockAuroraTransactionProcessor struct {
	mock.Mock
}

func (m *mockAuroraTransactionProcessor) ProcessTransaction(transaction io.LedgerTransaction) error {
	args := m.Called(transaction)
	return args.Error(0)
}

func (m *mockAuroraTransactionProcessor) Commit() error {
	args := m.Called()
	return args.Error(0)
}

type GroupChangeProcessorsTestSuiteLedger struct {
	suite.Suite
	processors *groupChangeProcessors
	processorA *mockAuroraChangeProcessor
	processorB *mockAuroraChangeProcessor
}

func TestGroupChangeProcessorsTestSuiteLedger(t *testing.T) {
	suite.Run(t, new(GroupChangeProcessorsTestSuiteLedger))
}

func (s *GroupChangeProcessorsTestSuiteLedger) SetupTest() {
	s.processorA = &mockAuroraChangeProcessor{}
	s.processorB = &mockAuroraChangeProcessor{}
	s.processors = &groupChangeProcessors{
		s.processorA,
		s.processorB,
	}
}

func (s *GroupChangeProcessorsTestSuiteLedger) TearDownTest() {
	s.processorA.AssertExpectations(s.T())
	s.processorB.AssertExpectations(s.T())
}

func (s *GroupChangeProcessorsTestSuiteLedger) TestProcessChangeFails() {
	change := io.Change{}
	s.processorA.
		On("ProcessChange", change).
		Return(errors.New("transient error")).Once()

	err := s.processors.ProcessChange(change)
	s.Assert().Error(err)
	s.Assert().EqualError(err, "error in *ingest.mockAuroraChangeProcessor.ProcessChange: transient error")
}

func (s *GroupChangeProcessorsTestSuiteLedger) TestProcessChangeSucceeds() {
	change := io.Change{}
	s.processorA.
		On("ProcessChange", change).
		Return(nil).Once()
	s.processorB.
		On("ProcessChange", change).
		Return(nil).Once()

	err := s.processors.ProcessChange(change)
	s.Assert().NoError(err)
}

func (s *GroupChangeProcessorsTestSuiteLedger) TestCommitFails() {
	s.processorA.
		On("Commit").
		Return(errors.New("transient error")).Once()

	err := s.processors.Commit()
	s.Assert().Error(err)
	s.Assert().EqualError(err, "error in *ingest.mockAuroraChangeProcessor.Commit: transient error")
}

func (s *GroupChangeProcessorsTestSuiteLedger) TestCommitSucceeds() {
	s.processorA.
		On("Commit").
		Return(nil).Once()
	s.processorB.
		On("Commit").
		Return(nil).Once()

	err := s.processors.Commit()
	s.Assert().NoError(err)
}

type GroupTransactionProcessorsTestSuiteLedger struct {
	suite.Suite
	processors *groupTransactionProcessors
	processorA *mockAuroraTransactionProcessor
	processorB *mockAuroraTransactionProcessor
}

func TestGroupTransactionProcessorsTestSuiteLedger(t *testing.T) {
	suite.Run(t, new(GroupTransactionProcessorsTestSuiteLedger))
}

func (s *GroupTransactionProcessorsTestSuiteLedger) SetupTest() {
	s.processorA = &mockAuroraTransactionProcessor{}
	s.processorB = &mockAuroraTransactionProcessor{}
	s.processors = &groupTransactionProcessors{
		s.processorA,
		s.processorB,
	}
}

func (s *GroupTransactionProcessorsTestSuiteLedger) TearDownTest() {
	s.processorA.AssertExpectations(s.T())
	s.processorB.AssertExpectations(s.T())
}

func (s *GroupTransactionProcessorsTestSuiteLedger) TestProcessTransactionFails() {
	transaction := io.LedgerTransaction{}
	s.processorA.
		On("ProcessTransaction", transaction).
		Return(errors.New("transient error")).Once()

	err := s.processors.ProcessTransaction(transaction)
	s.Assert().Error(err)
	s.Assert().EqualError(err, "error in *ingest.mockAuroraTransactionProcessor.ProcessTransaction: transient error")
}

func (s *GroupTransactionProcessorsTestSuiteLedger) TestProcessTransactionSucceeds() {
	transaction := io.LedgerTransaction{}
	s.processorA.
		On("ProcessTransaction", transaction).
		Return(nil).Once()
	s.processorB.
		On("ProcessTransaction", transaction).
		Return(nil).Once()

	err := s.processors.ProcessTransaction(transaction)
	s.Assert().NoError(err)
}

func (s *GroupTransactionProcessorsTestSuiteLedger) TestCommitFails() {
	s.processorA.
		On("Commit").
		Return(errors.New("transient error")).Once()

	err := s.processors.Commit()
	s.Assert().Error(err)
	s.Assert().EqualError(err, "error in *ingest.mockAuroraTransactionProcessor.Commit: transient error")
}

func (s *GroupTransactionProcessorsTestSuiteLedger) TestCommitSucceeds() {
	s.processorA.
		On("Commit").
		Return(nil).Once()
	s.processorB.
		On("Commit").
		Return(nil).Once()

	err := s.processors.Commit()
	s.Assert().NoError(err)
}
