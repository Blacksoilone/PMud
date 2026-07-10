package command

import "strings"

type Command struct {
	Raw  string
	Verb string
	Args []string
}

func Parse(line string) Command {
	split := strings.Fields(line)
	if len(split) == 0 {
		return Command{
			Raw:  line,
			Verb: "",
			Args: []string{},
		}
	}
	return Command{
		Raw:  line,
		Verb: split[0],
		Args: split[1:],
	}
}
