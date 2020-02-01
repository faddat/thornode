package common

import (
	"fmt"
	"strings"
)

var EmptyBlame = Blame{
	FailReason: "",
	BlameNodes: make(PubKeys, 0),
}

type Blame struct {
	FailReason string  `json:"fail_reason"`
	BlameNodes PubKeys `json:"blame_peers,omitempty"`
}

func (b Blame) IsEmpty() bool {
	return len(b.BlameNodes) == 0 || len(b.FailReason) == 0
}

func (b Blame) String() string {
	sb := strings.Builder{}
	sb.WriteString("reason:" + b.FailReason + "\n")
	sb.WriteString("nodes:" + fmt.Sprintf("%+v", b.BlameNodes) + "\n")
	return sb.String()
}
