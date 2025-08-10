package sixtysix

import (
	"testing"

	"go.rumenx.com/sixtysix/engine"
)

func actionPlay(card int) engine.Action {
	return engine.Action{Type: ActionPlay, Payload: map[string]any{"card": card}}
}

func TestInitialDealDeterministic(t *testing.T) {
	g := Game{}
	a := g.InitialState(42).(State)
	b := g.InitialState(42).(State)
	if a.TrumpSuit != b.TrumpSuit || a.TrumpCard != b.TrumpCard {
		t.Fatalf("expected deterministic trump")
	}
	if len(a.Hands[0]) != 6 || len(a.Hands[1]) != 6 || len(a.Stock) != 24-1-12 { // deck minus trump and initial hands
		t.Fatalf("unexpected deal sizes: %+v", a)
	}
}

func TestPlayAndTrickResolution(t *testing.T) {
	g := Game{}
	st := g.InitialState(1).(State)
	lead := st.Hands[st.Current][0]
	if err := g.Validate(st, actionPlay(lead)); err != nil {
		t.Fatalf("validate lead: %v", err)
	}
	ns, err := g.Apply(st, actionPlay(lead))
	if err != nil {
		t.Fatalf("apply lead: %v", err)
	}
	st = ns.(State)
	follow := st.Hands[st.Current][0]
	if err := g.Validate(st, actionPlay(follow)); err != nil {
		t.Fatalf("validate follow: %v", err)
	}
	ns, err = g.Apply(st, actionPlay(follow))
	if err != nil {
		t.Fatalf("apply follow: %v", err)
	}
	st = ns.(State)
	if len(st.Trick) != 0 {
		t.Fatalf("trick should be resolved")
	}
}

func TestCloseStockEnforcesFollowSuit(t *testing.T) {
	g := Game{}
	st := g.InitialState(7).(State)
	st.Closed = true
	lead := st.Hands[st.Current][0]
	ns, _ := g.Apply(st, actionPlay(lead))
	st = ns.(State)
	follower := 1 - st.Current
	ls := cardSuit(lead)
	var off int
	for _, c := range st.Hands[follower] {
		if cardSuit(c) != ls {
			off = c
			break
		}
	}
	if off != 0 { // has off-suit card but must follow if possible
		if err := g.Validate(st, actionPlay(off)); err == nil {
			t.Fatalf("expected follow-suit requirement when closed")
		}
	}
}

func TestDeclareMarriageAndExchange(t *testing.T) {
	g := Game{}
	st := g.InitialState(99).(State)
	if err := g.Validate(st, engine.Action{Type: ActionDeclare, Payload: map[string]any{"suit": st.TrumpSuit}}); err == nil {
		ns, _ := g.Apply(st, engine.Action{Type: ActionDeclare, Payload: map[string]any{"suit": st.TrumpSuit}})
		st = ns.(State)
	}
	if err := g.Validate(st, engine.Action{Type: ActionExchange}); err == nil {
		ns, _ := g.Apply(st, engine.Action{Type: ActionExchange})
		st = ns.(State)
		if cardSuit(st.TrumpCard) != st.TrumpSuit || cardVal(st.TrumpCard) != 0 {
			t.Fatalf("expected 9 of trump after exchange")
		}
	}
}

func TestLastTrickBonus(t *testing.T) {
	g := Game{}
	st := State{Current: 0, Scores: [2]int{0, 0}, Hands: [2][]int{{card(Hearts, 0)}, {card(Hearts, 11)}}, Stock: nil, Closed: true, TrumpSuit: Spades, TrumpCard: card(Spades, 0), Winner: -1}
	ns, err := g.Apply(st, actionPlay(card(Hearts, 0)))
	if err != nil {
		t.Fatalf("lead apply: %v", err)
	}
	st = ns.(State)
	ns, err = g.Apply(st, actionPlay(card(Hearts, 11)))
	if err != nil {
		t.Fatalf("follow apply: %v", err)
	}
	st = ns.(State)
	if len(st.Hands[0])+len(st.Hands[1]) != 0 {
		t.Fatalf("expected empty hands at end")
	}
	if st.Scores[1] < 21 {
		t.Fatalf("expected last trick bonus applied, scores=%v", st.Scores)
	}
}
