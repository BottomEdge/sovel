package main

import (
	"os"
	"time"

	"github.com/gdamore/tcell"
	"github.com/mattn/go-runewidth"
)

type Mode int

const (
	NovelTitle = "sovel project"
)

const (
	TextSpeed = 20
)

const (
	Title Mode = iota
	Novel
)

type View struct {
	screen        tcell.Screen
	width, height int
	mode          Mode
	x, y          int
	userInput     chan rune
	textSpeed     int
	duringLine    bool
	script        *Script
	nextNewPage   bool
}

func NewView(file *os.File) (*View, error) {
	screen, err := tcell.NewScreen()
	if err != nil {
		return nil, err
	}
	if err = screen.Init(); err != nil {
		return nil, err
	}
	width, height := screen.Size()

	userInput := make(chan rune)

	return &View{
		width:     width,
		height:    height,
		screen:    screen,
		userInput: userInput,
		textSpeed: TextSpeed,
		script:    NewScript(file),
	}, nil
}

func (v *View) Close() {
	v.screen.Fini()
}

func (v *View) Loop(quit chan struct{}) {
	for {
		ev := v.screen.PollEvent()
		switch ev := ev.(type) {
		case *tcell.EventKey:
			switch ev.Key() {
			case tcell.KeyRune:
				r := ev.Rune()
				if r == 'q' {
					close(quit)
				} else {
					switch v.mode {
					case Title:
						v.InputTitle(r)
					case Novel:
						_ = v.InputNovel(r)
					}
				}
			}
		}

	}
}

func (v *View) InputTitle(r rune) {
	v.mode = Novel
	v.CleanPage()
}

func (v *View) InputNovel(r rune) error {
	switch r {
	case 'j':
		if v.duringLine {
			v.userInput <- r
		} else {
			if v.nextNewPage {
				v.userInput <- r
				v.CleanPage()
				v.nextNewPage = false
			}
			actions, err := v.script.Next()
			if err != nil {
				return err
			}
			for _, action := range actions {
				switch action.operator {
				case Sentence:
					go v.NewLine(action.value)
				case PageFeed:
					v.nextNewPage = true
					go v.PageFeedSymbol()
				case PlayMusic:
				case StopMusic:
				case SuspendMusic:
				case ResumeMusic:
				}
			}
		}
	}
	return nil
}

func (v *View) Title() {
	v.mode = Title
	v.CleanPage()
	v.NewLine("Title: " + NovelTitle)
}

func (v *View) NewLine(s string) {
	v.duringLine = true
	v.x = 0
	isWait := true

	for _, r := range s {
		v.screen.SetContent(v.x, v.y, rune(r), nil, tcell.StyleDefault)
		v.screen.Show()

		if isWait {
			t := time.NewTicker(time.Duration(v.textSpeed) * time.Millisecond)
			select {
			case <-t.C:
			case <-v.userInput:
				isWait = false
			}
			t.Stop()
		}

		v.x += runewidth.RuneWidth(r)
		if v.x >= v.width-1 {
			v.x = 0
			v.y++
		}
	}

	v.y++
	v.duringLine = false
}

func (v *View) CleanPage() {
	v.screen.Fill(' ', tcell.StyleDefault)
	v.y = 0
}

func (v *View) PageFeedSymbol() {
	for {
		if !v.duringLine {
			break
		}
	}
	flag := true
	for {
		t := time.NewTicker(500 * time.Millisecond)
		select {
		case <-t.C:
			if flag {
				v.screen.SetContent(v.x, v.y-1, rune(';'), nil, tcell.StyleDefault)
				flag = !flag
			} else {
				v.screen.SetContent(v.x, v.y-1, rune(' '), nil, tcell.StyleDefault)
				flag = !flag
			}
			v.screen.Show()
		case <-v.userInput:
			return
		}
		t.Stop()
	}
}
