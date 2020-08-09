package common

type Operation struct {
	Opt         string
	ElapsedTime int64
}

type Chat struct {
	Username    string
	MsgType     string
	Message     string
	Source      string
	ElapsedTime int64
	Uid         int
}
