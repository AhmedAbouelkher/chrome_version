package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var (
	ErrInvalidVersion = errors.New("invalid version")
)

type Version struct {
	First int
	Sec   int
	Third int
	Last  int
}

func (v *Version) IsAfter(other *Version) bool {
	if v.First > other.First {
		return true
	}
	if v.First < other.First {
		return false
	}
	if v.Sec > other.Sec {
		return true
	}
	if v.Sec < other.Sec {
		return false
	}
	if v.Third > other.Third {
		return true
	}
	if v.Third < other.Third {
		return false
	}
	if v.Last > other.Last {
		return true
	}
	return false
}

func fromString(v string) (*Version, error) {
	s := strings.Split(v, ".")
	if len(s) != 4 {
		return nil, ErrInvalidVersion
	}

	return &Version{
		First: toInt(s[0]),
		Sec:   toInt(s[1]),
		Third: toInt(s[2]),
		Last:  toInt(s[3]),
	}, nil
}

func toInt(s string) int {
	i, err := strconv.Atoi(s)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func (v *Version) String() string {
	return fmt.Sprintf("%d.%d.%d.%d", v.First, v.Sec, v.Third, v.Last)
}

func main() {
	f, err := os.OpenFile("version.txt", os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	dbv, err := getDBVersion(f)
	if err != nil && !errors.Is(err, ErrInvalidVersion) {
		log.Fatal(err)
	}

	v, err := getVersion()
	if err != nil {
		log.Fatal(err)
	}

	if dbv == nil || v.IsAfter(dbv) {
		// write the version directly to version.txt
		writeToFile(f, v)
		// TODO: Send slack message
		return
	}

	log.Println("Finished")
}

func getDBVersion(f *os.File) (*Version, error) {
	d, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	o := strings.Trim(string(d), "\n")
	return fromString(o)
}

func getVersion() (*Version, error) {
	res, err := http.Get("https://omahaproxy.appspot.com/win")
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code: %d", res.StatusCode)
	}

	d, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	o := string(d)
	return fromString(o)
}

func writeToFile(f *os.File, v *Version) error {
	err := f.Truncate(0)
	if err != nil {
		return err
	}
	if _, err = f.Seek(0, 0); err != nil {
		return err
	}
	if _, err := f.WriteString(v.String()); err != nil {
		return err
	}
	return nil
}
