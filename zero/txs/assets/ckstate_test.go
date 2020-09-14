package assets

import (
	"testing"

	"github.com/dece-cash/go-dece/czero/c_type"

	"github.com/dece-cash/go-dece/zero/utils"
)

var dece_token = Token{
	utils.CurrencyToUint256("DECEs"),
	utils.NewU256(100),
}

var tk_ticket = Ticket{
	utils.CurrencyToUint256("TK"),
	c_type.RandUint256(),
}

var token_asset = Asset{&dece_token, nil}

var ticket_asset = Asset{nil, &tk_ticket}

var asset = Asset{&dece_token, &tk_ticket}

func TestCkState_OutPlus(t *testing.T) {
	ck := NewCKState(true, &dece_token)
	if ck.Check() == nil {
		t.Fail()
	}
	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	ck.AddIn(&ticket_asset)
	if ck.Check() == nil {
		t.Fail()
	}

	ck.AddOut(&asset)
	if ck.Check() == nil {
		t.Fail()
	}

	tkns, tkts := ck.GetList()
	if len(tkns) != 1 {
		t.Fail()
	}
	if len(tkts) != 0 {
		t.Fail()
	}
	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	tkns, tkts = ck.GetList()
	if len(tkns) != 0 {
		t.Fail()
	}
	if len(tkts) != 0 {
		t.Fail()
	}

}

func TestCkState_InPlus(t *testing.T) {
	ck := NewCKState(false, &dece_token)
	if ck.Check() == nil {
		t.Fail()
	}
	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	ck.AddIn(&ticket_asset)
	if ck.Check() == nil {
		t.Fail()
	}

	tkns, tkts := ck.GetList()
	if len(tkns) != 0 {
		t.Fail()
	}
	if len(tkts) != 1 {
		t.Fail()
	}

	ck.AddOut(&asset)
	if ck.Check() == nil {
		t.Fail()
	}

	ck.AddIn(&token_asset)
	if ck.Check() != nil {
		t.Fail()
	}
	tkns, tkts = ck.GetList()
	if len(tkns) != 0 {
		t.Fail()
	}
	if len(tkts) != 0 {
		t.Fail()
	}

}
