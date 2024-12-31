package message

type MType byte

 const (
    JOINR MType = iota
    ACC
    REJ
    NEWR
    NEW
    IDENTR
    IDENT
    CHT
    CHTE
)
