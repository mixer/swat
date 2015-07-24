package profile

import (
	"io"
	"os"
)

// Targeter is embedded and used to set the output for actions.
type targeter struct {
	writer io.Writer
	closer io.Closer
}

// Writes the output of the action to the writer.
func (t *targeter) ToWriter(w io.Writer) {
	t.writer = w
}

// Writes the output of the action to the file specified by the path.
func (t *targeter) ToFile(file string) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}

	t.writer = f
	t.closer = f
	return nil
}

func (t *targeter) end() {
	if t.closer != nil {
		t.closer.Close()
	}
}
