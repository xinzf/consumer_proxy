package errno

import "fmt"

type Errno struct {
	Code           int
	Message        string
	customMessages []string
}

func (this *Errno) Error() string {
	msg := this.Message

	for _, c := range this.customMessages {
		msg += " " + c
	}

	msg += fmt.Sprintf(" [code: %d]", this.Code)

	this.customMessages = []string{}
	return msg
}

func (this *Errno) Add(msg string) *Errno {
	this.customMessages = append(this.customMessages, msg)
	return this
}
