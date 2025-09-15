package goasteroids

import (
	"fmt"
	"os"
	"os/user"
	"runtime"
	"strconv"
	"strings"
)

func getHighScore() (int, error) {
	// Get the user name.
	u, err := user.Current()
	if err != nil {
		return 0, err
	}

	path := ""
	switch runtime.GOOS {
	case "darwin":
		path = fmt.Sprintf("/Users/%s/Library/Application Support/Go Asteroids", u.Username)
	case "windows":
		path = fmt.Sprintf("C:\\Users\\%s\\AppData", u.Username)
	default:
		path = fmt.Sprintf("/users/%s", u.Username)
	}

	if _, err := os.Stat(path); err != nil {
		if err := os.Mkdir(path, 0750); err != nil {
			return 0, err
		}
	}

	if _, err := os.Stat(path + "/high-score.txt"); err != nil {
		err := os.WriteFile(path+"/high-score.txt", []byte("0"), 0750)
		if err != nil {
			return 0, err
		}
	}

	contents, err := os.ReadFile(path + "/high-score.txt")
	if err != nil {
		return 0, err
	}
	score := string(contents)
	score = strings.TrimSpace(score)
	s, err := strconv.Atoi(string(score))
	if err != nil {
		return 0, err
	}

	return s, nil
}

func updateHighScore(score int) error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	path := ""
	switch runtime.GOOS {
	case "darwin":
		path = fmt.Sprintf("/Users/%s/Library/Application Support/Go Asteroids/high-score.txt", u.Username)
	case "windows":
		path = fmt.Sprintf("C:\\Users\\%s\\AppData\\high-score.txt", u.Username)
	default:
		path = fmt.Sprintf("/users/%s/high-score.txt", u.Username)
	}

	s := fmt.Sprintf("%d", score)
	if err := os.WriteFile(path, []byte(s), 0750); err != nil {
		return err
	}

	return nil
}
