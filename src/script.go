package main

import (
	"errors"
	"os"
	"strings"
)

type Operator int

const NumberOfReadByte int = 1024

const (
	Sentence Operator = iota
	PageFeed
	PlayMusic
	StopMusic
	SuspendMusic
	ResumeMusic
)

type Action struct {
	operator Operator
	value    string
}

type Script struct {
	file     *os.File
	buffer   []byte
	end      bool
	nextLine string
}

func NewScript(file *os.File) *Script {
	return &Script{file: file}
}

func (s *Script) ReadLine() (string, error) {
	var ss []string
	for i := 0; ; i++ {
		if i >= len(s.buffer) {
			if s.end {
				break
			}
			ss = append(ss, string(s.buffer))
			s.buffer = make([]byte, NumberOfReadByte)
			_, err := s.file.Read(s.buffer)
			if err != nil {
				if err.Error() == "EOF" {
					s.end = true
				}
			}
			i = 0
		} else if s.buffer[i] == 10 {
			ss = append(ss, string(s.buffer[:i]))
			s.buffer = s.buffer[i+1:]
			break
		}
	}
	if len(ss) == 0 {
		return "", errors.New("END")
	}
	return strings.Join(ss, ""), nil
}

func (s *Script) Next() ([]*Action, error) {
	actions := []*Action{}
	isSentence := false
	var line string
	var err error

	for {
		if s.nextLine != "" {
			line = s.nextLine
			s.nextLine = ""
		} else {
			line, err = s.ReadLine()
			if err != nil {
				return nil, err
			}
		}

		if line == "" {
			if actions[len(actions)-1].operator != PageFeed {
				actions = append(actions, &Action{operator: PageFeed, value: ""})
			}
		} else if line[0] == '\\' && line[1] != '\\' {
		} else if isSentence {
			s.nextLine = line
			break
		} else {
			actions = append(actions, &Action{operator: Sentence, value: line})
			isSentence = true
		}
	}
	return actions, nil
}
