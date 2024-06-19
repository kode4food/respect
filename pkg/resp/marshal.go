package resp

import (
	"io"
	"strconv"
)

var NewLine = []byte{CR, LF}

func writeSimple(tag Tag, data []byte, w io.Writer) error {
	if _, err := w.Write([]byte{byte(tag)}); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return writeNewline(w)
}

func writeBulk(tag Tag, data []byte, w io.Writer) error {
	if _, err := w.Write([]byte{byte(tag)}); err != nil {
		return err
	}
	if err := writeLen(data, w); err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return err
	}
	return writeNewline(w)
}

func writeValues(tag Tag, arr Values, w io.Writer) error {
	if _, err := w.Write([]byte{byte(tag)}); err != nil {
		return err
	}
	if err := writeLen(arr, w); err != nil {
		return err
	}
	for _, v := range arr {
		if err := v.Marshal(w); err != nil {
			return err
		}
	}
	return nil
}

func writeNewline(w io.Writer) error {
	_, err := w.Write(NewLine)
	return err
}

func writeLen[T any](a []T, w io.Writer) error {
	return writeInt(len(a), w)
}

func writeInt(i int, w io.Writer) error {
	s := strconv.Itoa(i)
	if _, err := w.Write([]byte(s)); err != nil {
		return err
	}
	return writeNewline(w)
}
