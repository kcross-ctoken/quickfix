package quickfix

import (
	"testing"
	"time"

	"github.com/quickfixgo/quickfix/config"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/internal"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

func newFIXString(val string) *FIXString {
	s := FIXString(val)
	return &s
}

type NewSessionTestSuite struct {
	suite.Suite

	SessionID
	MessageStoreFactory
	*SessionSettings
	LogFactory
	App *MockApp
}

func TestNewSessionTestSuite(t *testing.T) {
	suite.Run(t, new(NewSessionTestSuite))
}

func (s *NewSessionTestSuite) SetupTest() {
	s.SessionID = SessionID{BeginString: "FIX.4.2", TargetCompID: "TW", SenderCompID: "ISLD"}
	s.MessageStoreFactory = NewMemoryStoreFactory()
	s.SessionSettings = NewSessionSettings()
	s.LogFactory = nullLogFactory{}
	s.App = new(MockApp)
}

func (s *NewSessionTestSuite) TestDefaults() {
	session, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.Nil(err)
	s.NotNil(session)

	s.False(session.resetOnLogon)
	s.False(session.resetOnLogout)
	s.Nil(session.sessionTime, "By default, start and end time unset")
}

func (s *NewSessionTestSuite) TestResetOnLogon() {
	var tests = []struct {
		setting  string
		expected bool
	}{{"Y", true}, {"N", false}}

	for _, test := range tests {
		s.SetupTest()
		s.SessionSettings.Set(config.ResetOnLogon, test.setting)
		session, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
		s.Nil(err)
		s.NotNil(session)

		s.Equal(test.expected, session.resetOnLogon)
	}
}

func (s *NewSessionTestSuite) TestResetOnLogout() {
	var tests = []struct {
		setting  string
		expected bool
	}{{"Y", true}, {"N", false}}

	for _, test := range tests {
		s.SetupTest()
		s.SessionSettings.Set(config.ResetOnLogout, test.setting)
		session, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
		s.Nil(err)
		s.NotNil(session)

		s.Equal(test.expected, session.resetOnLogout)
	}
}

func (s *NewSessionTestSuite) TestStartAndEndTime() {
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	session, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.Nil(err)
	s.NotNil(session.sessionTime)

	s.Equal(
		*internal.NewUTCTimeRange(internal.NewTimeOfDay(12, 0, 0), internal.NewTimeOfDay(14, 0, 0)),
		*session.sessionTime,
	)
}

func (s *NewSessionTestSuite) TestStartAndEndTimeAndTimeZone() {
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	s.SessionSettings.Set(config.TimeZone, "Local")

	session, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.Nil(err)
	s.NotNil(session.sessionTime)

	s.Equal(
		*internal.NewTimeRangeInLocation(internal.NewTimeOfDay(12, 0, 0), internal.NewTimeOfDay(14, 0, 0), time.Local),
		*session.sessionTime,
	)
}

func (s *NewSessionTestSuite) TestStartAndEndTimeAndStartAndEndDay() {
	var tests = []struct {
		startDay, endDay string
	}{
		{"Sunday", "Thursday"},
		{"Sun", "Thu"},
	}

	for _, test := range tests {
		s.SetupTest()

		s.SessionSettings.Set(config.StartTime, "12:00:00")
		s.SessionSettings.Set(config.EndTime, "14:00:00")
		s.SessionSettings.Set(config.StartDay, test.startDay)
		s.SessionSettings.Set(config.EndDay, test.endDay)

		session, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
		s.Nil(err)
		s.NotNil(session.sessionTime)

		s.Equal(
			*internal.NewUTCWeekRange(
				internal.NewTimeOfDay(12, 0, 0), internal.NewTimeOfDay(14, 0, 0),
				time.Sunday, time.Thursday,
			),
			*session.sessionTime,
		)
	}
}

func (s *NewSessionTestSuite) TestStartAndEndTimeAndStartAndEndDayAndTimeZone() {
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	s.SessionSettings.Set(config.StartDay, "Sunday")
	s.SessionSettings.Set(config.EndDay, "Thursday")
	s.SessionSettings.Set(config.TimeZone, "Local")

	session, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.Nil(err)
	s.NotNil(session.sessionTime)

	s.Equal(
		*internal.NewWeekRangeInLocation(
			internal.NewTimeOfDay(12, 0, 0), internal.NewTimeOfDay(14, 0, 0),
			time.Sunday, time.Thursday, time.Local,
		),
		*session.sessionTime,
	)
}

func (s *NewSessionTestSuite) TestMissingStartOrEndTime() {
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	_, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)

	s.SetupTest()
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	_, err = newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)
}

func (s *NewSessionTestSuite) TestStartOrEndTimeParseError() {
	s.SessionSettings.Set(config.StartTime, "1200:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")

	_, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)

	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "")

	_, err = newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)
}

func (s *NewSessionTestSuite) TestInvalidTimeZone() {
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	s.SessionSettings.Set(config.TimeZone, "not valid")

	_, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)
}

func (s *NewSessionTestSuite) TestMissingStartOrEndDay() {
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	s.SessionSettings.Set(config.StartDay, "Thursday")
	_, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)

	s.SetupTest()
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	s.SessionSettings.Set(config.EndDay, "Sunday")
	_, err = newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)
}

func (s *NewSessionTestSuite) TestStartOrEndDayParseError() {
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	s.SessionSettings.Set(config.StartDay, "notvalid")
	s.SessionSettings.Set(config.EndDay, "Sunday")
	_, err := newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)

	s.SetupTest()
	s.SessionSettings.Set(config.StartTime, "12:00:00")
	s.SessionSettings.Set(config.EndTime, "14:00:00")
	s.SessionSettings.Set(config.StartDay, "Sunday")
	s.SessionSettings.Set(config.EndDay, "blah")

	_, err = newSession(s.SessionID, s.MessageStoreFactory, s.SessionSettings, s.LogFactory, s.App)
	s.NotNil(err)
}

type SessionSuite struct {
	SessionSuiteRig
}

func TestSessionSuite(t *testing.T) {
	suite.Run(t, new(SessionSuite))
}

func (s *SessionSuite) SetupTest() {
	s.Init()
	s.session.store.Reset()
	s.session.State = latentState{}
}

func (s *SessionSuite) TestCheckCorrectCompID() {
	s.session.sessionID.TargetCompID = "TAR"
	s.session.sessionID.SenderCompID = "SND"

	var testCases = []struct {
		senderCompID *FIXString
		targetCompID *FIXString
		returnsError bool
		rejectReason int
	}{
		{returnsError: true, rejectReason: rejectReasonRequiredTagMissing},
		{senderCompID: newFIXString("TAR"),
			returnsError: true,
			rejectReason: rejectReasonRequiredTagMissing},
		{senderCompID: newFIXString("TAR"),
			targetCompID: newFIXString("JCD"),
			returnsError: true,
			rejectReason: rejectReasonCompIDProblem},
		{senderCompID: newFIXString("JCD"),
			targetCompID: newFIXString("SND"),
			returnsError: true,
			rejectReason: rejectReasonCompIDProblem},
		{senderCompID: newFIXString("TAR"),
			targetCompID: newFIXString("SND"),
			returnsError: false},
	}

	for _, tc := range testCases {
		msg := NewMessage()

		if tc.senderCompID != nil {
			msg.Header.SetField(tagSenderCompID, tc.senderCompID)
		}

		if tc.targetCompID != nil {
			msg.Header.SetField(tagTargetCompID, tc.targetCompID)
		}

		rej := s.session.checkCompID(msg)

		if !tc.returnsError {
			s.Require().Nil(rej)
			continue
		}

		s.NotNil(rej)
		s.Equal(tc.rejectReason, rej.RejectReason())
	}
}

func (s *SessionSuite) TestCheckBeginString() {
	msg := NewMessage()

	msg.Header.SetField(tagBeginString, FIXString("FIX.4.4"))
	err := s.session.checkBeginString(msg)
	s.Require().NotNil(err, "wrong begin string should return error")
	s.IsType(incorrectBeginString{}, err)

	msg.Header.SetField(tagBeginString, FIXString(s.session.sessionID.BeginString))
	s.Nil(s.session.checkBeginString(msg))
}

func (s *SessionSuite) TestCheckTargetTooHigh() {
	msg := NewMessage()
	s.Require().Nil(s.session.store.SetNextTargetMsgSeqNum(45))

	err := s.session.checkTargetTooHigh(msg)
	s.Require().NotNil(err, "missing sequence number should return error")
	s.Equal(rejectReasonRequiredTagMissing, err.RejectReason())

	msg.Header.SetField(tagMsgSeqNum, FIXInt(47))
	err = s.session.checkTargetTooHigh(msg)
	s.Require().NotNil(err, "sequence number too high should return an error")
	s.IsType(targetTooHigh{}, err)

	//spot on
	msg.Header.SetField(tagMsgSeqNum, FIXInt(45))
	s.Nil(s.session.checkTargetTooHigh(msg))
}

func (s *SessionSuite) TestCheckSendingTime() {
	msg := NewMessage()

	err := s.session.checkSendingTime(msg)
	s.Require().NotNil(err, "sending time is a required field")
	s.Equal(rejectReasonRequiredTagMissing, err.RejectReason())

	sendingTime := time.Now().Add(time.Duration(-200) * time.Second)
	msg.Header.SetField(tagSendingTime, FIXUTCTimestamp{Time: sendingTime})

	err = s.session.checkSendingTime(msg)
	s.Require().NotNil(err, "sending time too late should give error")
	s.Equal(rejectReasonSendingTimeAccuracyProblem, err.RejectReason())

	sendingTime = time.Now().Add(time.Duration(200) * time.Second)
	msg.Header.SetField(tagSendingTime, FIXUTCTimestamp{Time: sendingTime})

	err = s.session.checkSendingTime(msg)
	s.Require().NotNil(err, "future sending time should give error")
	s.Equal(rejectReasonSendingTimeAccuracyProblem, err.RejectReason())

	sendingTime = time.Now()
	msg.Header.SetField(tagSendingTime, FIXUTCTimestamp{Time: sendingTime})

	s.Nil(s.session.checkSendingTime(msg), "sending time should be ok")
}

func (s *SessionSuite) TestCheckTargetTooLow() {
	msg := NewMessage()
	s.Require().Nil(s.session.store.SetNextTargetMsgSeqNum(45))

	err := s.session.checkTargetTooLow(msg)
	s.Require().NotNil(err, "sequence number is required")
	s.Equal(rejectReasonRequiredTagMissing, err.RejectReason())

	//too low
	msg.Header.SetField(tagMsgSeqNum, FIXInt(43))
	err = s.session.checkTargetTooLow(msg)
	s.NotNil(err, "sequence number too low should return error")
	s.IsType(targetTooLow{}, err)

	//spot on
	msg.Header.SetField(tagMsgSeqNum, FIXInt(45))
	s.Nil(s.session.checkTargetTooLow(msg))
}

func (s *SessionSuite) TestCheckSessionTimeNoStartTimeEndTime() {
	var tests = []struct {
		before, after sessionState
	}{
		{before: latentState{}},
		{before: logonState{}},
		{before: logoutState{}},
		{before: inSession{}},
		{before: resendState{}},
		{before: pendingTimeout{resendState{}}},
		{before: pendingTimeout{inSession{}}},
		{before: notSessionTime{}, after: latentState{}},
	}

	for _, test := range tests {
		s.SetupTest()
		s.session.sessionTime = nil
		s.session.State = test.before

		s.session.CheckSessionTime(s.session, time.Now())
		if test.after != nil {
			s.State(test.after)
		} else {
			s.State(test.before)
		}
	}
}

func (s *SessionSuite) TestCheckSessionTimeInRange() {
	var tests = []struct {
		before, after sessionState
		expectReset   bool
	}{
		{before: latentState{}},
		{before: logonState{}},
		{before: logoutState{}},
		{before: inSession{}},
		{before: resendState{}},
		{before: pendingTimeout{resendState{}}},
		{before: pendingTimeout{inSession{}}},
		{before: notSessionTime{}, after: latentState{}, expectReset: true},
	}

	for _, test := range tests {
		s.SetupTest()
		s.session.State = test.before

		now := time.Now().UTC()
		store := new(memoryStore)
		if test.before.IsSessionTime() {
			store.Reset()
		} else {
			store.creationTime = now.Add(time.Duration(-1) * time.Minute)
		}
		s.session.store = store
		s.Nil(s.session.store.IncrNextSenderMsgSeqNum())
		s.Nil(s.session.store.IncrNextTargetMsgSeqNum())

		s.session.sessionTime = internal.NewUTCTimeRange(
			internal.NewTimeOfDay(now.Clock()),
			internal.NewTimeOfDay(now.Add(time.Hour).Clock()),
		)

		s.session.CheckSessionTime(s.session, now)
		if test.after != nil {
			s.State(test.after)
		} else {
			s.State(test.before)
		}

		if test.expectReset {
			s.ExpectStoreReset()
		} else {
			s.NextSenderMsgSeqNum(2)
			s.NextSenderMsgSeqNum(2)
		}
	}
}

func (s *SessionSuite) TestCheckSessionTimeNotInRange() {
	var tests = []struct {
		before           sessionState
		initiateLogon    bool
		expectOnLogout   bool
		expectSendLogout bool
	}{
		{before: latentState{}},
		{before: logonState{}},
		{before: logonState{}, initiateLogon: true, expectOnLogout: true},
		{before: logoutState{}, expectOnLogout: true},
		{before: inSession{}, expectOnLogout: true, expectSendLogout: true},
		{before: resendState{}, expectOnLogout: true, expectSendLogout: true},
		{before: pendingTimeout{resendState{}}, expectOnLogout: true, expectSendLogout: true},
		{before: pendingTimeout{inSession{}}, expectOnLogout: true, expectSendLogout: true},
		{before: notSessionTime{}},
	}

	for _, test := range tests {
		s.SetupTest()
		s.session.State = test.before
		s.session.initiateLogon = test.initiateLogon
		s.Nil(s.session.store.IncrNextSenderMsgSeqNum())
		s.Nil(s.session.store.IncrNextTargetMsgSeqNum())

		now := time.Now().UTC()
		s.session.sessionTime = internal.NewUTCTimeRange(
			internal.NewTimeOfDay(now.Add(time.Hour).Clock()),
			internal.NewTimeOfDay(now.Add(time.Duration(2)*time.Hour).Clock()),
		)

		if test.expectOnLogout {
			s.MockApp.On("OnLogout")
		}
		if test.expectSendLogout {
			s.MockApp.On("ToAdmin")
		}
		s.session.CheckSessionTime(s.session, now)

		s.MockApp.AssertExpectations(s.T())
		s.State(notSessionTime{})

		s.NextTargetMsgSeqNum(2)
		if test.expectSendLogout {
			s.LastToAdminMessageSent()
			s.MessageType(enum.MsgType_LOGOUT, s.MockApp.lastToAdmin)
			s.NextSenderMsgSeqNum(3)
		} else {
			s.NextSenderMsgSeqNum(2)
		}
	}
}

func (s *SessionSuite) TestCheckSessionTimeInRangeButNotSameRangeAsStore() {
	var tests = []struct {
		before           sessionState
		initiateLogon    bool
		expectOnLogout   bool
		expectSendLogout bool
	}{
		{before: latentState{}},
		{before: logonState{}},
		{before: logonState{}, initiateLogon: true, expectOnLogout: true},
		{before: logoutState{}, expectOnLogout: true},
		{before: inSession{}, expectOnLogout: true, expectSendLogout: true},
		{before: resendState{}, expectOnLogout: true, expectSendLogout: true},
		{before: pendingTimeout{resendState{}}, expectOnLogout: true, expectSendLogout: true},
		{before: pendingTimeout{inSession{}}, expectOnLogout: true, expectSendLogout: true},
		{before: notSessionTime{}},
	}

	for _, test := range tests {
		s.SetupTest()
		s.session.State = test.before
		s.session.initiateLogon = test.initiateLogon
		s.store.Reset()
		s.Nil(s.session.store.IncrNextSenderMsgSeqNum())
		s.Nil(s.session.store.IncrNextTargetMsgSeqNum())

		now := time.Now().UTC()
		s.session.sessionTime = internal.NewUTCTimeRange(
			internal.NewTimeOfDay(now.Add(time.Duration(-1)*time.Hour).Clock()),
			internal.NewTimeOfDay(now.Add(time.Hour).Clock()),
		)

		if test.expectOnLogout {
			s.MockApp.On("OnLogout")
		}
		if test.expectSendLogout {
			s.MockApp.On("ToAdmin")
		}
		s.session.CheckSessionTime(s.session, now.AddDate(0, 0, 1))

		s.MockApp.AssertExpectations(s.T())
		s.State(latentState{})
		if test.expectSendLogout {
			s.LastToAdminMessageSent()
			s.MessageType(enum.MsgType_LOGOUT, s.MockApp.lastToAdmin)
			s.FieldEquals(tagMsgSeqNum, 2, s.MockApp.lastToAdmin.Header)
		}
		s.ExpectStoreReset()
	}
}

func (s *SessionSuite) TestOnAdminConnectInitiateLogon() {
	adminMsg := connect{
		messageOut:    s.Receiver.sendChannel,
		initiateLogon: true,
	}
	s.session.State = latentState{}
	s.session.heartBtInt = time.Duration(45) * time.Second
	s.session.store.IncrNextSenderMsgSeqNum()

	s.MockApp.On("ToAdmin")
	s.session.onAdmin(adminMsg)

	s.MockApp.AssertExpectations(s.T())
	s.True(s.session.initiateLogon)
	s.State(logonState{})
	s.LastToAdminMessageSent()
	s.MessageType(enum.MsgType_LOGON, s.MockApp.lastToAdmin)
	s.FieldEquals(tagHeartBtInt, 45, s.MockApp.lastToAdmin.Body)
	s.FieldEquals(tagMsgSeqNum, 2, s.MockApp.lastToAdmin.Header)
	s.NextSenderMsgSeqNum(3)
}

func (s *SessionSuite) TestOnAdminConnectAccept() {
	adminMsg := connect{
		messageOut: s.Receiver.sendChannel,
	}
	s.session.State = latentState{}
	s.session.store.IncrNextSenderMsgSeqNum()

	s.session.onAdmin(adminMsg)
	s.False(s.session.initiateLogon)
	s.State(logonState{})
	s.NoMessageSent()
	s.NextSenderMsgSeqNum(2)
}

func (s *SessionSuite) TestOnAdminConnectNotInSession() {
	var tests = []bool{true, false}

	for _, doInitiateLogon := range tests {
		s.SetupTest()
		s.session.State = notSessionTime{}
		s.session.store.IncrNextSenderMsgSeqNum()

		adminMsg := connect{
			messageOut:    s.Receiver.sendChannel,
			initiateLogon: doInitiateLogon,
		}

		s.session.onAdmin(adminMsg)

		s.State(notSessionTime{})
		s.NoMessageSent()
		s.Disconnected()
		s.NextSenderMsgSeqNum(2)
	}
}

func (s *SessionSuite) TestOnAdminStop() {
	s.session.State = logonState{}

	s.session.onAdmin(stopReq{})
	s.Disconnected()
	s.Stopped()
}

type SessionSendTestSuite struct {
	SessionSuiteRig
}

func TestSessionSendTestSuite(t *testing.T) {
	suite.Run(t, new(SessionSendTestSuite))
}

func (suite *SessionSendTestSuite) SetupTest() {
	suite.Init()
	suite.session.State = inSession{}
}

func (suite *SessionSendTestSuite) TestQueueForSendAppMessage() {
	suite.MockApp.On("ToApp").Return(nil)
	require.Nil(suite.T(), suite.queueForSend(suite.NewOrderSingle()))

	suite.MockApp.AssertExpectations(suite.T())
	suite.NoMessageSent()
	suite.MessagePersisted(suite.MockApp.lastToApp)
	suite.FieldEquals(tagMsgSeqNum, 1, suite.MockApp.lastToApp.Header)
	suite.NextSenderMsgSeqNum(2)
}

func (suite *SessionSendTestSuite) TestQueueForSendDoNotSendAppMessage() {
	suite.MockApp.On("ToApp").Return(DoNotSend)
	suite.Equal(DoNotSend, suite.queueForSend(suite.NewOrderSingle()))

	suite.MockApp.AssertExpectations(suite.T())
	suite.NoMessagePersisted(1)
	suite.NoMessageSent()
	suite.NextSenderMsgSeqNum(1)

	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.send(suite.Heartbeat()))

	suite.MockApp.AssertExpectations(suite.T())
	suite.LastToAdminMessageSent()
	suite.MessagePersisted(suite.MockApp.lastToAdmin)
	suite.NextSenderMsgSeqNum(2)
}

func (suite *SessionSendTestSuite) TestQueueForSendAdminMessage() {
	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.queueForSend(suite.Heartbeat()))

	suite.MockApp.AssertExpectations(suite.T())
	suite.MessagePersisted(suite.MockApp.lastToAdmin)
	suite.NoMessageSent()
	suite.NextSenderMsgSeqNum(2)
}

func (suite *SessionSendTestSuite) TestSendAppMessage() {
	suite.MockApp.On("ToApp").Return(nil)
	require.Nil(suite.T(), suite.send(suite.NewOrderSingle()))

	suite.MockApp.AssertExpectations(suite.T())
	suite.MessagePersisted(suite.MockApp.lastToApp)
	suite.LastToAppMessageSent()
	suite.NextSenderMsgSeqNum(2)
}

func (suite *SessionSendTestSuite) TestSendAppDoNotSendMessage() {
	suite.MockApp.On("ToApp").Return(DoNotSend)
	suite.Equal(DoNotSend, suite.send(suite.NewOrderSingle()))

	suite.MockApp.AssertExpectations(suite.T())
	suite.NextSenderMsgSeqNum(1)
	suite.NoMessageSent()
}

func (suite *SessionSendTestSuite) TestSendAdminMessage() {
	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.send(suite.Heartbeat()))
	suite.MockApp.AssertExpectations(suite.T())

	suite.LastToAdminMessageSent()
	suite.MessagePersisted(suite.MockApp.lastToAdmin)
}

func (suite *SessionSendTestSuite) TestSendFlushesQueue() {
	suite.MockApp.On("ToApp").Return(nil)
	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.queueForSend(suite.NewOrderSingle()))
	require.Nil(suite.T(), suite.queueForSend(suite.Heartbeat()))

	order1 := suite.MockApp.lastToApp
	heartbeat := suite.MockApp.lastToAdmin

	suite.MockApp.AssertExpectations(suite.T())
	suite.NoMessageSent()

	suite.MockApp.On("ToApp").Return(nil)
	require.Nil(suite.T(), suite.send(suite.NewOrderSingle()))
	suite.MockApp.AssertExpectations(suite.T())
	order2 := suite.MockApp.lastToApp
	suite.MessageSentEquals(order1)
	suite.MessageSentEquals(heartbeat)
	suite.MessageSentEquals(order2)
	suite.NoMessageSent()
}

func (suite *SessionSendTestSuite) TestSendNotLoggedOn() {
	suite.MockApp.On("ToApp").Return(nil)
	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.queueForSend(suite.NewOrderSingle()))
	require.Nil(suite.T(), suite.queueForSend(suite.Heartbeat()))

	suite.MockApp.AssertExpectations(suite.T())
	suite.NoMessageSent()

	var tests = []sessionState{logoutState{}, latentState{}, logonState{}}

	for _, test := range tests {
		suite.MockApp.On("ToApp").Return(nil)
		suite.session.State = test
		require.Nil(suite.T(), suite.send(suite.NewOrderSingle()))
		suite.MockApp.AssertExpectations(suite.T())
		suite.NoMessageSent()
	}
}

func (suite *SessionSendTestSuite) TestDropAndSendAdminMessage() {
	suite.MockApp.On("ToAdmin")
	suite.Require().Nil(suite.dropAndSend(suite.Heartbeat(), false))
	suite.MockApp.AssertExpectations(suite.T())

	suite.MessagePersisted(suite.MockApp.lastToAdmin)
	suite.LastToAdminMessageSent()
}

func (suite *SessionSendTestSuite) TestDropAndSendDropsQueue() {
	suite.MockApp.On("ToApp").Return(nil)
	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.queueForSend(suite.NewOrderSingle()))
	require.Nil(suite.T(), suite.queueForSend(suite.Heartbeat()))
	suite.MockApp.AssertExpectations(suite.T())

	suite.NoMessageSent()

	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.dropAndSend(suite.Logon(), false))
	suite.MockApp.AssertExpectations(suite.T())

	msg := suite.MockApp.lastToAdmin
	suite.MessageType(enum.MsgType_LOGON, msg)
	suite.FieldEquals(tagMsgSeqNum, 3, msg.Header)

	//only one message sent
	suite.LastToAdminMessageSent()
	suite.NoMessageSent()
}

func (suite *SessionSendTestSuite) TestDropAndSendDropsQueueWithReset() {
	suite.MockApp.On("ToApp").Return(nil)
	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.queueForSend(suite.NewOrderSingle()))
	require.Nil(suite.T(), suite.queueForSend(suite.Heartbeat()))
	suite.MockApp.AssertExpectations(suite.T())
	suite.NoMessageSent()

	suite.MockApp.On("ToAdmin")
	require.Nil(suite.T(), suite.dropAndSend(suite.Logon(), true))
	suite.MockApp.AssertExpectations(suite.T())
	msg := suite.MockApp.lastToAdmin

	suite.MessageType(enum.MsgType_LOGON, msg)
	suite.FieldEquals(tagMsgSeqNum, 1, msg.Header)

	//only one message sent
	suite.LastToAdminMessageSent()
	suite.NoMessageSent()
}
