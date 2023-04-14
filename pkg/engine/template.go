package engine

const (
	EndMagic = "!!"
	EndLen   = 0
)

func EndPrototype() Codec      { return &End{} }
func init()                    { Register(EndMagic, EndPrototype) }
func (x *End) Magic() string   { return EndMagic }
func (x *End) Len() int        { return EndLen }
func (x *End) Wrap(inner Mung) {}
func (x *End) GetMung() Mung   { return x }

type End struct{}

func (o Skins) End() Skins {
	return append(o, &End{})
}

func NewEnd() *End {
	return &End{}
}

func (x *End) Encode(s *Splice) (e error) {
	return
}

func (x *End) Decode(s *Splice) (e error) {
	return
}

func (x *End) Handle(s *Splice, p Mung,
	ng *Engine) (e error) {
	
	return
}

func (x *End) Account(res *SendData, sm *SessionManager,
	s *SessionData, last bool) (skip bool, sd *SessionData) {
	return
}
