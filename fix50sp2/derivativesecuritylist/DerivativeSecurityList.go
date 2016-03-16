//Package derivativesecuritylist msg type = AA.
package derivativesecuritylist

import (
	"github.com/quickfixgo/quickfix"
	"github.com/quickfixgo/quickfix/enum"
	"github.com/quickfixgo/quickfix/fix50sp2/applicationsequencecontrol"
	"github.com/quickfixgo/quickfix/fix50sp2/derivativesecuritydefinition"
	"github.com/quickfixgo/quickfix/fix50sp2/relsymderivsecgrp"
	"github.com/quickfixgo/quickfix/fix50sp2/underlyinginstrument"
	"github.com/quickfixgo/quickfix/fixt11"
	"time"
)

//Message is a DerivativeSecurityList FIX Message
type Message struct {
	FIXMsgType string `fix:"AA"`
	fixt11.Header
	//SecurityReqID is a non-required field for DerivativeSecurityList.
	SecurityReqID *string `fix:"320"`
	//SecurityResponseID is a non-required field for DerivativeSecurityList.
	SecurityResponseID *string `fix:"322"`
	//SecurityRequestResult is a non-required field for DerivativeSecurityList.
	SecurityRequestResult *int `fix:"560"`
	//UnderlyingInstrument is a non-required component for DerivativeSecurityList.
	UnderlyingInstrument *underlyinginstrument.UnderlyingInstrument
	//TotNoRelatedSym is a non-required field for DerivativeSecurityList.
	TotNoRelatedSym *int `fix:"393"`
	//LastFragment is a non-required field for DerivativeSecurityList.
	LastFragment *bool `fix:"893"`
	//RelSymDerivSecGrp is a non-required component for DerivativeSecurityList.
	RelSymDerivSecGrp *relsymderivsecgrp.RelSymDerivSecGrp
	//DerivativeSecurityDefinition is a non-required component for DerivativeSecurityList.
	DerivativeSecurityDefinition *derivativesecuritydefinition.DerivativeSecurityDefinition
	//ApplicationSequenceControl is a non-required component for DerivativeSecurityList.
	ApplicationSequenceControl *applicationsequencecontrol.ApplicationSequenceControl
	//SecurityReportID is a non-required field for DerivativeSecurityList.
	SecurityReportID *int `fix:"964"`
	//ClearingBusinessDate is a non-required field for DerivativeSecurityList.
	ClearingBusinessDate *string `fix:"715"`
	//TransactTime is a non-required field for DerivativeSecurityList.
	TransactTime *time.Time `fix:"60"`
	fixt11.Trailer
}

//Marshal converts Message to a quickfix.Message instance
func (m Message) Marshal() quickfix.Message { return quickfix.Marshal(m) }

func (m *Message) SetSecurityReqID(v string)      { m.SecurityReqID = &v }
func (m *Message) SetSecurityResponseID(v string) { m.SecurityResponseID = &v }
func (m *Message) SetSecurityRequestResult(v int) { m.SecurityRequestResult = &v }
func (m *Message) SetUnderlyingInstrument(v underlyinginstrument.UnderlyingInstrument) {
	m.UnderlyingInstrument = &v
}
func (m *Message) SetTotNoRelatedSym(v int) { m.TotNoRelatedSym = &v }
func (m *Message) SetLastFragment(v bool)   { m.LastFragment = &v }
func (m *Message) SetRelSymDerivSecGrp(v relsymderivsecgrp.RelSymDerivSecGrp) {
	m.RelSymDerivSecGrp = &v
}
func (m *Message) SetDerivativeSecurityDefinition(v derivativesecuritydefinition.DerivativeSecurityDefinition) {
	m.DerivativeSecurityDefinition = &v
}
func (m *Message) SetApplicationSequenceControl(v applicationsequencecontrol.ApplicationSequenceControl) {
	m.ApplicationSequenceControl = &v
}
func (m *Message) SetSecurityReportID(v int)        { m.SecurityReportID = &v }
func (m *Message) SetClearingBusinessDate(v string) { m.ClearingBusinessDate = &v }
func (m *Message) SetTransactTime(v time.Time)      { m.TransactTime = &v }

//A RouteOut is the callback type that should be implemented for routing Message
type RouteOut func(msg Message, sessionID quickfix.SessionID) quickfix.MessageRejectError

//Route returns the beginstring, message type, and MessageRoute for this Message type
func Route(router RouteOut) (string, string, quickfix.MessageRoute) {
	r := func(msg quickfix.Message, sessionID quickfix.SessionID) quickfix.MessageRejectError {
		m := new(Message)
		if err := quickfix.Unmarshal(msg, m); err != nil {
			return err
		}
		return router(*m, sessionID)
	}
	return enum.ApplVerID_FIX50SP2, "AA", r
}
