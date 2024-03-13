package main
import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
//	"slices"
//	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

func GetFiles(path string) []string {
	entries, _ := ioutil.ReadDir(path)
	var ret []string

	for _, e := range entries {
		ret = append(ret, e.Name())
	}

	return ret
}

type Ranger struct {
	wd      string
	sel     string
	history History

	historyMoment string

	left   []string
	middle []string
	right  []string

	topPane    *Bar
	leftPane   *FilesPane
	middlePane *FilesPane
	rightPane  *FilesPane
	bottomPane *Bar
}

func (r *Ranger) Init() error {
	var err error
	r.wd, err = os.Getwd()

	r.topPane = NewBar(&r.wd)
	r.leftPane = NewFilesPane(&r.left)
	r.middlePane = NewFilesPane(&r.middle)
	r.rightPane = NewFilesPane(&r.right)
	r.bottomPane = NewBar(&r.historyMoment)

	// KIVA KIVA HII
//	r.middlePane.SetSelectedEntryFromIndex(0)
	wdFiles := GetFiles(r.wd)
	if len(wdFiles) > 0 {
		r.sel = filepath.Join(r.wd, wdFiles[0])
	}

	r.history.AddToHistory(r.sel)
	r.UpdatePanes()

	return err
}

func (r *Ranger) GetRightPath() (string, error) {
	if r.middlePane.selectedEntry >= len(r.middle) {
		return "", errors.New("Out of bounds")
	}

	return r.GetSelectedFilePath(), nil
	// return filepath.Join(r.wd, r.middle[r.middlePane.selectedEntry]), nil
}

func (r *Ranger) UpdatePanes() {
	r.left = GetFiles(filepath.Dir(r.wd))
	r.middle = GetFiles(r.wd)
	r.right = GetFiles(r.sel)

	if r.wd != "/" {
		r.leftPane.SetSelectedEntryFromString(filepath.Base(r.wd))
	} else {
		r.left = []string{}
	}
	r.middlePane.SetSelectedEntryFromString(filepath.Base(r.sel))
	h, err := r.history.GetHistoryEntryForPath(r.sel)
	if err != nil {
		r.rightPane.SetSelectedEntryFromIndex(0)
		return
	}
	r.rightPane.SetSelectedEntryFromString(filepath.Base(h))
}

func (r *Ranger) GetSelectedFilePath() string {
	if r.middlePane.selectedEntry >= len(r.middle) {
		return ""
	}
	return filepath.Join(r.wd, r.middle[r.middlePane.selectedEntry])
}

func (r *Ranger) GoLeft() {
	if filepath.Dir(r.wd) == r.wd {
		return
	}

	r.sel = r.wd
	r.wd = filepath.Dir(r.wd)
}

func (r *Ranger) GoRight() {
	if len(GetFiles(r.sel)) <= 0 {
		return
	}

	r.wd = r.sel
	var err error
	r.sel, err = r.history.GetHistoryEntryForPath(r.wd)
	if err != nil {
		// FIXME
		r.sel = filepath.Join(r.wd, r.rightPane.GetSelectedEntryFromIndex(0))
//		r.sel = r.rightPane.GetSelectedEntryFromIndex(0)
	}
}

func (r *Ranger) GoUp() {
	if r.middlePane.selectedEntry - 1 < 0 {
		r.sel = filepath.Join(r.wd, r.middlePane.GetSelectedEntryFromIndex(0))
		return
	}

	r.sel = filepath.Join(r.wd, r.middlePane.GetSelectedEntryFromIndex(r.middlePane.selectedEntry - 1))
}

func (r *Ranger) GoDown() {
	if r.middlePane.selectedEntry + 1 >= len(r.middle) {
		r.sel = filepath.Join(r.wd, r.middlePane.GetSelectedEntryFromIndex(len(r.middle) - 1))
		return
	}

	r.sel = filepath.Join(r.wd, r.middlePane.GetSelectedEntryFromIndex(r.middlePane.selectedEntry + 1))
}

func main() {
	var ranger Ranger
	ranger.Init()

	app := tview.NewApplication()

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Rune() == 'q' {
			app.Stop()
			return nil
		}

		if event.Key() == tcell.KeyF1 {
			cmd := exec.Command("nano", ranger.GetSelectedFilePath())
			cmd.Stdin = os.Stdin
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
			if err := cmd.Run(); err != nil {
				log.Fatal(err)
			}
			ranger.UpdatePanes()
			return nil
		}

		wasMovementKey := true
		if event.Key() == tcell.KeyLeft || event.Rune() == 'h' {
			ranger.GoLeft()
		} else if event.Key() == tcell.KeyRight || event.Rune() == 'l' {
			ranger.GoRight()
		} else if event.Key() == tcell.KeyUp || event.Rune() == 'k' {
			ranger.GoUp()
		} else if event.Key() == tcell.KeyDown || event.Rune() == 'j' {
			ranger.GoDown()
		} else {
			wasMovementKey = false
		}

		if wasMovementKey {
			if !(event.Key() == tcell.KeyLeft || event.Rune() == 'h') {
				ranger.history.AddToHistory(ranger.sel)
			}

/*			ranger.historyMoment = ""
			for _, e := range ranger.history.history {
				ranger.historyMoment += filepath.Base(e) + ", "
			}*/
			ranger.UpdatePanes()
			return nil
		}

		if event.Rune() == ' ' {
			ranger.historyMoment = ranger.GetSelectedFilePath()
		}

		return event
	})

	flex := tview.NewFlex().SetDirection(tview.FlexRow).
		AddItem(ranger.topPane, 1, 0, false).
		AddItem(tview.NewFlex().SetDirection(tview.FlexColumn).
			AddItem(ranger.leftPane, 0, 1, false).
			AddItem(ranger.middlePane, 0, 2, false).
			AddItem(ranger.rightPane, 0, 2, false), 0, 1, false).
		AddItem(ranger.bottomPane, 1, 0, false)

	if err := app.SetRoot(flex, true).Run(); err != nil {
		panic(err)
	}
}
