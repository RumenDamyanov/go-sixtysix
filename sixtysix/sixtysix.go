package sixtysix

import (
	"errors"
	"math/rand"
	"slices"

	"go.rumenx.com/sixtysix/engine"
)

// Ranks and suits
const (
	Clubs = iota
	Diamonds
	Hearts
	Spades
)

var suits = []int{Clubs, Diamonds, Hearts, Spades}
var rankOrder = []int{11, 10, 4, 3, 2, 0}

func card(suit int, rankVal int) int { return suit*100 + rankVal }
func cardSuit(c int) int             { return c / 100 }
func cardVal(c int) int              { return c % 100 }
func trickPoints(c int) int          { return cardVal(c) }

type State struct {
	Current   int      `json:"current"`
	Scores    [2]int   `json:"scores"`
	Hands     [2][]int `json:"hands"`
	Stock     []int    `json:"stock"`
	Closed    bool     `json:"closed"`
	TrumpSuit int      `json:"trumpSuit"`
	TrumpCard int      `json:"trumpCard"`
	Trick     []int    `json:"trick"`
	Winner    int      `json:"winner"`
}

const (
	ActionDeal       = "deal"
	ActionPlay       = "play"
	ActionCloseStock = "closeStock"
	ActionDeclare    = "declare"
	ActionExchange   = "exchangeTrump"
)

type Game struct{}

func (Game) Name() string { return "sixtysix" }

func (Game) InitialState(seed int64) any {
	r := rand.New(rand.NewSource(seed))
	deck := newDeck()
	r.Shuffle(len(deck), func(i, j int) { deck[i], deck[j] = deck[j], deck[i] })
	trumpCard := deck[len(deck)-1]
	trumpSuit := cardSuit(trumpCard)
	hands := [2][]int{{}, {}}
	deal := func(p, n int) {
		for k := 0; k < n; k++ {
			hands[p] = append(hands[p], deck[0])
			deck = deck[1:]
		}
	}
	deal(0, 3)
	deal(1, 3)
	deal(0, 3)
	deal(1, 3)
	stock := append([]int(nil), deck[:len(deck)-1]...)
	st := State{Current: 0, Scores: [2]int{0, 0}, Hands: hands, Stock: stock, Closed: false, TrumpSuit: trumpSuit, TrumpCard: trumpCard, Winner: -1}
	for i := 0; i < 2; i++ {
		slices.SortFunc(st.Hands[i], func(a, b int) int {
			if cardSuit(a) != cardSuit(b) {
				return cardSuit(a) - cardSuit(b)
			}
			return cardVal(a) - cardVal(b)
		})
	}
	return st
}

func (Game) Validate(s any, a engine.Action) error {
	st := s.(State)
	if st.Winner != -1 {
		return errors.New("game over")
	}
	switch a.Type {
	case ActionPlay:
		c, ok := getInt(a.Payload, "card")
		if !ok {
			return errors.New("missing card")
		}
		if !contains(st.Hands[st.Current], c) {
			return errors.New("card not in hand")
		}
		if len(st.Trick) == 1 && (st.Closed || len(st.Stock) == 0) {
			lead := st.Trick[0]
			ls := cardSuit(lead)
			if cardSuit(c) != ls && hasSuit(st.Hands[st.Current], ls) {
				return errors.New("must follow suit")
			}
		}
		return nil
	case ActionCloseStock:
		if st.Closed || len(st.Stock) == 0 {
			return errors.New("cannot close")
		}
		if len(st.Trick) != 0 {
			return errors.New("cannot close mid-trick")
		}
		return nil
	case ActionDeclare:
		suit, ok := getInt(a.Payload, "suit")
		if !ok {
			return errors.New("missing suit")
		}
		k, q := card(suit, 4), card(suit, 3)
		if !(contains(st.Hands[st.Current], k) && contains(st.Hands[st.Current], q)) {
			return errors.New("no marriage")
		}
		if len(st.Trick) != 0 {
			return errors.New("declare only on lead")
		}
		return nil
	case ActionExchange:
		if st.Closed || len(st.Stock) == 0 {
			return errors.New("cannot exchange when stock closed or empty")
		}
		if len(st.Trick) != 0 {
			return errors.New("exchange only at lead")
		}
		if !contains(st.Hands[st.Current], card(st.TrumpSuit, 0)) {
			return errors.New("no nine of trump to exchange")
		}
		return nil
	default:
		return errors.New("unknown action")
	}
}

func (Game) Apply(s any, a engine.Action) (any, error) {
	st := s.(State)
	switch a.Type {
	case ActionPlay:
		c, _ := getInt(a.Payload, "card")
		actor := st.Current
		st.Hands[actor] = remove(st.Hands[actor], c)
		st.Trick = append(st.Trick, c)
		if len(st.Trick) == 2 {
			winner := trickWinner(st.Trick[0], st.Trick[1], st.TrumpSuit)
			pts := trickPoints(st.Trick[0]) + trickPoints(st.Trick[1])
			st.Scores[winner] += pts
			st.Trick = st.Trick[:0]
			if !st.Closed && len(st.Stock) > 0 {
				if len(st.Stock) >= 2 {
					st.Hands[winner] = append(st.Hands[winner], st.Stock[0])
					st.Hands[1-winner] = append(st.Hands[1-winner], st.Stock[1])
					st.Stock = st.Stock[2:]
				} else {
					st.Hands[winner] = append(st.Hands[winner], st.Stock[0])
					st.Stock = st.Stock[:0]
				}
			}
			st.Current = winner
			if len(st.Hands[0])+len(st.Hands[1]) == 0 {
				st.Scores[winner] += 10
			}
			if st.Scores[winner] >= 66 {
				st.Winner = winner
			}
		} else {
			st.Current = 1 - actor
		}
		return st, nil
	case ActionCloseStock:
		st.Closed = true
		return st, nil
	case ActionDeclare:
		suit, _ := getInt(a.Payload, "suit")
		pts := 20
		if suit == st.TrumpSuit {
			pts = 40
		}
		st.Scores[st.Current] += pts
		if st.Scores[st.Current] >= 66 {
			st.Winner = st.Current
		}
		return st, nil
	case ActionExchange:
		nine := card(st.TrumpSuit, 0)
		st.Hands[st.Current] = remove(st.Hands[st.Current], nine)
		st.Hands[st.Current] = append(st.Hands[st.Current], st.TrumpCard)
		st.TrumpCard = nine
		return st, nil
	default:
		return s, errors.New("unknown action")
	}
}

func newDeck() []int {
	d := make([]int, 0, 24)
	for _, s := range suits {
		for _, rv := range rankOrder {
			d = append(d, card(s, rv))
		}
	}
	return d
}
func contains(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
func hasSuit(xs []int, suit int) bool {
	for _, x := range xs {
		if cardSuit(x) == suit {
			return true
		}
	}
	return false
}
func remove(xs []int, v int) []int {
	for i, x := range xs {
		if x == v {
			return append(append([]int(nil), xs[:i]...), xs[i+1:]...)
		}
	}
	return xs
}
func trickWinner(lead, follow int, trump int) int {
	ls, fs := cardSuit(lead), cardSuit(follow)
	if fs == ls {
		if cardVal(follow) > cardVal(lead) {
			return 1
		}
		return 0
	}
	if fs == trump && ls != trump {
		return 1
	}
	return 0
}
func getInt(m map[string]any, k string) (int, bool) {
	if v, ok := m[k]; ok {
		switch t := v.(type) {
		case float64:
			return int(t), true
		case int:
			return t, true
		}
	}
	return 0, false
}
